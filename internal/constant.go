//
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

package internal

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
)

const (
	ApiTriggerRoute   = common.ApiBase + "/trigger"
	ApiAddSecretRoute = common.ApiBase + "/secret"
)

// SDKVersion indicates the version of the SDK - will be overwritten by build
var SDKVersion = "0.0.0"

// ApplicationVersion indicates the version of the application itself, not the SDK - will be overwritten by build
var ApplicationVersion = "0.0.0"

// Names for the Common Application Service Metrics
const (
	MessagesReceivedName              = "MessagesReceived"
	InvalidMessagesReceivedName       = "InvalidMessagesReceived"
	PipelineIdTxt                     = "{PipelineId}"
	PipelineMessagesProcessedName     = "PipelineMessagesProcessed-" + PipelineIdTxt
	PipelineMessageProcessingTimeName = "PipelineMessageProcessingTime-" + PipelineIdTxt
	PipelineProcessingErrorsName      = "PipelineProcessingErrors-" + PipelineIdTxt
	HttpExportSizeName                = "HttpExportSize"
	HttpExportErrorName               = "HttpExportError"
	MqttExportSizeName                = "MqttExportSize"
	MqttExportErrorName               = "MqttExportError"
	MessageBusSubscribeTopics         = "SubscribeTopics"
	MessageBusPublishTopic            = "PublishTopic"
)

// MetricsReservoirSize is the default Metrics Sample Reservoir size
const MetricsReservoirSize = 1028
