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

package webserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"
)

var dic *di.Container

func TestMain(m *testing.M) {
	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &common.ConfigurationStruct{}
		},
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return &mocks.SecretProvider{}
		},
	})

	os.Exit(m.Run())
}

func TestAddRoute(t *testing.T) {
	routePath := "/testRoute"
	testHandler := func(c echo.Context) error { return nil }

	webserver := NewWebServer(dic, echo.New(), uuid.NewString())
	webserver.AddRoute(routePath, testHandler, []string{http.MethodGet})

	// Malformed path no slash
	routePath = "testRoute"
	webserver.AddRoute(routePath, testHandler, nil)
}

func TestSetupTriggerRoute(t *testing.T) {
	envDisableSecurity := os.Getenv("EDGEX_DISABLE_JWT_VALIDATION")
	_ = os.Setenv("EDGEX_DISABLE_JWT_VALIDATION", "true")
	defer func() {
		_ = os.Setenv("EDGEX_DISABLE_JWT_VALIDATION", envDisableSecurity)
	}()

	webserver := NewWebServer(dic, echo.New(), uuid.NewString())

	handlerFunctionNotCalled := true
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("test"))
		require.NoError(t, err)
		handlerFunctionNotCalled = false
	}

	webserver.SetupTriggerRoute(internal.ApiTriggerRoute, handler)

	req, _ := http.NewRequest(http.MethodPost, internal.ApiTriggerRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()

	assert.Equal(t, "test", body)
	assert.False(t, handlerFunctionNotCalled, "expected handler function to be called")
}
