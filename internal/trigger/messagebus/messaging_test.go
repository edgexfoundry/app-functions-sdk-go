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

package messagebus

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/trigger/messagebus/mocks"
	interfaceMocks "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"

	"github.com/stretchr/testify/mock"

	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"
	triggerMocks "github.com/edgexfoundry/app-functions-sdk-go/v3/internal/trigger/mocks"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note the constant TriggerTypeMessageBus can not be used due to cyclic imports
const TriggerTypeMessageBus = "EDGEX-MESSAGEBUS"

var dic *di.Container

func TestMain(m *testing.M) {
	mockMessageClient := &mocks.MessageClient{}
	mockMessageClient.On("Connect").Return(nil)
	mockMessageClient.On("Subscribe", mock.Anything, mock.Anything).Return(nil)

	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
			return mockMessageClient
		},
	})

	os.Exit(m.Run())
}

func TestInitialize(t *testing.T) {

	config := sdkCommon.ConfigurationStruct{
		MessageBus: bootstrapConfig.MessageBusInfo{
			Topics: map[string]string{
				internal.MessageBusSubscribeTopics: "events",
				internal.MessageBusPublishTopic:    "publish",
			},
		},

		Trigger: sdkCommon.TriggerInfo{
			Type: TriggerTypeMessageBus,
		},
	}

	serviceBinding := &triggerMocks.ServiceBinding{}
	serviceBinding.On("Config").Return(&config)
	serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

	messageProcessor := &triggerMocks.MessageProcessor{}
	messageProcessor.On("ReceivedInvalidMessage")

	trigger := NewTrigger(serviceBinding, messageProcessor, dic)

	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)
	assert.NotNil(t, trigger.client, "Expected client to be set")
	assert.Equal(t, 1, len(trigger.topics))
	assert.Equal(t, "events", trigger.topics[0].Topic)
	assert.NotNil(t, trigger.topics[0].Messages)
}

func TestInitializeBadConfiguration(t *testing.T) {

	config := sdkCommon.ConfigurationStruct{
		Trigger: sdkCommon.TriggerInfo{
			Type: "junk",
		},
	}

	serviceBinding := &triggerMocks.ServiceBinding{}
	serviceBinding.On("Config").Return(&config)
	serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

	messageProcessor := &triggerMocks.MessageProcessor{}
	messageProcessor.On("ReceivedInvalidMessage")

	trigger := NewTrigger(serviceBinding, messageProcessor, dic)

	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	assert.Error(t, err)
}

func TestTrigger_responseHandler(t *testing.T) {
	const topicWithPlaceholder = "/topic/with/{ph}/placeholder"
	const formattedTopic = "topic/with/ph-value/placeholder"
	const setContentType = "content-type"
	const correlationId = "corrid-1233523"
	var setContentTypePayload = []byte("not-empty")
	var inferJsonPayload = []byte("{not-empty")
	var inferJsonArrayPayload = []byte("[not-empty")

	type fields struct {
		publishTopic string
	}
	type args struct {
		pipeline *interfaces.FunctionPipeline
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		setup   func(*triggerMocks.ServiceBinding, *interfaceMocks.AppFunctionContext, *mocks.MessageClient)
	}{
		{name: "no response data", wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, _ *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(nil)
		}},
		{name: "topic format failed", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: true, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, _ *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(setContentTypePayload)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return("", fmt.Errorf("apply values failed"))
		}},
		{name: "publish failed", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: true, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(setContentTypePayload)
			functionContext.On("ResponseContentType").Return(setContentType)
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				return fmt.Errorf("publish failed")
			})
		}},
		{name: "happy", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(setContentTypePayload)
			functionContext.On("ResponseContentType").Return(setContentType)
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				assert.Equal(t, correlationId, envelope.CorrelationID)
				assert.Equal(t, setContentType, envelope.ContentType)
				assert.Equal(t, setContentTypePayload, envelope.Payload)
				return nil
			})
		}},
		{name: "happy assume CBOR", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(setContentTypePayload)
			functionContext.On("ResponseContentType").Return("")
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				assert.Equal(t, correlationId, envelope.CorrelationID)
				assert.Equal(t, common.ContentTypeCBOR, envelope.ContentType)
				assert.Equal(t, setContentTypePayload, envelope.Payload)
				return nil
			})
		}},
		{name: "happy infer JSON", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(inferJsonPayload)
			functionContext.On("ResponseContentType").Return("")
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				assert.Equal(t, correlationId, envelope.CorrelationID)
				assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
				assert.Equal(t, inferJsonPayload, envelope.Payload)
				return nil
			})
		}},
		{name: "happy infer JSON array", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(inferJsonArrayPayload)
			functionContext.On("ResponseContentType").Return("")
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				assert.Equal(t, correlationId, envelope.CorrelationID)
				assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
				assert.Equal(t, inferJsonArrayPayload, envelope.Payload)
				return nil
			})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceBinding := &triggerMocks.ServiceBinding{}
			serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

			ctx := &interfaceMocks.AppFunctionContext{}
			client := &mocks.MessageClient{}

			if tt.setup != nil {
				tt.setup(serviceBinding, ctx, client)
			}

			trigger := &Trigger{
				serviceBinding:   serviceBinding,
				publishTopic:     tt.fields.publishTopic,
				messageProcessor: &triggerMocks.MessageProcessor{},
				client:           client,
			}
			if err := trigger.responseHandler(ctx, tt.args.pipeline); (err != nil) != tt.wantErr {
				t.Errorf("responseHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
