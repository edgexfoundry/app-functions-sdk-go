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

package rest

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDtos "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedCorrelationId = uuid.New().String()
var dic *di.Container

func TestMain(m *testing.M) {
	//secretProviderMock.on
	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return &mocks.SecretProvider{}
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &sdkCommon.ConfigurationStruct{}
		},
	})

	os.Exit(m.Run())
}

func TestPingRequest(t *testing.T) {
	serviceName := uuid.NewString()

	target := NewController(nil, dic, serviceName)

	recorder := doRequest(t, http.MethodGet, common.ApiPingRoute, target.Ping, nil)

	actual := commonDtos.PingResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	_, err = time.Parse(time.UnixDate, actual.Timestamp)
	assert.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actual.ApiVersion)
	assert.Equal(t, serviceName, actual.ServiceName)
}

func TestVersionRequest(t *testing.T) {
	serviceName := uuid.NewString()

	expectedAppVersion := "1.2.5"
	expectedSdkVersion := "1.3.1"

	internal.ApplicationVersion = expectedAppVersion
	internal.SDKVersion = expectedSdkVersion

	target := NewController(nil, dic, serviceName)

	recorder := doRequest(t, http.MethodGet, common.ApiVersion, target.Version, nil)

	actual := commonDtos.VersionSdkResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actual.ApiVersion)
	assert.Equal(t, expectedAppVersion, actual.Version)
	assert.Equal(t, expectedSdkVersion, actual.SdkVersion)
	assert.Equal(t, serviceName, actual.ServiceName)
}

func TestMetricsRequest(t *testing.T) {
	serviceName := uuid.NewString()

	target := NewController(nil, dic, serviceName)

	recorder := doRequest(t, http.MethodGet, common.ApiMetricsRoute, target.Metrics, nil)

	actual := commonDtos.MetricsResponse{
		Metrics: commonDtos.Metrics{
			MemAlloc:       math.MaxUint64,
			MemFrees:       math.MaxUint64,
			MemLiveObjects: math.MaxUint64,
			MemMallocs:     math.MaxUint64,
			MemSys:         math.MaxUint64,
			MemTotalAlloc:  math.MaxUint64,
			CpuBusyAvg:     0,
		},
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actual.ApiVersion)
	assert.Equal(t, serviceName, actual.ServiceName)

	// Since when -race flag is use some values may come back as 0 we need to use the max value to detect change
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemAlloc)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemFrees)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemLiveObjects)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemMallocs)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemSys)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemTotalAlloc)
	assert.NotEqual(t, 0, actual.Metrics.CpuBusyAvg)
}

func TestConfigRequest(t *testing.T) {
	serviceName := uuid.NewString()

	expectedConfig := sdkCommon.ConfigurationStruct{
		Writable: sdkCommon.WritableInfo{
			LogLevel: "DEBUG",
		},
		Registry: bootstrapConfig.RegistryInfo{
			Host: "localhost",
			Port: 8500,
			Type: "consul",
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &expectedConfig
		},
	})

	target := NewController(nil, dic, serviceName)

	recorder := doRequest(t, http.MethodGet, common.ApiConfigRoute, target.Config, nil)

	actualResponse := commonDtos.ConfigResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion)
	assert.Equal(t, serviceName, actualResponse.ServiceName)

	// actualResponse.Config is an interface{} so need to re-marshal/un-marshal into sdkCommon.ConfigurationStruct
	configJson, err := json.Marshal(actualResponse.Config)
	require.NoError(t, err)
	require.Less(t, 0, len(configJson))

	actualConfig := sdkCommon.ConfigurationStruct{}
	err = json.Unmarshal(configJson, &actualConfig)
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, actualConfig)
}

func TestConfigRequest_CustomConfig(t *testing.T) {
	serviceName := uuid.NewString()

	expectedConfig := sdkCommon.ConfigurationStruct{
		Writable: sdkCommon.WritableInfo{
			LogLevel: "DEBUG",
		},
		Registry: bootstrapConfig.RegistryInfo{
			Host: "localhost",
			Port: 8500,
			Type: "consul",
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &expectedConfig
		},
	})

	expectedCustomConfig := TestCustomConfig{
		"test custom config",
	}

	type fullConfig struct {
		sdkCommon.ConfigurationStruct
		CustomConfiguration TestCustomConfig
	}

	expectedFullConfig := fullConfig{
		expectedConfig,
		expectedCustomConfig,
	}

	target := NewController(nil, dic, serviceName)
	target.SetCustomConfigInfo(&expectedCustomConfig)

	recorder := doRequest(t, http.MethodGet, common.ApiConfigRoute, target.Config, nil)

	actualResponse := commonDtos.ConfigResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion)
	assert.Equal(t, serviceName, actualResponse.ServiceName)

	// actualResponse.Config is an interface{} so need to re-marshal/un-marshal into sdkCommon.ConfigurationStruct
	configJson, err := json.Marshal(actualResponse.Config)
	require.NoError(t, err)
	require.Less(t, 0, len(configJson))

	actualConfig := fullConfig{}
	err = json.Unmarshal(configJson, &actualConfig)
	require.NoError(t, err)

	assert.Equal(t, expectedFullConfig, actualConfig)
}

