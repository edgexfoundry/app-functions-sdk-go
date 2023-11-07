//
// Copyright (c) 2021 One Track Consulting
// Copyright (c) 2022 Intel Corporation
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

package trigger

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
)

type ServiceBinding interface {
	// DecodeMessage decodes the message received in the envelope and returns the data to be processed
	DecodeMessage(appContext *appfunction.Context, envelope types.MessageEnvelope) (interface{}, *runtime.MessageError, bool)
	// ProcessMessage provides access to the runtime's ProcessMessage function to process the decoded data
	ProcessMessage(appContext *appfunction.Context, data interface{}, pipeline *interfaces.FunctionPipeline) *runtime.MessageError
	// GetMatchingPipelines provides access to the runtime's GetMatchingPipelines function
	GetMatchingPipelines(incomingTopic string) []*interfaces.FunctionPipeline
	// GetDefaultPipeline provides access to the runtime's GetDefaultPipeline function
	GetDefaultPipeline() *interfaces.FunctionPipeline
	// BuildContext creates a context for a given message envelope
	BuildContext(env types.MessageEnvelope) interfaces.AppFunctionContext
	// SecretProvider provides access to this service's secret provider for the trigger
	SecretProvider() messaging.SecretDataProvider
	// Config provides access to this service's configuration for the trigger
	Config() *common.ConfigurationStruct
	// LoggingClient provides access to this service's logging clietn for the trigger
	LoggingClient() logger.LoggingClient
	// LoadCustomConfig provides access to the service's LoadCustomConfig function
	LoadCustomConfig(config interfaces.UpdatableConfig, sectionName string) error
}
