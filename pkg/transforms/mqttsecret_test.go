//
// Copyright (c) 2021 Intel Corporation
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

// This test will only be executed if the tag brokerRunning is added when running
// the tests with a command like:
// go test -tags brokerRunning
package transforms

import (
	"errors"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMQTTSecretSender_setRetryDataPersistFalse(t *testing.T) {
	ctx.SetRetryData(nil)
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, false)
	sender.mqttConfig = MQTTSecretConfig{}
	sender.setRetryData(ctx, []byte("data"))
	assert.Nil(t, ctx.RetryData())
}

func TestMQTTSecretSender_setRetryDataPersistTrue(t *testing.T) {
	ctx.SetRetryData(nil)
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, true)
	sender.mqttConfig = MQTTSecretConfig{}
	sender.setRetryData(ctx, []byte("data"))
	assert.Equal(t, []byte("data"), ctx.RetryData())
}

func TestMQTTSecretSender_MQTTSendNodata(t *testing.T) {
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, true)
	sender.mqttConfig = MQTTSecretConfig{}
	continuePipeline, result := sender.MQTTSend(ctx, nil)
	require.False(t, continuePipeline)
	require.Error(t, result.(error))
}

func TestMQTTSecretSender_waitForPublishCompleted(t *testing.T) {
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, true)
	token := &publishToken{completed: true}

	err := sender.waitForPublish(token)

	require.NoError(t, err)
	require.Equal(t, defaultMQTTPublishTimeout, token.timeout)
}

func TestMQTTSecretSender_waitForPublishError(t *testing.T) {
	expected := errors.New("publish failed")
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, true)
	token := &publishToken{completed: true, err: expected}

	err := sender.waitForPublish(token)

	require.ErrorIs(t, err, expected)
}

func TestMQTTSecretSender_waitForPublishTimeout(t *testing.T) {
	sender := NewMQTTSecretSender(MQTTSecretConfig{PublishTimeout: "5ms"}, true)
	token := &publishToken{completed: false}

	err := sender.waitForPublish(token)

	require.ErrorContains(t, err, "timed out after 5ms")
	require.Equal(t, 5*time.Millisecond, token.timeout)
}

var _ MQTT.Token = (*publishToken)(nil)

type publishToken struct {
	completed bool
	err       error
	timeout   time.Duration
	done      chan struct{}
}

func (token *publishToken) Wait() bool {
	return token.completed
}

func (token *publishToken) WaitTimeout(timeout time.Duration) bool {
	token.timeout = timeout
	return token.completed
}

func (token *publishToken) Done() <-chan struct{} {
	if token.done == nil {
		token.done = make(chan struct{})
		if token.completed {
			close(token.done)
		}
	}
	return token.done
}

func (token *publishToken) Error() error {
	return token.err
}