type TestCustomConfig struct {
	Sample string
}

func (t TestCustomConfig) UpdateFromRaw(_ interface{}) bool {
	return true
}

func TestAddSecretRequest(t *testing.T) {
	expectedRequestId := "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"

	mockProvider := &mocks.SecretProvider{}
	mockProvider.On("StoreSecret", "/mqtt", map[string]string{"password": "password", "username": "username"}).Return(nil)
	mockProvider.On("StoreSecret", "mqtt", map[string]string{"password": "password", "username": "username"}).Return(nil)
	mockProvider.On("StoreSecret", "/no", map[string]string{"password": "password", "username": "username"}).Return(errors.New("invalid w/o Vault"))

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &sdkCommon.ConfigurationStruct{}
		},
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockProvider
		},
	})

	target := NewController(nil, dic, uuid.NewString())
	assert.NotNil(t, target)

	validRequest := commonDtos.SecretRequest{
		BaseRequest: commonDtos.BaseRequest{RequestId: expectedRequestId, Versionable: commonDtos.NewVersionable()},
		Path:        "mqtt",
		SecretData: []commonDtos.SecretDataKeyValue{
			{Key: "username", Value: "username"},
			{Key: "password", Value: "password"},
		},
	}

	NoPath := validRequest
	NoPath.Path = ""
	validPathWithSlash := validRequest
	validPathWithSlash.Path = "/mqtt"
	validNoRequestId := validRequest
	validNoRequestId.RequestId = ""
	badRequestId := validRequest
	badRequestId.RequestId = "bad requestId"
	noSecrets := validRequest
	noSecrets.SecretData = []commonDtos.SecretDataKeyValue{}
	missingSecretKey := validRequest
	missingSecretKey.SecretData = []commonDtos.SecretDataKeyValue{
		{Key: "", Value: "username"},
	}
	missingSecretValue := validRequest
	missingSecretValue.SecretData = []commonDtos.SecretDataKeyValue{
		{Key: "username", Value: ""},
	}
	noSecretStore := validRequest
	noSecretStore.Path = "/no"

	tests := []struct {
		Name               string
		Request            commonDtos.SecretRequest
		ExpectedRequestId  string
		SecretsPath        string
		SecretStoreEnabled string
		ErrorExpected      bool
		ExpectedStatusCode int
	}{
		{"Valid - sub-path no trailing slash, SecretsPath has trailing slash", validRequest, expectedRequestId, "my-secrets/", "true", false, http.StatusCreated},
		{"Valid - sub-path only with trailing slash", validPathWithSlash, expectedRequestId, "my-secrets", "true", false, http.StatusCreated},
		{"Valid - both trailing slashes", validPathWithSlash, expectedRequestId, "my-secrets/", "true", false, http.StatusCreated},
		{"Valid - no requestId", validNoRequestId, "", "", "true", false, http.StatusCreated},
		{"Invalid - no path", NoPath, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - bad requestId", badRequestId, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - no secrets", noSecrets, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - missing secret key", missingSecretKey, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - missing secret value", missingSecretValue, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - No Secret Store", noSecretStore, "", "", "false", true, http.StatusInternalServerError},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			_ = os.Setenv(secret.EnvSecretStore, testCase.SecretStoreEnabled)

			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))

			req, err := http.NewRequest(http.MethodPost, internal.ApiAddSecretRoute, reader)
			require.NoError(t, err)
			req.Header.Set(common.CorrelationHeader, expectedCorrelationId)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(target.AddSecret)
			handler.ServeHTTP(recorder, req)

			actualResponse := commonDtos.BaseResponse{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion, "Api Version not as expected")
			assert.Equal(t, testCase.ExpectedStatusCode, actualResponse.StatusCode, "BaseResponse status code not as expected")

			if testCase.ErrorExpected {
				assert.NotEmpty(t, actualResponse.Message, "Message is empty")
				return // Test complete for error cases
			}

			assert.Equal(t, testCase.ExpectedRequestId, actualResponse.RequestId, "RequestID not as expected")
			assert.Empty(t, actualResponse.Message, "Message not empty, as expected")
		})
	}
}

func doRequest(t *testing.T, method string, api string, handler http.HandlerFunc, body io.Reader) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, api, body)
	require.NoError(t, err)
	req.Header.Set(common.CorrelationHeader, expectedCorrelationId)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	expectedStatusCode := http.StatusOK
	if method == http.MethodPost {
		expectedStatusCode = http.StatusMultiStatus
	}

	assert.Equal(t, expectedStatusCode, recorder.Code, "Wrong status code")
	assert.Equal(t, common.ContentTypeJSON, recorder.Result().Header.Get(common.ContentType), "Content type not set or not JSON")
	assert.Equal(t, expectedCorrelationId, recorder.Result().Header.Get(common.CorrelationHeader), "CorrelationHeader not as expected")

	require.NotEmpty(t, recorder.Body.String(), "Response body is empty")

	return recorder
}
