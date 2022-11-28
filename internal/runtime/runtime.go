//
// Copyright (c) 2022 Intel Corporation
// Copyright (c) 2021 One Track Consulting
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/fxamacker/cbor/v2"
	gometrics "github.com/rcrowley/go-metrics"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

const (
	TopicWildCard       = "#"
	TopicLevelSeparator = "/"
)

func NewFunctionPipeline(id string, topics []string, transforms []interfaces.AppFunction) interfaces.FunctionPipeline {
	pipeline := interfaces.FunctionPipeline{
		Id:                    id,
		Transforms:            transforms,
		Topics:                topics,
		Hash:                  calculatePipelineHash(transforms),
		MessagesProcessed:     gometrics.NewCounter(),
		MessageProcessingTime: gometrics.NewTimer(),
		ProcessingErrors:      gometrics.NewCounter(),
	}

	return pipeline
}

// AppServiceRuntime represents the golang runtime environment
type AppServiceRuntime struct {
	TargetType    interface{}
	ServiceKey    string
	pipelines     map[string]*interfaces.FunctionPipeline
	isBusyCopying sync.Mutex
	storeForward  storeForwardInfo
	lc            logger.LoggingClient
	dic           *di.Container
}

type MessageError struct {
	Err       error
	ErrorCode int
}

// NewAppServiceRuntime creates and initializes the AppServiceRuntime instance
func NewAppServiceRuntime(serviceKey string, targetType interface{}, dic *di.Container) *AppServiceRuntime {
	asr := &AppServiceRuntime{
		ServiceKey: serviceKey,
		TargetType: targetType,
		dic:        dic,
		pipelines:  make(map[string]*interfaces.FunctionPipeline),
	}

	asr.storeForward.dic = dic
	asr.storeForward.runtime = asr
	asr.lc = bootstrapContainer.LoggingClientFrom(asr.dic.Get)

	return asr
}

// SetDefaultFunctionsPipeline sets the default function pipeline
func (asr *AppServiceRuntime) SetDefaultFunctionsPipeline(transforms []interfaces.AppFunction) {
	pipeline := asr.GetDefaultPipeline() // ensures the default pipeline exists
	asr.SetFunctionsPipelineTransforms(pipeline.Id, transforms)
}

// SetFunctionsPipelineTransforms sets the transforms for an existing function pipeline.
// Non-existent pipelines are ignored
func (asr *AppServiceRuntime) SetFunctionsPipelineTransforms(id string, transforms []interfaces.AppFunction) {
	pipeline := asr.pipelines[id]
	if pipeline != nil {
		asr.isBusyCopying.Lock()
		pipeline.Transforms = transforms
		pipeline.Hash = calculatePipelineHash(transforms)
		asr.isBusyCopying.Unlock()
		asr.lc.Infof("Transforms set for `%s` pipeline", id)
	} else {
		asr.lc.Warnf("Unable to set transforms for `%s` pipeline: Pipeline not found", id)
	}
}

// SetFunctionsPipelineTopics sets the topics for an existing function pipeline.
// Non-existent pipelines are ignored
func (asr *AppServiceRuntime) SetFunctionsPipelineTopics(id string, topics []string) {
	pipeline := asr.pipelines[id]
	if pipeline != nil {
		asr.isBusyCopying.Lock()
		pipeline.Topics = topics
		asr.isBusyCopying.Unlock()
		asr.lc.Infof("Topics set for `%s` pipeline", id)
	} else {
		asr.lc.Warnf("Unable to set topica for `%s` pipeline: Pipeline not found", id)
	}
}

// ClearAllFunctionsPipelineTransforms clears the transforms for all existing function pipelines.
func (asr *AppServiceRuntime) ClearAllFunctionsPipelineTransforms() {
	asr.isBusyCopying.Lock()
	for index := range asr.pipelines {
		asr.pipelines[index].Transforms = nil
		asr.pipelines[index].Hash = ""
	}
	asr.isBusyCopying.Unlock()
}

// RemoveAllFunctionPipelines removes all existing function pipelines
func (gr *GolangRuntime) RemoveAllFunctionPipelines() {
	metricManager := bootstrapContainer.MetricsManagerFrom(gr.dic.Get)

	gr.isBusyCopying.Lock()
	for id := range gr.pipelines {
		gr.unregisterPipelineMetric(metricManager, internal.PipelineMessagesProcessedName, id)
		gr.unregisterPipelineMetric(metricManager, internal.PipelineMessageProcessingTimeName, id)
		gr.unregisterPipelineMetric(metricManager, internal.PipelineProcessingErrorsName, id)
		delete(gr.pipelines, id)
	}
	gr.isBusyCopying.Unlock()
}

