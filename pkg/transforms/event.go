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

package transforms

import (
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
)

type EventWrapper struct {
	profileName  string
	deviceName   string
	resourceName string
	valueType    string
	mediaType    string
}

// NewEventWrapperSimpleReading is provided to interact with EventWrapper to add a simple reading
func NewEventWrapperSimpleReading(profileName string, deviceName string, resourceName string, valueType string) *EventWrapper {
	eventWrapper := &EventWrapper{
		profileName:  profileName,
		deviceName:   deviceName,
		resourceName: resourceName,
		valueType:    valueType,
	}
	return eventWrapper
}

// NewEventWrapperBinaryReading is provided to interact with EventWrapper to add a binary reading
func NewEventWrapperBinaryReading(profileName string, deviceName string, resourceName string, mediaType string) *EventWrapper {
	eventWrapper := &EventWrapper{
		profileName:  profileName,
		deviceName:   deviceName,
		resourceName: resourceName,
		valueType:    common.ValueTypeBinary,
		mediaType:    mediaType,
	}
	return eventWrapper
}

// NewEventWrapperObjectReading is provided to interact with EventWrapper to add a object reading type
func NewEventWrapperObjectReading(profileName string, deviceName string, resourceName string) *EventWrapper {
	eventWrapper := &EventWrapper{
		profileName:  profileName,
		deviceName:   deviceName,
		resourceName: resourceName,
		valueType:    common.ValueTypeObject,
	}
	return eventWrapper
}

// // Wrap creates an EventRequest using the Event/Reading metadata that have been set.
func (ew *EventWrapper) Wrap(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	ctx.LoggingClient().Info("Creating Event...")

	if data == nil {
		return false, fmt.Errorf("function EventWrapper in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	event := dtos.NewEvent(ew.profileName, ew.deviceName, ew.resourceName)

	switch ew.valueType {
	case common.ValueTypeBinary:
		reading, err := util.CoerceType(data)
		if err != nil {
			return false, err
		}
		event.AddBinaryReading(ew.resourceName, reading, ew.mediaType)

	case common.ValueTypeString:
		reading, err := util.CoerceType(data)
		if err != nil {
			return false, err
		}
		err = event.AddSimpleReading(ew.resourceName, ew.valueType, string(reading))
		if err != nil {
			return false, fmt.Errorf("error adding Reading in pipeline '%s': %s", ctx.PipelineId(), err.Error())
		}

	case common.ValueTypeObject:
		event.AddObjectReading(ew.resourceName, data)

	default:
		err := event.AddSimpleReading(ew.resourceName, ew.valueType, data)
		if err != nil {
			return false, fmt.Errorf("error adding Reading in pipeline '%s': %s", ctx.PipelineId(), err.Error())
		}
	}

	// unsetting content type to send back as event
	ctx.SetResponseContentType("")
	ctx.AddValue(interfaces.PROFILENAME, ew.profileName)
	ctx.AddValue(interfaces.DEVICENAME, ew.deviceName)
	ctx.AddValue(interfaces.SOURCENAME, ew.resourceName)

	// need to wrap in Add Event Request for Core Data to process it if published to the MessageBus
	eventRequest := requests.NewAddEventRequest(event)

	return true, eventRequest
}
