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
	"net/http"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	commonDtos "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/bootstrap/container"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"

	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/gorilla/mux"
)

// Controller controller for V2 REST APIs
type Controller struct {
	router         *mux.Router
	secretProvider bootstrapInterfaces.SecretProvider
	lc             logger.LoggingClient
	config         *sdkCommon.ConfigurationStruct
	customConfig   interfaces.UpdatableConfig
	serviceName    string
}

// NewController creates and initializes an Controller
func NewController(router *mux.Router, dic *di.Container, serviceName string) *Controller {
	return &Controller{
		router:         router,
		secretProvider: bootstrapContainer.SecretProviderFrom(dic.Get),
		lc:             bootstrapContainer.LoggingClientFrom(dic.Get),
		config:         container.ConfigurationFrom(dic.Get),
		serviceName:    serviceName,
	}
}

// SetCustomConfigInfo sets the custom configuration, which is used to include the service's custom config in the /config endpoint response.
func (c *Controller) SetCustomConfigInfo(customConfig interfaces.UpdatableConfig) {
	c.customConfig = customConfig
}

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *Controller) Ping(writer http.ResponseWriter, request *http.Request) {
	response := commonDtos.NewPingResponse(c.serviceName)
	c.sendResponse(writer, request, common.ApiPingRoute, response, http.StatusOK)
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *Controller) Version(writer http.ResponseWriter, request *http.Request) {
	response := commonDtos.NewVersionSdkResponse(internal.ApplicationVersion, internal.SDKVersion, c.serviceName)
	c.sendResponse(writer, request, common.ApiVersionRoute, response, http.StatusOK)
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *Controller) Config(writer http.ResponseWriter, request *http.Request) {
	var fullConfig interface{}

	if c.customConfig == nil {
		// case of no custom configs
		fullConfig = *c.config
	} else {
		// create a struct combining the common configuration and custom configuration sections
		fullConfig = struct {
			sdkCommon.ConfigurationStruct
			CustomConfiguration interfaces.UpdatableConfig
		}{
			*c.config,
			c.customConfig,
		}
	}

	response := commonDtos.NewConfigResponse(fullConfig, c.serviceName)
	c.sendResponse(writer, request, common.ApiVersionRoute, response, http.StatusOK)
}

// AddSecret handles the request to add App Service exclusive secret to the Secret Store
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *Controller) AddSecret(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		_ = request.Body.Close()
	}()

	secretRequest := commonDtos.SecretRequest{}
	err := json.NewDecoder(request.Body).Decode(&secretRequest)
	if err != nil {
		c.sendError(writer, request, errors.KindContractInvalid, "JSON decode failed", err, "")
		return
	}

	path, secret := c.prepareSecret(secretRequest)

	if err := c.secretProvider.StoreSecret(path, secret); err != nil {
		c.sendError(writer, request, errors.KindServerError, "Storing secret failed", err, secretRequest.RequestId)
		return
	}

	response := commonDtos.NewBaseResponse(secretRequest.RequestId, "", http.StatusCreated)
	c.sendResponse(writer, request, internal.ApiAddSecretRoute, response, http.StatusCreated)
}

func (c *Controller) sendError(
	writer http.ResponseWriter,
	request *http.Request,
	errKind errors.ErrKind,
	message string,
	err error,
	requestID string) {
	edgexErr := errors.NewCommonEdgeX(errKind, message, err)
	c.lc.Error(edgexErr.Error())
	c.lc.Debug(edgexErr.DebugMessages())
	response := commonDtos.NewBaseResponse(requestID, edgexErr.Message(), edgexErr.Code())
	c.sendResponse(writer, request, internal.ApiAddSecretRoute, response, edgexErr.Code())
}

// sendResponse puts together the response packet for the V2 API
func (c *Controller) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	api string,
	response interface{},
	statusCode int) {

	correlationID := request.Header.Get(common.CorrelationHeader)

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	data, err := json.Marshal(response)
	if err != nil {
		c.lc.Errorf("Unable to marshal %s response: %w, %s=%s", api, err, common.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = writer.Write(data)
	if err != nil {
		c.lc.Errorf("Unable to write %s response: %w, %s=%s", api, err, common.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Controller) prepareSecret(request commonDtos.SecretRequest) (string, map[string]string) {
	var secretsKV = make(map[string]string)
	for _, secret := range request.SecretData {
		secretsKV[secret.Key] = secret.Value
	}

	path := strings.TrimSpace(request.Path)

	return path, secretsKV
}