// AddFunctionsPipeline is thread safe to set transforms
func (asr *AppServiceRuntime) AddFunctionsPipeline(id string, topics []string, transforms []interfaces.AppFunction) error {
	_, exists := asr.pipelines[id]
	if exists {
		return fmt.Errorf("pipeline with Id='%s' already exists", id)
	}

	_ = asr.addFunctionsPipeline(id, topics, transforms)
	return nil
}

func (asr *AppServiceRuntime) addFunctionsPipeline(id string, topics []string, transforms []interfaces.AppFunction) *interfaces.FunctionPipeline {
	pipeline := NewFunctionPipeline(id, topics, transforms)
	asr.isBusyCopying.Lock()
	asr.pipelines[id] = &pipeline
	asr.isBusyCopying.Unlock()

	metricManager := bootstrapContainer.MetricsManagerFrom(asr.dic.Get)
	asr.registerPipelineMetric(metricManager, internal.PipelineMessagesProcessedName, pipeline.Id, pipeline.MessagesProcessed)
	asr.registerPipelineMetric(metricManager, internal.PipelineMessageProcessingTimeName, pipeline.Id, pipeline.MessageProcessingTime)
	asr.registerPipelineMetric(metricManager, internal.PipelineProcessingErrorsName, pipeline.Id, pipeline.ProcessingErrors)

	return &pipeline
}

func (asr *AppServiceRuntime) registerPipelineMetric(metricManager bootstrapInterfaces.MetricsManager, metricName string, pipelineId string, metric interface{}) {
	registeredName := strings.Replace(metricName, internal.PipelineIdTxt, pipelineId, 1)
	err := metricManager.Register(registeredName, metric, map[string]string{"pipeline": pipelineId})
	if err != nil {
		asr.lc.Warnf("Unable to register %s metric. Metric will not be reported : %s", registeredName, err.Error())
	} else {
		asr.lc.Infof("%s metric has been registered and will be reported (if enabled)", registeredName)
	}
}

func (gr *GolangRuntime) unregisterPipelineMetric(metricManager bootstrapInterfaces.MetricsManager, metricName string, pipelineId string) {
	registeredName := strings.Replace(metricName, internal.PipelineIdTxt, pipelineId, 1)
	metricManager.Unregister(registeredName)
}

// ProcessMessage sends the contents of the message through the functions pipeline
func (asr *AppServiceRuntime) ProcessMessage(appContext *appfunction.Context, target interface{}, pipeline *interfaces.FunctionPipeline) *MessageError {
	if len(pipeline.Transforms) == 0 {
		err := fmt.Errorf("no transforms configured for pipleline Id='%s'. Please check log for earlier errors loading pipeline", pipeline.Id)
		asr.logError(err, appContext.CorrelationID())
		return &MessageError{Err: err, ErrorCode: http.StatusInternalServerError}
	}

	appContext.AddValue(interfaces.PIPELINEID, pipeline.Id)

	asr.lc.Debugf("Pipeline '%s' processing message %d Transforms", pipeline.Id, len(pipeline.Transforms))

	// Make copy of transform functions to avoid disruption of pipeline when updating the pipeline from registry
	asr.isBusyCopying.Lock()
	execPipeline := &interfaces.FunctionPipeline{
		Id:                    pipeline.Id,
		Transforms:            make([]interfaces.AppFunction, len(pipeline.Transforms)),
		Topics:                pipeline.Topics,
		Hash:                  pipeline.Hash,
		MessagesProcessed:     pipeline.MessagesProcessed,
		MessageProcessingTime: pipeline.MessageProcessingTime,
		ProcessingErrors:      pipeline.ProcessingErrors,
	}
	copy(execPipeline.Transforms, pipeline.Transforms)
	asr.isBusyCopying.Unlock()

	return asr.ExecutePipeline(target, appContext, execPipeline, 0, false)
}

