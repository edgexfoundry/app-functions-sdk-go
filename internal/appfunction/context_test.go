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

package appfunction

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/bootstrap/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"

	appCommon "github.com/edgexfoundry/app-functions-sdk-go/v4/internal/common"
	bootstrapMocks "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	messageMocks "github.com/edgexfoundry/go-mod-messaging/v4/messaging/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var target *Context
var dic *di.Container
var baseUrl = "http://localhost:"
var mockSecretProvider interfaces.SecretProvider

func TestMain(m *testing.M) {
	mockSecretProvider = &mocks.SecretProvider{}

	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSecretProvider
		},
	})
	target = NewContext("", dic, "")

	os.Exit(m.Run())
}

func TestContext_EventClient(t *testing.T) {
	actual := target.EventClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.EventClientName: func(get di.Get) interface{} {
			return clients.NewEventClient(baseUrl+"59880", nil, false)
		},
	})

	actual = target.EventClient()
	assert.NotNil(t, actual)
}

func TestContext_ReadingClient(t *testing.T) {
	actual := target.ReadingClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.ReadingClientName: func(get di.Get) interface{} {
			return clients.NewReadingClient(baseUrl+"59880", nil, false)
		},
	})

	actual = target.ReadingClient()
	assert.NotNil(t, actual)
}

func TestContext_CommandClient(t *testing.T) {
	actual := target.CommandClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.CommandClientName: func(get di.Get) interface{} {
			return clients.NewCommandClient(baseUrl+"59882", nil, false)
		},
	})

	actual = target.CommandClient()
	assert.NotNil(t, actual)
}

func TestContext_DeviceServiceClient(t *testing.T) {
	actual := target.DeviceServiceClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceServiceClientName: func(get di.Get) interface{} {
			return clients.NewDeviceServiceClient(baseUrl+"59881", nil, false)
		},
	})

	actual = target.DeviceServiceClient()
	assert.NotNil(t, actual)

}

func TestContext_DeviceProfileClient(t *testing.T) {
	actual := target.DeviceProfileClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return clients.NewDeviceProfileClient(baseUrl+"59881", nil, false)
		},
	})

	actual = target.DeviceProfileClient()
	assert.NotNil(t, actual)
}

func TestContext_DeviceClient(t *testing.T) {
	actual := target.DeviceClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return clients.NewDeviceClient(baseUrl+"59881", nil, false)
		},
	})

	actual = target.DeviceClient()
	assert.NotNil(t, actual)

}

func TestContext_NotificationClient(t *testing.T) {
	actual := target.NotificationClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.NotificationClientName: func(get di.Get) interface{} {
			return clients.NewNotificationClient(baseUrl+"59860", nil, false)
		},
	})

	actual = target.NotificationClient()
	assert.NotNil(t, actual)

}

func TestContext_SubscriptionClient(t *testing.T) {
	actual := target.SubscriptionClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SubscriptionClientName: func(get di.Get) interface{} {
			return clients.NewSubscriptionClient(baseUrl+"59860", nil, false)
		},
	})

	actual = target.SubscriptionClient()
	assert.NotNil(t, actual)
}

func TestContext_MetricsManager(t *testing.T) {
	actual := target.MetricsManager()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return &mocks.MetricsManager{}
		},
	})

	actual = target.MetricsManager()
	assert.NotNil(t, actual)
}

func TestContext_SecretProvider(t *testing.T) {
	actual := target.SecretProvider()
	require.NotNil(t, actual)
	assert.Equal(t, mockSecretProvider, actual)
}

func TestContext_LoggingClient(t *testing.T) {
	actual := target.LoggingClient()
	assert.NotNil(t, actual)
}

func TestContext_CorrelationID(t *testing.T) {
	expected := "123-3456"
	target.correlationID = expected

	actual := target.CorrelationID()

	assert.Equal(t, expected, actual)
}

func TestContext_SetCorrelationID(t *testing.T) {
	expected := "567-098"

	target.SetCorrelationID(expected)
	actual := target.correlationID

	assert.Equal(t, expected, actual)
}

