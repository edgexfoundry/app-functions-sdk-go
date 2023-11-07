//
// Copyright (c) 2023 Intel Corporation
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

package transforms

import (
	"fmt"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"
	coreCommon "github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	gometrics "github.com/rcrowley/go-metrics"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/secure"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/util"
)

// MQTTSecretSender ...
type MQTTSecretSender struct {
	lock                 sync.Mutex
	client               MQTT.Client
	mqttConfig           MQTTSecretConfig
	persistOnError       bool
	opts                 *MQTT.ClientOptions
	secretsLastRetrieved time.Time
	topicFormatter       StringValuesFormatter
	mqttSizeMetrics      gometrics.Histogram
	mqttErrorMetric      gometrics.Counter
}

// MQTTSecretConfig ...
type MQTTSecretConfig struct {
	// BrokerAddress should be set to the complete broker address i.e. mqtts://mosquitto:8883/mybroker
	BrokerAddress string
	// ClientId to connect with the broker with.
	ClientId string
	// The name of the secret in secret provider to retrieve your secrets
	SecretName string
	// AutoReconnect indicated whether or not to retry connection if disconnected
	AutoReconnect bool
	// KeepAlive is the interval duration between client sending keepalive ping to broker
	KeepAlive string
	// ConnectTimeout is the duration for timing out on connecting to the broker
	ConnectTimeout string
	// Topic that you wish to publish to
	Topic string
	// QoS for MQTT Connection
	QoS byte
	// Retain setting for MQTT Connection
	Retain bool
	// SkipCertVerify
	SkipCertVerify bool
	// AuthMode indicates what to use when connecting to the broker. Options are "none", "cacert" , "usernamepassword", "clientcert".
	// If a CA Cert exists in the SecretName then it will be used for all modes except "none".
	AuthMode string
	// Will contains the Last Will configuration for the MQTT Client
	Will common.WillConfig
}

// NewMQTTSecretSender ...
func NewMQTTSecretSender(mqttConfig MQTTSecretConfig, persistOnError bool) *MQTTSecretSender {
	opts := MQTT.NewClientOptions()

	opts.AddBroker(mqttConfig.BrokerAddress)
	opts.SetClientID(mqttConfig.ClientId)
	opts.SetAutoReconnect(mqttConfig.AutoReconnect)

	//avoid casing issues
	mqttConfig.AuthMode = strings.ToLower(mqttConfig.AuthMode)
	sender := &MQTTSecretSender{
		client:         nil,
		mqttConfig:     mqttConfig,
		persistOnError: persistOnError,
		opts:           opts,
	}

	return sender
}

// NewMQTTSecretSenderWithTopicFormatter allows passing a function to build a final publish topic
// from the combination of the configured topic and the input parameters passed to MQTTSend
func NewMQTTSecretSenderWithTopicFormatter(mqttConfig MQTTSecretConfig, persistOnError bool, topicFormatter StringValuesFormatter) *MQTTSecretSender {
	sender := NewMQTTSecretSender(mqttConfig, persistOnError)
	sender.topicFormatter = topicFormatter
	return sender
}

func (sender *MQTTSecretSender) initializeMQTTClient(ctx interfaces.AppFunctionContext) error {
	sender.lock.Lock()
	defer sender.lock.Unlock()

	// If the conditions changed while waiting for the lock, i.e. other thread completed the initialization,
	// then skip doing anything
	secretProvider := ctx.SecretProvider()
	if sender.client != nil && !sender.secretsLastRetrieved.Before(secretProvider.SecretsLastUpdated()) {
		return nil
	}

	ctx.LoggingClient().Info("Initializing MQTT Client")

	config := sender.mqttConfig
	mqttFactory := secure.NewMqttFactory(ctx.SecretProvider(), ctx.LoggingClient(), config.AuthMode, config.SecretName, config.SkipCertVerify)

	if len(sender.mqttConfig.KeepAlive) > 0 {
		keepAlive, err := time.ParseDuration(sender.mqttConfig.KeepAlive)
		if err != nil {
			return fmt.Errorf("in pipeline '%s', unable to parse KeepAlive value of '%s': %s", ctx.PipelineId(), sender.mqttConfig.KeepAlive, err.Error())
		}

		sender.opts.SetKeepAlive(keepAlive)
	}

	if len(sender.mqttConfig.ConnectTimeout) > 0 {
		timeout, err := time.ParseDuration(sender.mqttConfig.ConnectTimeout)
		if err != nil {
			return fmt.Errorf("in pipeline '%s', unable to parse ConnectTimeout value of '%s': %s", ctx.PipelineId(), sender.mqttConfig.ConnectTimeout, err.Error())
		}

		sender.opts.SetConnectTimeout(timeout)
	}

	if config.Will.Enabled {
		sender.opts.SetWill(config.Will.Topic, config.Will.Payload, config.Will.Qos, config.Will.Retained)
		ctx.LoggingClient().Infof("Last Will options set for MQTT Export: %+v", config.Will)
	}

	client, err := mqttFactory.Create(sender.opts)
	if err != nil {
		return fmt.Errorf("in pipeline '%s', unable to create MQTT Client: %s", ctx.PipelineId(), err.Error())
	}

	sender.client = client
	sender.secretsLastRetrieved = time.Now()

	return nil
}