// DecodeMessage decode the message wrapped in the MessageEnvelope and return the data to be processed.
func (asr *AppServiceRuntime) DecodeMessage(appContext *appfunction.Context, envelope types.MessageEnvelope) (interface{}, *MessageError, bool) {
	// Default Target Type for the function pipeline is an Event DTO.
	// The Event DTO can be wrapped in an AddEventRequest DTO or just be the un-wrapped Event DTO,
	// which is handled dynamically below.
	if asr.TargetType == nil {
		asr.TargetType = &dtos.Event{}
	}

	if reflect.TypeOf(asr.TargetType).Kind() != reflect.Ptr {
		err := errors.New("TargetType must be a pointer, not a value of the target type")
		asr.logError(err, envelope.CorrelationID)
		return nil, &MessageError{Err: err, ErrorCode: http.StatusInternalServerError}, false
	}

	// Must make a copy of the type so that data isn't retained between calls for custom types
	target := reflect.New(reflect.ValueOf(asr.TargetType).Elem().Type()).Interface()

	switch target.(type) {
	case *[]byte:
		asr.lc.Debug("Expecting raw byte data")
		target = &envelope.Payload

	case *dtos.Event:
		asr.lc.Debug("Expecting an AddEventRequest or Event DTO")

		// Dynamically process either AddEventRequest or Event DTO
		event, err := asr.processEventPayload(envelope)
		if err != nil {
			err = fmt.Errorf("unable to process payload %s", err.Error())
			asr.logError(err, envelope.CorrelationID)
			return nil, &MessageError{Err: err, ErrorCode: http.StatusBadRequest}, true
		}

		if asr.lc.LogLevel() == models.DebugLog {
			asr.debugLogEvent(event)
		}

		appContext.AddValue(interfaces.DEVICENAME, event.DeviceName)
		appContext.AddValue(interfaces.PROFILENAME, event.ProfileName)
		appContext.AddValue(interfaces.SOURCENAME, event.SourceName)

		target = event

	default:
		customTypeName := di.TypeInstanceToName(target)
		asr.lc.Debugf("Expecting a custom type of %s", customTypeName)

		// Expecting a custom type so just unmarshal into the target type.
		if err := asr.unmarshalPayload(envelope, target); err != nil {
			err = fmt.Errorf("unable to process custom object received of type '%s': %s", customTypeName, err.Error())
			asr.logError(err, envelope.CorrelationID)
			return nil, &MessageError{Err: err, ErrorCode: http.StatusBadRequest}, true
		}
	}

	appContext.SetCorrelationID(envelope.CorrelationID)
	appContext.SetInputContentType(envelope.ContentType)
	appContext.AddValue(interfaces.RECEIVEDTOPIC, envelope.ReceivedTopic)

	// All functions expect an object, not a pointer to an object, so must use reflection to
	// dereference to pointer to the object
	target = reflect.ValueOf(target).Elem().Interface()

	return target, nil, false
}

func (asr *AppServiceRuntime) ExecutePipeline(
	target interface{},
	appContext *appfunction.Context,
	pipeline *interfaces.FunctionPipeline,
	startPosition int,
	isRetry bool) *MessageError {

	var result interface{}
	var continuePipeline bool

	for functionIndex, trxFunc := range pipeline.Transforms {
		if functionIndex < startPosition {
			continue
		}

		appContext.SetRetryData(nil)

		if result == nil {
			continuePipeline, result = trxFunc(appContext, target)
		} else {
			continuePipeline, result = trxFunc(appContext, result)
		}

		if !continuePipeline {
			if result != nil {
				if err, ok := result.(error); ok {
					appContext.LoggingClient().Errorf(
						"Pipeline (%s) function #%d resulted in error: %s (%s=%s)",
						pipeline.Id,
						functionIndex,
						err.Error(),
						common.CorrelationHeader,
						appContext.CorrelationID())
					if appContext.RetryData() != nil && !isRetry {
						asr.storeForward.storeForLaterRetry(appContext.RetryData(), appContext, pipeline, functionIndex)
					}

					pipeline.ProcessingErrors.Inc(1)
					return &MessageError{Err: err, ErrorCode: http.StatusUnprocessableEntity}
				}
			}
			break
		}
	}

	return nil
}

func (asr *AppServiceRuntime) StartStoreAndForward(
	appWg *sync.WaitGroup,
	appCtx context.Context,
	enabledWg *sync.WaitGroup,
	enabledCtx context.Context,
	serviceKey string) {

	asr.storeForward.startStoreAndForwardRetryLoop(appWg, appCtx, enabledWg, enabledCtx, serviceKey)
}

func (asr *AppServiceRuntime) processEventPayload(envelope types.MessageEnvelope) (*dtos.Event, error) {

	asr.lc.Debug("Attempting to process Payload as an AddEventRequest DTO")
	requestDto := requests.AddEventRequest{}

	// Note that DTO validation is called during the unmarshaling
	// which results in a KindContractInvalid error
	requestDtoErr := asr.unmarshalPayload(envelope, &requestDto)
	if requestDtoErr == nil {
		asr.lc.Debug("Using Event DTO from AddEventRequest DTO")

		// Determine that we have an AddEventRequest DTO
		return &requestDto.Event, nil
	}

	// Check for validation error
	if edgexErrors.Kind(requestDtoErr) != edgexErrors.KindContractInvalid {
		return nil, requestDtoErr
	}

	// KindContractInvalid indicates that we likely don't have an AddEventRequest
	// so try to process as Event
	asr.lc.Debug("Attempting to process Payload as an Event DTO")
	event := &dtos.Event{}
	err := asr.unmarshalPayload(envelope, event)
	if err == nil {
		err = common.Validate(event)
		if err == nil {
			asr.lc.Debug("Using Event DTO received")
			return event, nil
		}
	}

	// Check for validation error
	if edgexErrors.Kind(err) != edgexErrors.KindContractInvalid {
		return nil, err
	}

	// Still unable to process so assume have invalid AddEventRequest DTO
	return nil, requestDtoErr
}