func TestContext_InputContentType(t *testing.T) {
	expected := common.ContentTypeXML
	target.inputContentType = expected

	actual := target.InputContentType()

	assert.Equal(t, expected, actual)
}

func TestContext_SetInputContentType(t *testing.T) {
	expected := common.ContentTypeCBOR

	target.SetInputContentType(expected)
	actual := target.inputContentType

	assert.Equal(t, expected, actual)
}

func TestContext_ResponseContentType(t *testing.T) {
	expected := common.ContentTypeJSON
	target.responseContentType = expected

	actual := target.ResponseContentType()

	assert.Equal(t, expected, actual)
}

func TestContext_SetResponseContentType(t *testing.T) {
	expected := common.ContentTypeText

	target.SetResponseContentType(expected)
	actual := target.responseContentType

	assert.Equal(t, expected, actual)
}

func TestContext_SetResponseData(t *testing.T) {
	expected := []byte("response data")

	target.SetResponseData(expected)
	actual := target.responseData

	assert.Equal(t, expected, actual)
}

func TestContext_ResponseData(t *testing.T) {
	expected := []byte("response data")
	target.responseData = expected

	actual := target.ResponseData()

	assert.Equal(t, expected, actual)
}

func TestContext_SetRetryData(t *testing.T) {
	expected := []byte("retry data")

	target.SetRetryData(expected)
	actual := target.retryData

	assert.Equal(t, expected, actual)
}

func TestContext_RetryData(t *testing.T) {
	expected := []byte("retry data")
	target.retryData = expected

	actual := target.RetryData()

	assert.Equal(t, expected, actual)
}

func TestContext_SecretsLastUpdated(t *testing.T) {
	expected := time.Now()
	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("SecretsLastUpdated").Return(expected, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSecretProvider
		},
	})

	actual := target.SecretProvider().SecretsLastUpdated()
	assert.Equal(t, expected, actual)
}

func TestContext_AddValue(t *testing.T) {
	k := uuid.NewString()
	v := uuid.NewString()

	target.AddValue(k, v)

	res, found := target.contextData[strings.ToLower(k)]

	require.True(t, found, "item should be present in context map")
	require.Equal(t, v, res, "and it should be what we put there")
}

func TestContext_GetValue(t *testing.T) {
	k := uuid.NewString()
	v := uuid.NewString()

	target.contextData[strings.ToLower(k)] = v

	res, found := target.GetValue(k)

	require.True(t, found, "indicate item found in context map")
	require.Equal(t, v, res, "and it should be what we put there")
}

func TestContext_GetValue_NotPresent(t *testing.T) {
	k := uuid.NewString()

	res, found := target.GetValue(k)

	require.False(t, found, "should indicate item not found in context map")
	require.Equal(t, "", res, "and default string is returned")
}

func TestContext_RemoveValue(t *testing.T) {
	k := uuid.NewString()
	v := uuid.NewString()

	target.contextData[k] = v

	target.RemoveValue(k)

	_, found := target.contextData[strings.ToLower(k)]

	require.False(t, found, "item should not be present in context map")
}

func TestContext_RemoveValue_Not_Present(t *testing.T) {
	k := uuid.NewString()

	_, found := target.contextData[strings.ToLower(k)]

	require.False(t, found, "item should not be present in context map")

	target.RemoveValue(k)
}

func TestContext_GetAllValues(t *testing.T) {
	orig := map[string]string{
		"key1": "val",
		"key2": "val2",
	}

	target.contextData = orig

	res := target.GetAllValues()

	// pointers used to compare underlying memory
	require.NotSame(t, &orig, &res, "Returned map should be a copy")

	for k, v := range orig {
		assert.Equal(t, v, res[k], fmt.Sprintf("Source and result do not match at key %s", k))
	}
}

func TestContext_ApplyValues_No_Placeholders(t *testing.T) {
	data := map[string]string{
		"key1": "val",
		"key2": "val2",
	}

	input := uuid.NewString()

	target.contextData = data

	res, err := target.ApplyValues(input)

	require.NoError(t, err)
	require.Equal(t, res, input)
}

