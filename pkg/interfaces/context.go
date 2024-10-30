//
// Copyright (c) 2023 Intel Corporation
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

package interfaces

import (
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
)

const (
	DEVICENAME    = "devicename"
	PROFILENAME   = "profilename"
	SOURCENAME    = "sourcename"
	RECEIVEDTOPIC = "receivedtopic"
	PIPELINEID    = "pipelineid"
)

// AppFunction is a type alias for a application pipeline function.
// appCtx is a reference to the AppFunctionContext below.
// data is the data to be operated on by the function.
// bool return value indicates if the pipeline should continue executing (true) or not (false)
// interface{} is either the data to pass to the next function (continue executing) or
// an error (stop executing due to error) or nil (done executing)
type AppFunction = func(appCxt AppFunctionContext, data interface{}) (bool, interface{})

// AppFunctionContext defines the interface for an Edgex Application Service Context provided to
// App Functions when executing in the Functions Pipeline.
type AppFunctionContext interface {
	// Clone returns a copy of the context that can be manipulated independently.
	Clone() AppFunctionContext
	// CorrelationID returns the correlation ID associated with the context.
	CorrelationID() string
	// InputContentType returns the content type of the data that initiated the pipeline execution. Only useful when
	// the TargetType for the pipeline is []byte, otherwise the data with be the type specified by TargetType.
	InputContentType() string
	// SetResponseData sets the response data that will be returned to the trigger when pipeline execution is complete.
	SetResponseData(data []byte)
	// ResponseData returns the data that will be returned to the trigger when pipeline execution is complete.
	ResponseData() []byte
	// SetResponseContentType sets the content type that will be returned to the trigger when pipeline
	// execution is complete.
	SetResponseContentType(string)
	// ResponseContentType returns the content type that will be returned to the trigger when pipeline
	// execution is complete.
	ResponseContentType() string
	// SetRetryData set the data that is to be retried later as part of the Store and Forward capability.
	// Used when there was failure sending the data to an external source.
	SetRetryData(data []byte)
	// TriggerRetryFailedData sets the flag to trigger the retry of any stored failed export data.
	TriggerRetryFailedData()
	// SecretProvider returns the SecretProvider instance
	SecretProvider() bootstrapInterfaces.SecretProvider
	// LoggingClient returns the Logger client
	LoggingClient() logger.LoggingClient
	// EventClient returns the Event client. Note if Core Data is not specified in the Clients configuration,
	// this will return nil.
	EventClient() interfaces.EventClient
	// ReadingClient returns the Reading client. Note if Core Data is not specified in the Clients configuration,
	// this will return nil.
	ReadingClient() interfaces.ReadingClient
	// CommandClient returns the Command client. Note if Core Command is not specified in the Clients configuration,
	// this will return nil.
	CommandClient() interfaces.CommandClient
	// NotificationClient returns the Notification client. Note if Support Notifications is not specified in the
	// Clients configuration, this will return nil.
	NotificationClient() interfaces.NotificationClient
	// SubscriptionClient returns the Subscription client. Note if Support Notifications is not specified in the
	// Clients configuration, this will return nil.
	SubscriptionClient() interfaces.SubscriptionClient
	// DeviceServiceClient returns the DeviceService client. Note if Core Metadata is not specified in the
	// Clients configuration, this will return nil.
	DeviceServiceClient() interfaces.DeviceServiceClient
	// DeviceProfileClient returns the DeviceProfile client. Note if Core Metadata is not specified in the
	// Clients configuration, this will return nil.
	DeviceProfileClient() interfaces.DeviceProfileClient
	// DeviceClient returns the Device client. Note if Core Metadata is not specified in the
	// Clients configuration, this will return nil.
	DeviceClient() interfaces.DeviceClient
	// MetricsManager returns the Metrics Manager used to register counter, gauge, gaugeFloat64 or timer metric types from
	// github.com/rcrowley/go-metrics
	MetricsManager() bootstrapInterfaces.MetricsManager
	// GetDeviceResource retrieves the DeviceResource for given profileName and resourceName.
	// Resources retrieved are cached so multiple calls for same profileName and resourceName don't result in multiple
	// unneeded HTTP calls to Core Metadata
	GetDeviceResource(profileName string, resourceName string) (dtos.DeviceResource, error)
	// AddValue stores a value for access within other functions in pipeline
	AddValue(key string, value string)
	// RemoveValue deletes a value stored in the context at the given key
	RemoveValue(key string)
	// GetValue attempts to retrieve a value stored in the context at the given key
	GetValue(key string) (string, bool)
	// GetAllValues returns a read-only copy of all data stored in the context
	GetAllValues() map[string]string
	// ApplyValues looks in the provided string for placeholders of the form
	// '{any-value-key}' and attempts to replace with the value stored under
	// the key in context storage.  An error will be returned if any placeholders
	// are not matched to a value in the context.
	ApplyValues(format string) (string, error)
	// PipelineId returns the ID of the pipeline that is executing
	PipelineId() string
	// Publish pushes data to the MessageBus using configured topic
	Publish(data any, contentType string) error
	// PublishWithTopic pushes data to the MessageBus using given topic
	PublishWithTopic(topic string, data any, contentType string) error
}
