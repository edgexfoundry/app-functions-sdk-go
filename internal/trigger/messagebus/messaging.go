//
// Copyright (c) 2021 Intel Corporation
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

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	bootstrapMessaging "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
)

// Trigger implements Trigger to support MessageBusData
type Trigger struct {
	messageProcessor trigger.MessageProcessor
	serviceBinding   trigger.ServiceBinding
	topics           []types.TopicChannel
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
	var err error

	lc := trigger.serviceBinding.LoggingClient()
	config := trigger.serviceBinding.Config()

	lc.Infof("Initializing Message Bus Trigger for '%s'", config.Trigger.EdgexMessageBus.Type)

	clientConfig := trigger.createMessagingClientConfig(config.Trigger.EdgexMessageBus)

	if err := trigger.setOptionalAuthData(&clientConfig, lc); err != nil {
		return nil, err
	}

	trigger.client, err = messaging.NewMessageClient(clientConfig)
	if err != nil {
		return nil, err
	}

	trigger.dic.Update(di.ServiceConstructorMap{
		container.MessagingClientName: func(get di.Get) interface{} {
			return trigger.client
		},
	})

	subscribeTopics := strings.TrimSpace(config.Trigger.EdgexMessageBus.SubscribeHost.SubscribeTopics)

	if len(subscribeTopics) == 0 {
		// Still allows subscribing to blank topic to receive all messages
		trigger.topics = append(trigger.topics, types.TopicChannel{Topic: subscribeTopics, Messages: make(chan types.MessageEnvelope)})
	} else {
		topics := util.DeleteEmptyAndTrim(strings.FieldsFunc(subscribeTopics, util.SplitComma))
		for _, topic := range topics {
			trigger.topics = append(trigger.topics, types.TopicChannel{Topic: topic, Messages: make(chan types.MessageEnvelope)})
		}
	}

	messageErrors := make(chan error)

	err = trigger.client.Connect()
	if err != nil {
		return nil, err
	}

	lc.Infof("Subscribing to topic(s): '%s' @ %s://%s:%d",
		subscribeTopics,
		config.Trigger.EdgexMessageBus.SubscribeHost.Protocol,
		config.Trigger.EdgexMessageBus.SubscribeHost.Host,
		config.Trigger.EdgexMessageBus.SubscribeHost.Port)

	publishTopic := config.Trigger.EdgexMessageBus.PublishHost.PublishTopic

	if len(config.Trigger.EdgexMessageBus.PublishHost.Host) > 0 {
		lc.Infof("Publishing to topic: '%s' @ %s://%s:%d",
			publishTopic,
			config.Trigger.EdgexMessageBus.PublishHost.Protocol,
			config.Trigger.EdgexMessageBus.PublishHost.Host,
			config.Trigger.EdgexMessageBus.PublishHost.Port)
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
				lc.Errorf("Failed to receive message from bus, %v", msgErr)

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
			trigger.serviceBinding.LoggingClient().Errorf("MQTT Trigger: Failed to process message on pipeline(s): %s", processErr.Error())
		}
	}()
}

func (trigger *Trigger) responseHandler(appContext interfaces.AppFunctionContext, pipeline *interfaces.FunctionPipeline) error {
	if appContext.ResponseData() != nil {
		lc := trigger.serviceBinding.LoggingClient()
		config := trigger.serviceBinding.Config()

		publishTopic, err := appContext.ApplyValues(config.Trigger.EdgexMessageBus.PublishHost.PublishTopic)

		if err != nil {
			lc.Errorf("MessageBus Trigger: Unable to format output topic '%s' for pipeline '%s': %s",
				config.Trigger.EdgexMessageBus.PublishHost.PublishTopic,
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

func (_ *Trigger) createMessagingClientConfig(localConfig sdkCommon.MessageBusConfig) types.MessageBusConfig {
	clientConfig := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     localConfig.PublishHost.Host,
			Port:     localConfig.PublishHost.Port,
			Protocol: localConfig.PublishHost.Protocol,
		},
		SubscribeHost: types.HostInfo{
			Host:     localConfig.SubscribeHost.Host,
			Port:     localConfig.SubscribeHost.Port,
			Protocol: localConfig.SubscribeHost.Protocol,
		},
		Type:     localConfig.Type,
		Optional: deepCopy(localConfig.Optional),
	}

	return clientConfig
}

func deepCopy(target map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range target {
		result[key] = value
	}
	return result
}

func (trigger *Trigger) setOptionalAuthData(messageBusConfig *types.MessageBusConfig, lc logger.LoggingClient) error {
	authMode := strings.ToLower(strings.TrimSpace(messageBusConfig.Optional[bootstrapMessaging.AuthModeKey]))
	if len(authMode) == 0 || authMode == bootstrapMessaging.AuthModeNone {
		return nil
	}

	secretName := messageBusConfig.Optional[bootstrapMessaging.SecretNameKey]

	lc.Infof("Setting options for secure MessageBus with AuthMode='%s' and SecretName='%s", authMode, secretName)

	secretProvider := trigger.serviceBinding.SecretProvider()
	if secretProvider == nil {
		return errors.New("secret provider is missing. Make sure it is specified to be used in bootstrap.Run()")
	}

	secretData, err := bootstrapMessaging.GetSecretData(authMode, secretName, secretProvider)
	if err != nil {
		return fmt.Errorf("unable to get Secret Data for secure message bus: %w", err)
	}

	if err := bootstrapMessaging.ValidateSecretData(authMode, secretName, secretData); err != nil {
		return fmt.Errorf("secret Data for secure message bus invalid: %w", err)
	}

	if messageBusConfig.Optional == nil {
		messageBusConfig.Optional = map[string]string{}
	}

	// Since already validated, these are the only modes that can be set at this point.
	switch authMode {
	case bootstrapMessaging.AuthModeUsernamePassword:
		messageBusConfig.Optional[bootstrapMessaging.OptionsUsernameKey] = secretData.Username
		messageBusConfig.Optional[bootstrapMessaging.OptionsPasswordKey] = secretData.Password
	case bootstrapMessaging.AuthModeCert:
		messageBusConfig.Optional[bootstrapMessaging.OptionsCertPEMBlockKey] = string(secretData.CertPemBlock)
		messageBusConfig.Optional[bootstrapMessaging.OptionsKeyPEMBlockKey] = string(secretData.KeyPemBlock)
	case bootstrapMessaging.AuthModeCA:
		messageBusConfig.Optional[bootstrapMessaging.OptionsCaPEMBlockKey] = string(secretData.CaPemBlock)
	}

	return nil
}