func TestContext_ApplyValues_Placeholders(t *testing.T) {
	data := map[string]string{
		"key1": "val",
		"key2": "val2",
	}

	input := "{key1}-{key2}"

	target.contextData = data

	res, err := target.ApplyValues(input)

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s-%s", data["key1"], data["key2"]), res)
}

func TestContext_ApplyValues_MissingPlaceholder(t *testing.T) {
	data := map[string]string{
		"key1": "val",
		"key2": "val2",
	}

	input := "{key1}-{key2}-{key3}"

	target.contextData = data

	res, err := target.ApplyValues(input)

	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("failed to replace all context placeholders in input ('%s-%s-{key3}' after replacements)", data["key1"], data["key2"]), err.Error())
	require.Equal(t, "", res)
}

func TestContext_GetDeviceResource(t *testing.T) {
	mockClient := clientMocks.DeviceProfileClient{}
	mockClient.On("DeviceResourceByProfileNameAndResourceName", mock.Anything, mock.Anything, mock.Anything).Return(responses.DeviceResourceResponse{}, nil)
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return &mockClient
		},
	})

	_, err := target.GetDeviceResource("MyProfile", "MyResource")
	require.NoError(t, err)
}

func TestContext_GetDeviceResource_Error(t *testing.T) {
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return nil
		},
	})

	_, err := target.GetDeviceResource("MyProfile", "MyResource")
	require.Error(t, err)
}

func TestContext_Publish(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		publishErr      error
		triggerConfig   *appCommon.ConfigurationStruct
		bootstrapConfig config.BootstrapConfiguration
		expectedError   error
		expectedTopic   string
	}{
		{
			name: "No message bus config",
			triggerConfig: &appCommon.ConfigurationStruct{
				Trigger: appCommon.TriggerInfo{
					PublishTopic: "test",
				},
			},
			bootstrapConfig: config.BootstrapConfiguration{
				MessageBus: &config.MessageBusInfo{
					Disabled:        true,
					BaseTopicPrefix: "test",
				},
			},
			message:       "test",
			expectedTopic: "test/test",
			expectedError: errors.New(messageBusDisabledErr),
		},
		{
			name: "valid publish",
			triggerConfig: &appCommon.ConfigurationStruct{
				Trigger: appCommon.TriggerInfo{
					PublishTopic: "test",
				},
			},
			bootstrapConfig: config.BootstrapConfiguration{
				MessageBus: &config.MessageBusInfo{
					Disabled:        false,
					BaseTopicPrefix: "test",
				},
			},
			expectedTopic: "test/test",
			message:       "test",
			publishErr:    nil,
			expectedError: nil,
		},
		{
			name: "publish error",
			triggerConfig: &appCommon.ConfigurationStruct{
				Trigger: appCommon.TriggerInfo{
					PublishTopic: "test",
				},
			},
			bootstrapConfig: config.BootstrapConfiguration{
				MessageBus: &config.MessageBusInfo{
					Disabled:        false,
					BaseTopicPrefix: "test",
				},
			},
			expectedTopic: "test/test",
			message:       "test",
			publishErr:    errors.New("failed"),
			expectedError: errors.New("failed to publish data to messagebus: failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockMessageClient *messageMocks.MessageClient
			var mockConfig *bootstrapMocks.Configuration
			mockMessageClient = nil
			if !tt.bootstrapConfig.MessageBus.Disabled {
				mockMessageClient = &messageMocks.MessageClient{}
				mockMessageClient.On("Connect").Return(nil)
				mockMessageClient.On("Publish", mock.Anything, tt.expectedTopic).Return(tt.publishErr)
				mockConfig = &bootstrapMocks.Configuration{}
				mockConfig.On("GetBootstrap").Return(tt.bootstrapConfig)
			}

			publishDic := di.NewContainer(di.ServiceConstructorMap{
				bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
					return logger.NewMockClient()
				},
				bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
					// ensures nil pointer is returned in PublishWIthTopic
					if mockMessageClient == nil {
						return nil
					}
					return mockMessageClient
				},
				container.ConfigurationName: func(get di.Get) interface{} {
					return tt.triggerConfig
				},
				bootstrapContainer.ConfigurationInterfaceName: func(get di.Get) interface{} {
					return mockConfig
				},
			})

			appContext := NewContext("", publishDic, "")

			err := appContext.Publish(tt.message, common.ContentTypeJSON)
			require.Equal(t, tt.expectedError, err)
		})
	}
}