func (asr *AppServiceRuntime) unmarshalPayload(envelope types.MessageEnvelope, target interface{}) error {
	var err error

	contentType := strings.Split(envelope.ContentType, ";")[0]

	switch contentType {
	case common.ContentTypeJSON:
		err = json.Unmarshal(envelope.Payload, target)

	case common.ContentTypeCBOR:
		err = cbor.Unmarshal(envelope.Payload, target)

	default:
		err = fmt.Errorf("unsupported content-type '%s' recieved", envelope.ContentType)
	}

	return err
}

func (asr *AppServiceRuntime) debugLogEvent(event *dtos.Event) {
	asr.lc.Debugf("Event Received with ProfileName=%s, DeviceName=%s and ReadingCount=%d",
		event.ProfileName,
		event.DeviceName,
		len(event.Readings))
	if len(event.Tags) > 0 {
		asr.lc.Debugf("Event tags are: [%v]", event.Tags)
	} else {
		asr.lc.Debug("Event has no tags")
	}

	for index, reading := range event.Readings {
		switch strings.ToLower(reading.ValueType) {
		case strings.ToLower(common.ValueTypeBinary):
			asr.lc.Debugf("Reading #%d received with ResourceName=%s, ValueType=%s, MediaType=%s and BinaryValue of size=`%d`",
				index+1,
				reading.ResourceName,
				reading.ValueType,
				reading.MediaType,
				len(reading.BinaryValue))
		default:
			asr.lc.Debugf("Reading #%d received with ResourceName=%s, ValueType=%s, Value=`%s`",
				index+1,
				reading.ResourceName,
				reading.ValueType,
				reading.Value)
		}
	}
}

func (asr *AppServiceRuntime) logError(err error, correlationID string) {
	asr.lc.Errorf("%s. %s=%s", err.Error(), common.CorrelationHeader, correlationID)
}

func (asr *AppServiceRuntime) GetDefaultPipeline() *interfaces.FunctionPipeline {
	pipeline := asr.pipelines[interfaces.DefaultPipelineId]
	if pipeline == nil {
		pipeline = asr.addFunctionsPipeline(interfaces.DefaultPipelineId, []string{TopicWildCard}, nil)
	}
	return pipeline
}

func (asr *AppServiceRuntime) GetMatchingPipelines(incomingTopic string) []*interfaces.FunctionPipeline {
	var matches []*interfaces.FunctionPipeline

	if len(asr.pipelines) == 0 {
		return matches
	}

	for _, pipeline := range asr.pipelines {
		if topicMatches(incomingTopic, pipeline.Topics) {
			matches = append(matches, pipeline)
		}
	}

	return matches
}

func (asr *AppServiceRuntime) GetPipelineById(id string) *interfaces.FunctionPipeline {
	return asr.pipelines[id]
}

func topicMatches(incomingTopic string, pipelineTopics []string) bool {
	for _, pipelineTopic := range pipelineTopics {
		if pipelineTopic == TopicWildCard {
			return true
		}

		wildcardCount := strings.Count(pipelineTopic, TopicWildCard)
		switch wildcardCount {
		case 0:
			if incomingTopic == pipelineTopic {
				return true
			}
		default:
			pipelineLevels := strings.Split(pipelineTopic, TopicLevelSeparator)
			incomingLevels := strings.Split(incomingTopic, TopicLevelSeparator)

			if len(pipelineLevels) > len(incomingLevels) {
				continue
			}

			for index, level := range pipelineLevels {
				if level == TopicWildCard {
					incomingLevels[index] = TopicWildCard
				}
			}

			incomingWithWildCards := strings.Join(incomingLevels, "/")
			if strings.Index(incomingWithWildCards, pipelineTopic) == 0 {
				return true
			}
		}
	}
	return false
}

func calculatePipelineHash(transforms []interfaces.AppFunction) string {
	hash := "Pipeline-functions: "
	for _, item := range transforms {
		name := runtime.FuncForPC(reflect.ValueOf(item).Pointer()).Name()
		hash = hash + " " + name
	}

	return hash
}
