//
// Copyright (c) 2021 One Track Consulting
// Copyright (c) 2022 Intel Corporation
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

package app

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/trigger"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"

	"github.com/hashicorp/go-multierror"
	gometrics "github.com/rcrowley/go-metrics"
)

type simpleTriggerServiceBinding struct {
	*Service
	*runtime.FunctionsPipelineRuntime
}

func (b *simpleTriggerServiceBinding) SecretProvider() messaging.SecretDataProvider {
	return container.SecretProviderFrom(b.Service.dic.Get)
}

func NewTriggerServiceBinding(svc *Service) trigger.ServiceBinding {
	return &simpleTriggerServiceBinding{
		svc,
		svc.runtime,
	}
}

func (b *simpleTriggerServiceBinding) DIC() *di.Container {
	return b.Service.dic
}

func (b *simpleTriggerServiceBinding) BuildContext(env types.MessageEnvelope) interfaces.AppFunctionContext {
	return appfunction.NewContext(env.CorrelationID, b.Service.dic, env.ContentType)
}

func (b *simpleTriggerServiceBinding) Config() *common.ConfigurationStruct {
	return b.Service.config
}

// triggerMessageProcessor wraps the ServiceBinding interface so that we can attach methods
type triggerMessageProcessor struct {
	serviceBinding trigger.ServiceBinding
	// messagesReceived includes all messages received (valid and invalid)
	messagesReceived        gometrics.Counter
	invalidMessagesReceived gometrics.Counter
}

func NewTriggerMessageProcessor(bnd trigger.ServiceBinding, metricsManager bootstrapInterfaces.MetricsManager) *triggerMessageProcessor {
	lc := bnd.LoggingClient()

	mp := &triggerMessageProcessor{
		serviceBinding:          bnd,
		messagesReceived:        gometrics.NewCounter(),
		invalidMessagesReceived: gometrics.NewCounter(),
	}

	if err := metricsManager.Register(internal.MessagesReceivedName, mp.messagesReceived, nil); err != nil {
		lc.Warnf("%s metric failed to register and will not be reported: %s", internal.MessagesReceivedName, err.Error())
	} else {
		lc.Infof("%s metric has been registered and will be reported", internal.MessagesReceivedName)
	}

	if err := metricsManager.Register(internal.InvalidMessagesReceivedName, mp.invalidMessagesReceived, nil); err != nil {
		lc.Warnf("%s metric failed to register and will not be reported: %s", internal.InvalidMessagesReceivedName, err.Error())
	} else {
		lc.Infof("%s metric has been registered and will be reported (if enabled)", internal.InvalidMessagesReceivedName)
	}

	return mp
}

// MessageReceived provides runtime orchestration to pass the envelope / context to configured pipeline(s) along with a response callback to execute on each completion.
func (mp *triggerMessageProcessor) MessageReceived(ctx interfaces.AppFunctionContext, envelope types.MessageEnvelope, responseHandler interfaces.PipelineResponseHandler) error {
	mp.messagesReceived.Inc(1)
	lc := mp.serviceBinding.LoggingClient()
	lc.Debugf("trigger attempting to find pipeline(s) for topic %s", envelope.ReceivedTopic)

	// ensure we have a context established that we can safely cast to *appfunction.Context to pass to runtime
	if _, ok := ctx.(*appfunction.Context); ctx == nil || !ok {
		ctx = mp.serviceBinding.BuildContext(envelope)
	}

	pipelines := mp.serviceBinding.GetMatchingPipelines(envelope.ReceivedTopic)

	lc.Debugf("trigger found %d pipeline(s) that match the incoming topic '%s'", len(pipelines), envelope.ReceivedTopic)

	var finalErr error
	errorCollectionLock := sync.RWMutex{}

	pipelinesWaitGroup := sync.WaitGroup{}

	appContext, ok := ctx.(*appfunction.Context)
	if !ok {
		return fmt.Errorf("context received was not *appfunction.Context (%T)", ctx)
	}

	targetData, err, isInvalidMessage := mp.serviceBinding.DecodeMessage(appContext, envelope)
	if err != nil {
		if isInvalidMessage {
			mp.invalidMessagesReceived.Inc(1)
		}
		return fmt.Errorf("unable to decode message: %s", err.Err.Error())
	}

	for _, pipeline := range pipelines {
		pipelinesWaitGroup.Add(1)
		pipeline.MessagesProcessed.Inc(1)

		go func(p *interfaces.FunctionPipeline, wg *sync.WaitGroup, errCollector func(error)) {
			startedAt := time.Now()
			defer p.MessageProcessingTime.UpdateSince(startedAt)
			defer wg.Done()

			lc.Debugf("trigger sending message to pipeline %s (%s)", p.Id, envelope.CorrelationID)

			childCtx, ok := ctx.Clone().(*appfunction.Context)

			if !ok {
				errCollector(fmt.Errorf("context received was not *appfunction.Context (%T)", childCtx))
				return
			}

			if msgErr := mp.serviceBinding.ProcessMessage(childCtx, targetData, p); msgErr != nil {
				lc.Errorf("message error in pipeline %s (%s): %s", p.Id, envelope.CorrelationID, msgErr.Err.Error())
				errCollector(msgErr.Err)
			} else {
				if responseHandler != nil {
					if outputErr := responseHandler(childCtx, p); outputErr != nil {
						lc.Errorf("failed to process output for message '%s' on pipeline %s: %s", ctx.CorrelationID(), p.Id, outputErr.Error())
						errCollector(outputErr)
						return
					}
				}
				lc.Debugf("trigger successfully processed message '%s' in pipeline %s", p.Id, envelope.CorrelationID)
			}
		}(pipeline, &pipelinesWaitGroup, func(e error) {
			errorCollectionLock.Lock()
			defer errorCollectionLock.Unlock()
			finalErr = multierror.Append(finalErr, e)
		})
	}

	pipelinesWaitGroup.Wait()

	return finalErr
}

// ReceivedInvalidMessage is called when an error occurs decoding the message from the MessageBus into the MessageEnvelope
func (mp *triggerMessageProcessor) ReceivedInvalidMessage() {
	// messagesReceived includes all message received
	mp.messagesReceived.Inc(1)
	// invalidMessagesReceived is just the invalid messages received
	mp.invalidMessagesReceived.Inc(1)
}
