//
// Copyright (c) 2020 Intel Corporation
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
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

type collectedMsg struct {
	topic string
	env   types.MessageEnvelope
}

// Trigger implements Trigger to support MessageBusData
type Trigger struct {
	Configuration *common.ConfigurationStruct
	Runtime       *runtime.GolangRuntime
	client        messaging.MessageClient
	topics        []types.TopicChannel
	collectedMsgs chan collectedMsg
	EdgeXClients  common.EdgeXClients
}

// Initialize ...
func (trigger *Trigger) Initialize(appWg *sync.WaitGroup, appCtx context.Context, background <-chan types.MessageEnvelope) (bootstrap.Deferred, error) {
	var err error
	logger := trigger.EdgeXClients.LoggingClient

	logger.Info(fmt.Sprintf("Initializing Message Bus Trigger for '%s'", trigger.Configuration.MessageBus.Type))

	trigger.client, err = messaging.NewMessageClient(trigger.Configuration.MessageBus)
	if err != nil {
		return nil, err
	}

	trigger.collectedMsgs = make(chan collectedMsg)

	for _, tpc := range strings.Split(trigger.Configuration.Binding.SubscribeTopics, ",") {
		trigger.topics = append(trigger.topics, types.TopicChannel{Topic: tpc, Messages: make(chan types.MessageEnvelope)})
	}

	messageErrors := make(chan error)

	err = trigger.client.Connect()
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Subscribing to topics: '%s' @ %s://%s:%d",
		trigger.Configuration.Binding.SubscribeTopics,
		trigger.Configuration.MessageBus.SubscribeHost.Protocol,
		trigger.Configuration.MessageBus.SubscribeHost.Host,
		trigger.Configuration.MessageBus.SubscribeHost.Port))

	for _, tc := range trigger.topics {
		go func(t types.TopicChannel) {
			for {
				select {
				case <-appCtx.Done():
					return

				case msg := <-t.Messages:
					trigger.collectedMsgs <- collectedMsg{
						topic: t.Topic,
						env:   msg,
					}
				}
			}
		}(tc)
	}

	err = trigger.client.Subscribe(trigger.topics, messageErrors)

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to subscribe to topic(s): [%s] (%s)", trigger.Configuration.Binding.SubscribeTopics, err.Error()))
		return nil, err
	}

	receiveMessage := true

	if len(trigger.Configuration.MessageBus.PublishHost.Host) > 0 {
		logger.Info(fmt.Sprintf("Publishing to topic: '%s' @ %s://%s:%d",
			trigger.Configuration.Binding.PublishTopic,
			trigger.Configuration.MessageBus.PublishHost.Protocol,
			trigger.Configuration.MessageBus.PublishHost.Host,
			trigger.Configuration.MessageBus.PublishHost.Port))
	}

	appWg.Add(1)

	go func() {
		defer appWg.Done()

		for receiveMessage {
			select {
			case <-appCtx.Done():
				return

			case msgErr := <-messageErrors:
				logger.Error(fmt.Sprintf("Failed to receive message from bus, %v", msgErr))

			case msgs := <-trigger.collectedMsgs:
				go func() {
					logger.Trace("Received message from bus", "topic", msgs.topic, clients.CorrelationHeader, msgs.env.CorrelationID)

					edgexContext := &appcontext.Context{
						CorrelationID:         msgs.env.CorrelationID,
						Configuration:         trigger.Configuration,
						LoggingClient:         trigger.EdgeXClients.LoggingClient,
						EventClient:           trigger.EdgeXClients.EventClient,
						ValueDescriptorClient: trigger.EdgeXClients.ValueDescriptorClient,
						CommandClient:         trigger.EdgeXClients.CommandClient,
						NotificationsClient:   trigger.EdgeXClients.NotificationsClient,
					}

					messageError := trigger.Runtime.ProcessMessage(edgexContext, msgs.env)
					if messageError != nil {
						// ProcessMessage logs the error, so no need to log it here.
						return
					}

					if edgexContext.OutputData != nil {
						outputEnvelope := types.MessageEnvelope{
							CorrelationID: edgexContext.CorrelationID,
							Payload:       edgexContext.OutputData,
							ContentType:   clients.ContentTypeJSON,
						}
						err := trigger.client.Publish(outputEnvelope, trigger.Configuration.Binding.PublishTopic)
						if err != nil {
							logger.Error(fmt.Sprintf("Failed to publish Message to bus, %v", err))
							return
						}

						logger.Trace("Published message to bus", "topic", trigger.Configuration.Binding.PublishTopic, clients.CorrelationHeader, msgs.env.CorrelationID)
					}
				}()
			case bg := <-background:
				go func() {
					err := trigger.client.Publish(bg, trigger.Configuration.Binding.PublishTopic)
					if err != nil {
						logger.Error(fmt.Sprintf("Failed to publish background Message to bus, %v", err))
						return
					}

					logger.Trace("Published background message to bus", "topic", trigger.Configuration.Binding.PublishTopic, clients.CorrelationHeader, bg.CorrelationID)
				}()
			}
		}
	}()

	deferred := func() {
		logger.Info("Disconnecting from the message bus")
		err := trigger.client.Disconnect()
		if err != nil {
			logger.Error("Unable to disconnect from the message bus", "error", err.Error())
		}
	}
	return deferred, nil
}