func (sender *MQTTSecretSender) connectToBroker(ctx interfaces.AppFunctionContext, exportData []byte) error {
	sender.lock.Lock()
	defer sender.lock.Unlock()

	// If other thread made the connection while this one was waiting for the lock
	// then skip trying to connect
	if sender.client.IsConnected() {
		return nil
	}

	ctx.LoggingClient().Info("Connecting to mqtt server for export")
	if token := sender.client.Connect(); token.Wait() && token.Error() != nil {
		sender.setRetryData(ctx, exportData)
		subMessage := "dropping event"
		if sender.persistOnError {
			subMessage = "persisting Event for later retry"
		}
		return fmt.Errorf("in pipeline '%s', could not connect to mqtt server for export, %s. Error: %s", ctx.PipelineId(), subMessage, token.Error().Error())
	}
	ctx.LoggingClient().Infof("Connected to mqtt server for export in pipeline '%s'", ctx.PipelineId())
	return nil
}

// MQTTSend sends data from the previous function to the specified MQTT broker.
// If no previous function exists, then the event that triggered the pipeline will be used.
func (sender *MQTTSecretSender) MQTTSend(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		// We didn't receive a result
		return false, fmt.Errorf("function MQTTSend in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	exportData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}
	// if we haven't initialized the client yet OR the cache has been invalidated (due to new/updated secrets) we need to (re)initialize the client
	secretProvider := ctx.SecretProvider()
	if sender.client == nil || sender.secretsLastRetrieved.Before(secretProvider.SecretsLastUpdated()) {
		err := sender.initializeMQTTClient(ctx)
		if err != nil {
			return false, err
		}
	}

	publishTopic, err := sender.topicFormatter.invoke(sender.mqttConfig.Topic, ctx, data)
	if err != nil {
		return false, fmt.Errorf("in pipeline '%s', MQTT topic formatting failed: %s", ctx.PipelineId(), err.Error())
	}

	tagValue := fmt.Sprintf("%s/%s", sender.mqttConfig.BrokerAddress, publishTopic)
	tag := map[string]string{"address/topic": tagValue}

	createRegisterMetric(ctx,
		func() string { return fmt.Sprintf("%s-%s", internal.MqttExportErrorsName, tagValue) },
		func() any { return sender.mqttErrorMetric },
		func() { sender.mqttErrorMetric = gometrics.NewCounter() },
		tag)

	createRegisterMetric(ctx,
		func() string { return fmt.Sprintf("%s-%s", internal.MqttExportSizeName, tagValue) },
		func() any { return sender.mqttSizeMetrics },
		func() {
			sender.mqttSizeMetrics = gometrics.NewHistogram(gometrics.NewUniformSample(internal.MetricsReservoirSize))
		},
		tag)

	if !sender.client.IsConnected() {
		err := sender.connectToBroker(ctx, exportData)
		if err != nil {
			sender.mqttErrorMetric.Inc(1)
			return false, err
		}
	}

	if !sender.client.IsConnectionOpen() {
		sender.mqttErrorMetric.Inc(1)
		sender.setRetryData(ctx, exportData)
		subMessage := "dropping event"
		if sender.persistOnError {
			subMessage = "persisting Event for later retry"
		}
		return false, fmt.Errorf("in pipeline '%s', connection to mqtt server for export not open, %s", ctx.PipelineId(), subMessage)
	}

	token := sender.client.Publish(publishTopic, sender.mqttConfig.QoS, sender.mqttConfig.Retain, exportData)
	token.Wait()
	if token.Error() != nil {
		sender.mqttErrorMetric.Inc(1)
		sender.setRetryData(ctx, exportData)
		return false, token.Error()
	}

	// capture the size for metrics
	exportDataBytes := len(exportData)
	sender.mqttSizeMetrics.Update(int64(exportDataBytes))

	ctx.LoggingClient().Debugf("Sent %d bytes of data to MQTT Broker in pipeline '%s'", exportDataBytes, ctx.PipelineId())
	ctx.LoggingClient().Tracef("Data exported", "Transport", "MQTT", "pipeline", ctx.PipelineId(), coreCommon.CorrelationHeader, ctx.CorrelationID())

	return true, nil
}

func (sender *MQTTSecretSender) setRetryData(ctx interfaces.AppFunctionContext, exportData []byte) {
	if sender.persistOnError {
		ctx.SetRetryData(exportData)
	}
}
