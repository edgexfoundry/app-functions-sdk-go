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

package mqtt

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/trigger"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/secure"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	commonContracts "github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	defaultRetryDuration = 600
	defaultRetryInterval = 5
)

// Trigger implements Trigger to support Triggers
type Trigger struct {
	messageProcessor trigger.MessageProcessor
	serviceBinding   trigger.ServiceBinding
	mqttClient       pahoMqtt.Client
	qos              byte
	retain           bool
	publishTopic     string
}

func NewTrigger(bnd trigger.ServiceBinding, mp trigger.MessageProcessor) *Trigger {
	t := &Trigger{
		messageProcessor: mp,
		serviceBinding:   bnd,
	}

	return t
}

// Initialize initializes the Trigger for an external MQTT broker
func (trigger *Trigger) Initialize(_ *sync.WaitGroup, ctx context.Context, background <-chan interfaces.BackgroundMessage) (bootstrap.Deferred, error) {
	// Convenience shortcuts
	lc := trigger.serviceBinding.LoggingClient()
	config := trigger.serviceBinding.Config()

	brokerConfig := config.Trigger.ExternalMqtt
	topics := config.Trigger.SubscribeTopics

	trigger.qos = brokerConfig.QoS
	trigger.retain = brokerConfig.Retain
	trigger.publishTopic = config.Trigger.PublishTopic

	lc.Info("Initializing MQTT Trigger")

	if background != nil {
		return nil, errors.New("background publishing not supported for services using MQTT trigger")
	}

	if len(strings.TrimSpace(topics)) == 0 {
		return nil, fmt.Errorf("missing SubscribeTopics for MQTT Trigger. Must be present in [Trigger.ExternalMqtt] section")
	}

	brokerUrl, err := url.Parse(brokerConfig.Url)
	if err != nil {
		return nil, fmt.Errorf("invalid MQTT Broker Url '%s': %s", config.Trigger.ExternalMqtt.Url, err.Error())
	}

	opts := pahoMqtt.NewClientOptions()
	opts.AutoReconnect = brokerConfig.AutoReconnect
	opts.OnConnect = trigger.onConnectHandler
	opts.ClientID = brokerConfig.ClientId
	if len(brokerConfig.ConnectTimeout) > 0 {
		duration, err := time.ParseDuration(brokerConfig.ConnectTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid MQTT ConnectTimeout '%s': %s", brokerConfig.ConnectTimeout, err.Error())
		}
		opts.ConnectTimeout = duration
	}
	opts.KeepAlive = brokerConfig.KeepAlive
	opts.Servers = []*url.URL{brokerUrl}

	will := brokerConfig.Will
	if will.Enabled {
		opts.SetWill(will.Topic, will.Payload, will.Qos, will.Retained)
		lc.Infof("Last Will options set for MQTT Trigger: %+v", will)
	}

	if brokerConfig.RetryDuration <= 0 {
		brokerConfig.RetryDuration = defaultRetryDuration
	}
	if brokerConfig.RetryInterval <= 0 {
		brokerConfig.RetryInterval = defaultRetryInterval
	}

	sp := trigger.serviceBinding.SecretProvider()
	var mqttClient pahoMqtt.Client
	timer := startup.NewTimer(brokerConfig.RetryDuration, brokerConfig.RetryInterval)

	for timer.HasNotElapsed() {
		if mqttClient, err = createMqttClient(sp, lc, brokerConfig, opts); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, errors.New("aborted MQTT Trigger initialization")
		default:
			lc.Warnf("%s. Attempt to create MQTT client again after %d seconds...", err.Error(), brokerConfig.RetryInterval)
			timer.SleepForInterval()
		}
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create MQTT Client: %s", err.Error())
	}

	deferred := func() {
		lc.Info("Disconnecting from broker for MQTT trigger")
		trigger.mqttClient.Disconnect(0)
	}

	trigger.mqttClient = mqttClient

	return deferred, nil
}

