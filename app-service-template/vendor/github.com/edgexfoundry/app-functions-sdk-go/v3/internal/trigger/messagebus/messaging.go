//
// Copyright (c) 2023 Intel Corporation
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

package messagebus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-messaging/v3/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/trigger"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/util"
)

// Trigger implements Trigger to support MessageBusData
type Trigger struct {
	messageProcessor trigger.MessageProcessor
	serviceBinding   trigger.ServiceBinding
	topics           []types.TopicChannel
	publishTopic     string
	client           messaging.MessageClient
	dic              *di.Container
}

func NewTrigger(bnd trigger.ServiceBinding, mp trigger.MessageProcessor, dic *di.Container) *Trigger {
	return &Trigger{
		messageProcessor: mp,
		serviceBinding:   bnd,
		dic:              dic,
	}
}

// Initialize ...
func (trigger *Trigger) Initialize(appWg *sync.WaitGroup, appCtx context.Context, background <-chan interfaces.BackgroundMessage) (bootstrap.Deferred, error) {
	lc := trigger.serviceBinding.LoggingClient()
	config := trigger.serviceBinding.Config()

	lc.Infof("Initializing EdgeX Message Bus Trigger for '%s'", config.MessageBus.Type)

	// MessageBus client is now created and connected by the MessageBus bootstrap handler and placed in the DIC.
	trigger.client = container.MessagingClientFrom(trigger.dic.Get)
	if trigger.client == nil {
		return nil, errors.New("unable to find MessageBus Client. Make sure it is configured properly")
	}

	subscribeTopics := strings.TrimSpace(config.Trigger.SubscribeTopics)
	if len(subscribeTopics) == 0 {
		errMsg := "'%s' can not be an empty string. Must contain one or more topic separated by commas, " +
			"missing common config? Use -cp or -cc flags for common config"
		return nil, fmt.Errorf(errMsg, internal.MessageBusSubscribeTopics)
	}

	topics := util.DeleteEmptyAndTrim(strings.FieldsFunc(subscribeTopics, util.SplitComma))
	subscribeTopics = ""
	for _, topic := range topics {
		topic = common.BuildTopic(config.MessageBus.GetBaseTopicPrefix(), topic)
		trigger.topics = append(trigger.topics, types.TopicChannel{Topic: topic, Messages: make(chan types.MessageEnvelope)})
		lc.Infof("Subscribing to topic: %s", topic)
	}

	messageErrors := make(chan error)

	trigger.publishTopic = strings.TrimSpace(config.Trigger.PublishTopic)
	if len(trigger.publishTopic) > 0 {
		trigger.publishTopic = common.BuildTopic(config.MessageBus.GetBaseTopicPrefix(), trigger.publishTopic)
		lc.Infof("Publishing to topic: %s", trigger.publishTopic)
	} else {
		lc.Infof("Publish topic not set for Trigger. Response data, if set, will not be published")
	}

	// Need to have a go func for each subscription, so we know with topic the data was received for.
	for _, topic := range trigger.topics {
		appWg.Add(1)
		go func(triggerTopic types.TopicChannel) {
			defer appWg.Done()
			lc.Infof("Waiting for messages from the MessageBus on the '%s' topic", triggerTopic.Topic)

			for {
				select {
				case <-appCtx.Done():
					lc.Infof("Exiting waiting for MessageBus '%s' topic messages", triggerTopic.Topic)
					return
				case message := <-triggerTopic.Messages:
					trigger.messageHandler(lc, triggerTopic, message)
				}
			}
		}(topic)
	}

	// Need an addition go func to handle errors and background publishing to the message bus.
	appWg.Add(1)
	go func() {
		defer appWg.Done()
		for {
			select {
			case <-appCtx.Done():
				lc.Info("Exiting waiting for MessageBus errors and background publishing")
				return

			case msgErr := <-messageErrors:
				lc.Errorf("error receiving message from bus, %s", msgErr.Error())
				// This will occur if an invalid message from the MessageBus fails to decode into the MessageEnvelope
				// Must let the message processor know, so it can update the service metrics it is managing.
				trigger.messageProcessor.ReceivedInvalidMessage()

			case bg := <-background:
				go func() {
					topic := bg.Topic()
					msg := bg.Message()

					err := trigger.client.Publish(msg, topic)
					if err != nil {
						lc.Errorf("Failed to publish background Message to bus, %v", err)
						return
					}

					lc.Debugf("Published background message to bus on %s topic", topic)
					lc.Tracef("%s=%s", common.CorrelationHeader, msg.CorrelationID)
				}()
			}
		}
	}()

	if err := trigger.client.Subscribe(trigger.topics, messageErrors); err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic(s) '%s': %s", subscribeTopics, err.Error())
	}

	deferred := func() {
		lc.Info("Disconnecting from the message bus")
		err := trigger.client.Disconnect()
		if err != nil {
			lc.Errorf("Unable to disconnect from the message bus: %s", err.Error())
		}
	}
	return deferred, nil
}

func (trigger *Trigger) messageHandler(logger logger.LoggingClient, _ types.TopicChannel, message types.MessageEnvelope) {
	logger.Debugf("MessageBus Trigger: Received message with %d bytes on topic '%s'. Content-Type=%s",
		len(message.Payload),
		message.ReceivedTopic,
		message.ContentType)
	logger.Tracef("MessageBus Trigger: Received message with %s=%s", common.CorrelationHeader, message.CorrelationID)

	appContext := trigger.serviceBinding.BuildContext(message)

	go func() {
		processErr := trigger.messageProcessor.MessageReceived(appContext, message, trigger.responseHandler)
		if processErr != nil {
			trigger.serviceBinding.LoggingClient().Errorf("MessageBus Trigger: Failed to process message on pipeline(s): %s", processErr.Error())
		}
	}()
}

func (trigger *Trigger) responseHandler(appContext interfaces.AppFunctionContext, pipeline *interfaces.FunctionPipeline) error {
	if appContext.ResponseData() != nil {
		lc := trigger.serviceBinding.LoggingClient()

		publishTopic, err := appContext.ApplyValues(trigger.publishTopic)

		if err != nil {
			lc.Errorf("MessageBus Trigger: Unable to format output topic '%s' for pipeline '%s': %s",
				trigger.publishTopic,
				pipeline.Id,
				err.Error())
			return err
		}

		var contentType string

		if appContext.ResponseContentType() != "" {
			contentType = appContext.ResponseContentType()
		} else {
			contentType = common.ContentTypeJSON
			if appContext.ResponseData()[0] != byte('{') && appContext.ResponseData()[0] != byte('[') {
				// If not JSON then assume it is CBOR
				contentType = common.ContentTypeCBOR
			}
		}
		outputEnvelope := types.MessageEnvelope{
			CorrelationID: appContext.CorrelationID(),
			Payload:       appContext.ResponseData(),
			ContentType:   contentType,
		}

		err = trigger.client.Publish(outputEnvelope, publishTopic)

		if err != nil {
			lc.Errorf("MessageBus trigger: Could not publish to topic '%s' for pipeline '%s': %s",
				publishTopic,
				pipeline.Id,
				err.Error())
			return err
		}

		lc.Debugf("MessageBus Trigger: Published response message for pipeline '%s' on topic '%s' with %d bytes",
			pipeline.Id,
			publishTopic,
			len(appContext.ResponseData()))
		lc.Tracef("MessageBus Trigger published message: %s=%s", common.CorrelationHeader, appContext.CorrelationID())
	}
	return nil
}
