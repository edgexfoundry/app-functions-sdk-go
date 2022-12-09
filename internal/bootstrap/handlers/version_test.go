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

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/bootstrap/container"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-registry/v3/registry"
	"github.com/stretchr/testify/assert"
)

func TestValidateVersionMatch(t *testing.T) {
	startupTimer := startup.NewStartUpTimer("unit-test")

	clientConfigs := make(map[string]config.ClientInfo)
	clientConfigs[common.CoreMetaDataServiceKey] = config.ClientInfo{
		Protocol: "http",
		Host:     "localhost",
		Port:     0, // Will be replaced by local test webserver's port
	}

	configuration := &sdkCommon.ConfigurationStruct{
		Writable: sdkCommon.WritableInfo{
			LogLevel: "DEBUG",
		},
		Clients: clientConfigs,
	}

	lc := logger.NewMockClient()
	var registryClient registry.Client = nil

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		bootstrapContainer.RegistryClientInterfaceName: func(get di.Get) interface{} {
			return registryClient
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	tests := []struct {
		Name             string
		CoreVersion      string
		SdkVersion       string
		skipVersionCheck bool
		ExpectFailure    bool
	}{
		{"Compatible Versions", "1.1.0", "v1.0.0", false, false},
		{"SDK Dev Compatible Versions", "2.0.0", "v2.0.0-dev.11", false, false},
		{"Core Dev Compatible Versions", "1.2.1-dev.1", "v1.2.0", false, false},
		{"Both Dev Compatible Versions", "1.2.1-dev.1", "v1.2.0-dev.4", false, false},
		{"Un-compatible Versions", "2.0.0", "v1.0.0", false, true},
		{"Skip Version Check", "2.0.0", "v1.0.0", true, false},
		{"Running in Debugger", "1.0.0", "v0.0.0", false, false},
		{"SDK Beta Version", "1.0.0", "v0.2.0", false, false},
		{"SDK Version malformed", "1.0.0", "", false, true},
		{"Core prerelease version", CorePreReleaseVersion, "v1.0.0", false, false},
		{"Core developer version", CoreDeveloperVersion, "v1.0.0", false, false},
		{"Core version malformed", "12", "v1.0.0", false, true},
		{"Core version JSON bad", "", "v1.0.0", false, true},
		{"Core version JSON empty", "{}", "v1.0.0", false, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			handler := func(w http.ResponseWriter, r *http.Request) {
				var versionJson string
				if test.CoreVersion == "{}" {
					versionJson = "{}"
				} else if test.CoreVersion == "" {
					versionJson = ""
				} else {
					versionJson = fmt.Sprintf(`{"version" : "%s"}`, test.CoreVersion)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(versionJson))
			}

			// create test server with handler
			testServer := httptest.NewServer(http.HandlerFunc(handler))
			defer testServer.Close()

			testServerUrl, _ := url.Parse(testServer.URL)
			port, _ := strconv.Atoi(testServerUrl.Port())
			coreService := configuration.Clients[common.CoreMetaDataServiceKey]
			coreService.Port = port
			configuration.Clients[common.CoreMetaDataServiceKey] = coreService

			validator := NewVersionValidator(test.skipVersionCheck, test.SdkVersion)
			result := validator.BootstrapHandler(context.Background(), &sync.WaitGroup{}, startupTimer, dic)
			assert.Equal(t, test.ExpectFailure, !result)
		})
	}
}
