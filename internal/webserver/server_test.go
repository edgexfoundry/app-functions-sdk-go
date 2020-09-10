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

package webserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/security"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/telemetry"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var logClient logger.LoggingClient
var config *common.ConfigurationStruct

func TestMain(m *testing.M) {
	logClient = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
	config = &common.ConfigurationStruct{}
	m.Run()
}

func TestAddRoute(t *testing.T) {
	routePath := "/testRoute"
	testHandler := func(_ http.ResponseWriter, _ *http.Request) {}
	sp := security.NewSecretProvider(logClient, config)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	err := webserver.AddRoute(routePath, testHandler)
	assert.NoError(t, err, "Not expecting an error")

	// Malformed path no slash
	routePath = "testRoute"
	err = webserver.AddRoute(routePath, testHandler)
	assert.Error(t, err, "Expecting an error")
}

func TestEncode(t *testing.T) {
	sp := security.NewSecretProvider(logClient, config)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	writer := httptest.NewRecorder()
	var junkData interface{}
	// something that will always fail to marshal
	junkData = math.Inf(1)
	webserver.encode(junkData, writer)
	body := writer.Body.String()
	assert.NotEqual(t, math.Inf(1), body)
}

func TestConfigureAndPingRoute(t *testing.T) {

	sp := security.NewSecretProvider(logClient, config)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest(http.MethodGet, clients.ApiPingRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()
	assert.Equal(t, "pong", body)

}

func TestConfigureAndVersionRoute(t *testing.T) {

	sp := security.NewSecretProvider(logClient, config)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest(http.MethodGet, clients.ApiVersionRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()
	assert.Equal(t, "{\"version\":\"0.0.0\",\"sdk_version\":\"0.0.0\"}\n", body)

}
func TestConfigureAndConfigRoute(t *testing.T) {

	sp := security.NewSecretProvider(logClient, config)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest(http.MethodGet, clients.ApiConfigRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	expected := `{"Writable":{"LogLevel":"","Pipeline":{"ExecutionOrder":"","UseTargetTypeOfByteArray":false,"Functions":null},"StoreAndForward":{"Enabled":false,"RetryInterval":"","MaxRetryCount":0},"InsecureSecrets":null},"Logging":{"EnableRemote":false,"File":""},"Registry":{"Host":"","Port":0,"Type":""},"Service":{"BootTimeout":"","CheckInterval":"","Host":"","HTTPSCert":"","HTTPSKey":"","ServerBindAddr":"","Port":0,"Protocol":"","StartupMsg":"","ReadMaxLimit":0,"Timeout":""},"MessageBus":{"PublishHost":{"Host":"","Port":0,"Protocol":""},"SubscribeHost":{"Host":"","Port":0,"Protocol":""},"Type":"","Optional":null},"Binding":{"Type":"","SubscribeTopic":"","PublishTopic":""},"ApplicationSettings":null,"Clients":null,"Database":{"Type":"","Host":"","Port":0,"Timeout":"","Username":"","Password":"","MaxIdle":0,"BatchSize":0},"SecretStore":{"Host":"","Port":0,"Path":"","Protocol":"","Namespace":"","RootCaCertPath":"","ServerName":"","Authentication":{"AuthType":"","AuthToken":""},"AdditionalRetryAttempts":0,"RetryWaitPeriod":"","TokenFile":""},"SecretStoreExclusive":{"Host":"","Port":0,"Path":"","Protocol":"","Namespace":"","RootCaCertPath":"","ServerName":"","Authentication":{"AuthType":"","AuthToken":""},"AdditionalRetryAttempts":0,"RetryWaitPeriod":"","TokenFile":""}}` + "\n"

	body := rr.Body.String()
	assert.Equal(t, expected, body)
}

func TestConfigureAndMetricsRoute(t *testing.T) {
	sp := newsecretProviderMock(config)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest(http.MethodGet, clients.ApiMetricsRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()
	metrics := telemetry.SystemUsage{}
	json.Unmarshal([]byte(body), &metrics)
	assert.NotNil(t, body, "Metrics not populated")
	assert.NotZero(t, metrics.Memory.Alloc, "Expected Alloc value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.Frees, "Expected Frees value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.LiveObjects, "Expected LiveObjects value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.Mallocs, "Expected Mallocs value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.Sys, "Expected Sys value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.TotalAlloc, "Expected TotalAlloc value of metrics to be non-zero")
	assert.NotNil(t, metrics.CpuBusyAvg, "Expected CpuBusyAvg value of metrics to be not nil")
}

func TestSetupTriggerRoute(t *testing.T) {
	sp := newsecretProviderMock(config)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())

	handlerFunctionNotCalled := true
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
		handlerFunctionNotCalled = false
	}

	webserver.SetupTriggerRoute(internal.ApiTriggerRoute, handler)

	req, _ := http.NewRequest(http.MethodGet, internal.ApiTriggerRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()

	assert.Equal(t, "test", body)
	assert.False(t, handlerFunctionNotCalled, "expected handler function to be called")
}

func TestPostSecretRoute(t *testing.T) {

	sp := newsecretProviderMock(config)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	tests := []struct {
		name           string
		payload        []byte
		expectedStatus int
	}{
		{
			name:           "PostSecretRoute: Good case with one secret",
			payload:        []byte(`{"path":"MyPath","secrets":[{"key":"MySecretKey","value":"MySecretValue"}]}`),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PostSecretRoute: Good case with two secrets",
			payload:        []byte(`{"path":"MyPath","secrets":[{"key":"MySecretKey1","value":"MySecretValue1"}, {"key":"MySecretKey2","value":"MySecretValue2"}]}`),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PostSecretRoute: missing secrets",
			payload:        []byte(`{"path":"MyPath"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PostSecretRoute: malformed payload",
			payload:        []byte(`<"path"="MyPath">`),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, internal.ApiSecretsRoute, bytes.NewReader(currentTest.payload))
			rr := httptest.NewRecorder()
			webserver.router.ServeHTTP(rr, req)
			assert.Equal(t, currentTest.expectedStatus, rr.Result().StatusCode, "Expected secret doesn't match postSecret")
		})
	}
}

func TestValidateSecretRoute(t *testing.T) {
	secretDataBadPath := SecretData{Path: "/!$%&/foo", Secrets: []KeyValue{KeyValue{Key: "key", Value: "val"}}}
	assert.Error(t, secretDataBadPath.validateSecretData())

	secretDataEmptyKey := SecretData{Path: "/foo/bar", Secrets: []KeyValue{KeyValue{Key: "", Value: "val"}}}
	assert.Error(t, secretDataEmptyKey.validateSecretData())

	secretDataGoodPath := SecretData{Path: "/foo/bar", Secrets: []KeyValue{KeyValue{Key: "key", Value: "val"}}}
	assert.NoError(t, secretDataGoodPath.validateSecretData())
}

type secretProviderMock struct {
	config          *common.ConfigurationStruct
	mockSecretStore map[string]map[string]string // secret's path, key, value

	//used to track when secrets have last been retrieved
	secretsLastUpdated time.Time
}

// newsecretProviderMock returns a new mock secret provider
func newsecretProviderMock(config *common.ConfigurationStruct) *secretProviderMock {
	sp := &secretProviderMock{}
	sp.config = config
	sp.mockSecretStore = make(map[string]map[string]string)
	return sp
}

// Initialize does nothing.
func (s *secretProviderMock) Initialize(_ context.Context) bool {
	return true
}

// StoreSecrets saves secrets to the mock secret store.
func (s *secretProviderMock) StoreSecrets(path string, secrets map[string]string) error {
	testFullPath := s.config.SecretStoreExclusive.Path + path
	// Base path should not have any leading slashes, only trailing or none, for this test to work
	if strings.Contains(testFullPath, "//") || !strings.Contains(testFullPath, "/") {
		return fmt.Errorf("Path is malformed: path=%s", path)
	}

	if !s.isSecurityEnabled() {
		return fmt.Errorf("Storing secrets is not supported when running in insecure mode")
	}
	s.mockSecretStore[path] = secrets
	return nil
}

// GetSecrets retrieves secrets from a mock secret store.
func (s *secretProviderMock) GetSecrets(path string, _ ...string) (map[string]string, error) {
	secrets, ok := s.mockSecretStore[path]
	if !ok {
		return nil, fmt.Errorf("no secrets for path '%s' found", path)
	}
	return secrets, nil
}

// GetDatabaseCredentials retrieves the login credentials for the database from mock secret store
func (s *secretProviderMock) GetDatabaseCredentials(database db.DatabaseInfo) (common.Credentials, error) {
	credentials, ok := s.mockSecretStore[database.Type]
	if !ok {
		return common.Credentials{}, fmt.Errorf("no credentials for type '%s' found", database.Type)
	}

	return common.Credentials{
		Username: credentials["username"],
		Password: credentials["password"],
	}, nil
}

// InsecureSecretsUpdated resets LastUpdate is not running in secure mode.If running in secure mode, changes to
// InsecureSecrets have no impact and are not used.
func (s *secretProviderMock) InsecureSecretsUpdated() {
	s.secretsLastUpdated = time.Now()
}

// SecretsLastUpdated returns the time stamp when the provider secrets cache was latest updated
func (s *secretProviderMock) SecretsLastUpdated() time.Time {
	return s.secretsLastUpdated
}

// isSecurityEnabled determines if security has been enabled.
func (s *secretProviderMock) isSecurityEnabled() bool {
	env := os.Getenv(security.EnvSecretStore)
	return env != "false"
}
