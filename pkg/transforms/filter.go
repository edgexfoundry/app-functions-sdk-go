//
// Copyright (c) 2019 Intel Corporation
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
	"errors"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Filter houses various the parameters for which filter transforms filter on
type Filter struct {
	FilterValues []string
	FilterOut    bool
}

// NewFilter creates, initializes and returns a new instance of Filter
func NewFilter(filterValues []string) Filter {
	return Filter{FilterValues: filterValues}
}

// FilterByDeviceName filters for data coming from specific devices. It filters out those messages whose Event is
// for devices not in FilterValues. For example, data generated by a motor does not get passed to functions only
// interested in data from a thermostat. This function will return an error and stop the pipeline if a non-edgex event
// is received or if no data is received.
func (f Filter) FilterByDeviceName(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {

	edgexcontext.LoggingClient.Debug("Filtering by DeviceID")

	if len(params) < 1 {
		return false, errors.New("no Event Received")
	}

	deviceIDs := f.FilterValues
	event, ok := params[0].(models.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	// No deviceIDs to filter for, so pass events thru rather than filtering them all out.
	if len(deviceIDs) == 0 {
		return true, event
	}

	for _, devID := range deviceIDs {
		if event.Device == devID {
			if f.FilterOut {
				edgexcontext.LoggingClient.Trace(fmt.Sprintf("Event not accepted: %s", event.Device))
				return false, nil
			} else {
				edgexcontext.LoggingClient.Trace(fmt.Sprintf("Event accepted: %s", event.Device))
				return true, event
			}
		}
	}
	if f.FilterOut {
		edgexcontext.LoggingClient.Trace(fmt.Sprintf("Event accepted: %s", event.Device))
		return true, event
	}
	return false, nil

}

// FilterByValueDescriptor filters for data from certain types of IoT objects, such as temperatures, motion, and so forth.
// Reading types not in FilterValues are removed leaving just the readings that match one of the values in FilterValues.
// For example, pressure reading data does not go to functions only interested in motion data.
// This function will return an error and stop the pipeline if a non-edgex event is received or if no data is received.
func (f Filter) FilterByValueDescriptor(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {

	edgexcontext.LoggingClient.Debug("Filtering by ValueDescriptor")

	if len(params) < 1 {
		return false, errors.New("no Event Received")
	}

	existingEvent, ok := params[0].(models.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	// No filter values, so pass all event and all readings thru, rather than filtering them all out.
	if len(f.FilterValues) == 0 {
		return true, existingEvent
	}

	auxEvent := models.Event{
		Pushed:   existingEvent.Pushed,
		Device:   existingEvent.Device,
		Created:  existingEvent.Created,
		Modified: existingEvent.Modified,
		Origin:   existingEvent.Origin,
		Readings: []models.Reading{},
	}

	if f.FilterOut {
		for _, reading := range existingEvent.Readings {
			readingFilteredOut := false
			for _, filterID := range f.FilterValues {
				if reading.Name == filterID {
					edgexcontext.LoggingClient.Trace(fmt.Sprintf("Reading filtered out: %s", reading.Name))
					readingFilteredOut = true
				}
			}
			if !readingFilteredOut {
				auxEvent.Readings = append(auxEvent.Readings, reading)
			}
		}
	} else {
		for _, filterID := range f.FilterValues {
			for _, reading := range existingEvent.Readings {
				if reading.Name == filterID {
					edgexcontext.LoggingClient.Trace(fmt.Sprintf("Reading accepted: %s", reading.Name))
					auxEvent.Readings = append(auxEvent.Readings, reading)
				}
			}
		}
	}
	thereExistReadings := len(auxEvent.Readings) > 0
	var returnResult models.Event
	if thereExistReadings {
		returnResult = auxEvent
	}
	return thereExistReadings, returnResult
}