func (trigger *Trigger) onConnectHandler(mqttClient pahoMqtt.Client) {
	// Convenience shortcuts
	lc := trigger.serviceBinding.LoggingClient()
	config := trigger.serviceBinding.Config()
	topics := util.DeleteEmptyAndTrim(strings.FieldsFunc(config.Trigger.SubscribeTopics, util.SplitComma))
	topicMap := map[string]byte{}
	for _, topic := range topics {
		topicMap[topic] = config.Trigger.ExternalMqtt.QoS
	}
	if token := mqttClient.SubscribeMultiple(topicMap, trigger.messageHandler); token.Wait() && token.Error() != nil {
		mqttClient.Disconnect(0)
		lc.Errorf("could not subscribe to topics '%v' for MQTT trigger: %s \n",
			topicMap, token.Error().Error())
	}

	lc.Infof("Subscribed to topic(s) '%s' for MQTT trigger", config.Trigger.SubscribeTopics)
}

func (trigger *Trigger) messageHandler(_ pahoMqtt.Client, mqttMessage pahoMqtt.Message) {
	// Convenience shortcuts
	lc := trigger.serviceBinding.LoggingClient()

	data := mqttMessage.Payload()
	contentType := commonContracts.ContentTypeJSON
	if data[0] != byte('{') && data[0] != byte('[') {
		// If not JSON then assume it is CBOR
		contentType = commonContracts.ContentTypeCBOR
	}

	correlationID := uuid.New().String()

	message := types.MessageEnvelope{
		CorrelationID: correlationID,
		ContentType:   contentType,
		Payload:       data,
		ReceivedTopic: mqttMessage.Topic(),
	}

	lc.Debugf("MQTT Trigger: Received message with %d bytes on topic '%s'. Content-Type=%s",
		len(message.Payload),
		message.ReceivedTopic,
		message.ContentType)
	lc.Tracef("%s=%s", commonContracts.CorrelationHeader, correlationID)

	ctx := trigger.serviceBinding.BuildContext(message)

	go func() {
		processErr := trigger.messageProcessor.MessageReceived(ctx, message, trigger.responseHandler)
		if processErr != nil {
			lc.Errorf("MQTT Trigger: Failed to process message on pipeline(s): %s", processErr.Error())
		}
	}()
}

func (trigger *Trigger) responseHandler(appContext interfaces.AppFunctionContext, pipeline *interfaces.FunctionPipeline) error {
	if len(appContext.ResponseData()) > 0 && len(trigger.publishTopic) > 0 {
		lc := trigger.serviceBinding.LoggingClient()

		formattedTopic, err := appContext.ApplyValues(trigger.publishTopic)

		if err != nil {
			lc.Errorf("MQTT trigger: Unable to format topic '%s' for pipeline '%s': %s",
				trigger.publishTopic,
				pipeline.Id,
				err.Error())
			return err
		}

		if token := trigger.mqttClient.Publish(formattedTopic, trigger.qos, trigger.retain, appContext.ResponseData()); token.Wait() && token.Error() != nil {
			lc.Errorf("MQTT trigger: Could not publish to topic '%s' for pipeline '%s': %s",
				formattedTopic,
				pipeline.Id,
				token.Error())
			return token.Error()
		} else {
			lc.Debugf("MQTT Trigger: Published response message for pipeline '%s' on topic '%s' with %d bytes",
				pipeline.Id,
				formattedTopic,
				len(appContext.ResponseData()))
			lc.Tracef("MQTT Trigger published message: %s=%s", commonContracts.CorrelationHeader, appContext.CorrelationID())
		}
	}
	return nil
}

func createMqttClient(sp messaging.SecretDataProvider, lc logger.LoggingClient, config common.ExternalMqttConfig,
	opts *pahoMqtt.ClientOptions) (pahoMqtt.Client, error) {
	mqttFactory := secure.NewMqttFactory(
		sp,
		lc,
		config.AuthMode,
		config.SecretName,
		config.SkipCertVerify,
	)
	mqttClient, err := mqttFactory.Create(opts)
	if err != nil {
		return nil, fmt.Errorf("unable to create secure MQTT Client: %s", err.Error())
	}

	lc.Infof("Connecting to mqtt broker for MQTT trigger at: %s", config.Url)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("could not connect to broker for MQTT trigger: %s", token.Error().Error())
	}

	lc.Info("Connected to mqtt server for MQTT trigger")
	return mqttClient, nil
}