func TestService_PublishWithTopic(t *testing.T) {
	tests := []struct {
		name            string
		topic           string
		message         string
		publishErr      error
		expectedTopic   string
		expectedError   error
		bootstrapConfig config.BootstrapConfiguration
		wantErr         bool
	}{
		{
			name: "No message bus config",
			bootstrapConfig: config.BootstrapConfiguration{
				MessageBus: &config.MessageBusInfo{
					Disabled: true,
				},
			},
			message:       "test",
			expectedError: errors.New(messageBusDisabledErr),
			wantErr:       true,
		},
		{
			name: "valid publish",
			bootstrapConfig: config.BootstrapConfiguration{
				MessageBus: &config.MessageBusInfo{
					Disabled:        false,
					BaseTopicPrefix: "test",
				},
			},
			topic:         "test_topic",
			expectedTopic: "test/test_topic",
			message:       "test",
			publishErr:    nil,
			wantErr:       false,
		},
		{
			name: "publish error",
			bootstrapConfig: config.BootstrapConfiguration{
				MessageBus: &config.MessageBusInfo{
					Disabled:        false,
					BaseTopicPrefix: "test",
				},
			},
			topic:         "test_topic",
			expectedTopic: "test/test_topic",
			message:       "test",
			publishErr:    errors.New("failed"),
			expectedError: errors.New("failed to publish data to messagebus: failed"),
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockMessageClient *messageMocks.MessageClient
			var mockConfig *bootstrapMocks.Configuration
			if !tt.bootstrapConfig.MessageBus.Disabled {
				mockMessageClient = &messageMocks.MessageClient{}
				mockMessageClient.On("Connect").Return(nil)
				mockMessageClient.On("Publish", mock.Anything, tt.expectedTopic).Return(tt.publishErr)
				mockConfig = &bootstrapMocks.Configuration{}
				mockConfig.On("GetBootstrap").Return(tt.bootstrapConfig)
			}

			publishDic := di.NewContainer(di.ServiceConstructorMap{
				bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
					return logger.NewMockClient()
				},
				bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
					// ensures nil pointer is returned in PublishWIthTopic
					if mockMessageClient == nil {
						return nil
					}
					return mockMessageClient
				},
				bootstrapContainer.ConfigurationInterfaceName: func(get di.Get) interface{} {
					return mockConfig
				},
			})

			appContext := NewContext("", publishDic, "")

			err := appContext.PublishWithTopic(tt.topic, tt.message, common.ContentTypeJSON)
			require.Equal(t, tt.expectedError, err)

		})
	}
}

func TestContext_Clone(t *testing.T) {
	sut := Context{
		Dic:                 dic,
		correlationID:       uuid.NewString(),
		inputContentType:    common.ContentTypeJSON,
		responseData:        nil,
		retryData:           nil,
		responseContentType: common.ContentTypeJSON,
		contextData: map[string]string{
			"test":  "val1",
			"test2": "val2",
		},
		valuePlaceholderSpec: regexp.MustCompile(""),
	}

	clone, ok := sut.Clone().(*Context)

	require.True(t, ok)
	require.NotNil(t, clone)
	require.NotSame(t, sut, clone)

	assert.Equal(t, sut.correlationID, clone.correlationID)
	assert.Equal(t, sut.inputContentType, clone.inputContentType)
	assert.Equal(t, sut.responseData, clone.responseData)
	assert.Equal(t, sut.retryData, clone.retryData)
	assert.Equal(t, sut.responseContentType, clone.responseContentType)
	assert.Equal(t, sut.valuePlaceholderSpec, clone.valuePlaceholderSpec)

	require.NotSame(t, sut.contextData, clone.contextData)

	assert.Equal(t, len(sut.contextData), len(clone.contextData))

	for k, v := range sut.contextData {
		assert.Equal(t, v, clone.contextData[k])
	}
}
