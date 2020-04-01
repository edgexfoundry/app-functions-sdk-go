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
	"encoding/json"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/urlclient"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var context *appcontext.Context
var lc logger.LoggingClient

const (
	devID1 = "id1"
	devID2 = "id2"
)

func init() {
	lc := logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
	eventClient := coredata.NewEventClient(
		urlclient.New(nil, nil, nil, "", "", 0, "http://test"+clients.ApiEventRoute),
	)

	context = &appcontext.Context{
		LoggingClient: lc,
		EventClient:   eventClient,
	}
}
func TestTransformToXML(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`
	conv := NewConversion()

	continuePipeline, result := conv.TransformToXML(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))
}
func TestTransformToXMLNoParameters(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context)

	assert.Equal(t, "No Event Received", result.(error).Error())
	assert.False(t, continuePipeline)
}
func TestTransformToXMLNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context, "")

	assert.Equal(t, "Unexpected type received", result.(error).Error())
	assert.False(t, continuePipeline)

}
func TestTransformToXMLMultipleParametersValid(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context, eventIn, "", "", "")
	require.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))
}
func TestTransformToXMLMultipleParametersTwoEvents(t *testing.T) {
	// Event from device 1
	eventIn1 := models.Event{
		Device: devID1,
	}
	// Event from device 1
	eventIn2 := models.Event{
		Device: devID2,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id2</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context, eventIn2, eventIn1, "", "")

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

}

func TestTransformToJSON(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `{"device":"id1"}`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))
}
func TestTransformToJSONNoEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context)

	assert.Equal(t, "No Event Received", result.(error).Error())
	assert.False(t, continuePipeline)

}
func TestTransformToJSONNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, "")
	require.EqualError(t, result.(error), "Unexpected type received")
	assert.False(t, continuePipeline)

}
func TestTransformToJSONMultipleParametersValid(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `{"device":"id1"}`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, eventIn, "", "", "")
	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

}
func TestTransformToJSONMultipleParametersTwoEvents(t *testing.T) {
	// Event from device 1
	eventIn1 := models.Event{
		Device: devID1,
	}
	// Event from device 2
	eventIn2 := models.Event{
		Device: devID2,
	}
	expectedResult := `{"device":"id2"}`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, eventIn2, eventIn1, "", "")

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

}

func TestTransformToCloudEvent(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		ID:       "event-" + devID1,
		Device:   devID1,
		Readings: []models.Reading{models.Reading{Id: "123-abc", Name: "test-reading"}},
	}
	expectedJSON := `{"data":{"id":"event-id1","device":"id1","readings":[{"id":"123-abc","name":"test-reading"}]},"id":"event-id1","source":"id1","specversion":"1.0","type":"models.Event"}`
	var expectedResult cloudevents.Event
	json.Unmarshal([]byte(expectedJSON), &expectedResult)
	conv := NewConversion()
	continuePipeline, result := conv.TransformToCloudEvent(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result)
}

func TestTransformToCloudEventWrongEvent(t *testing.T) {
	eventIn := "Not a models.Event, a string"
	conv := NewConversion()
	continuePipeline, result := conv.TransformToCloudEvent(context, eventIn)

	assert.Error(t, result.(error))
	assert.False(t, continuePipeline)
}

func TestTransformToCloudEventNoReadings(t *testing.T) {
	eventIn := models.Event{
		ID:     "event-" + devID1,
		Device: devID1,
	}
	conv := NewConversion()
	continuePipeline, result := conv.TransformToCloudEvent(context, eventIn)

	assert.Error(t, result.(error))
	assert.False(t, continuePipeline)
}

func TestTransformToCloudEventNoEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToCloudEvent(context)

	assert.Error(t, result.(error))
	assert.False(t, continuePipeline)
}

func TestTransformFromCloudEvent(t *testing.T) {
	cloudeventJSON := `{"data":{"id":"event-id1","device":"id1","readings":[{"id":"123-abc","name":"test-reading", "value":"123"}]},"id":"event-id1","source":"id1","specversion":"1.0","type":"models.Event"}`
	var cloudEvent cloudevents.Event
	json.Unmarshal([]byte(cloudeventJSON), &cloudEvent)

	conv := NewConversion()
	continuePipeline, result := conv.TransformFromCloudEvent(context, cloudEvent)

	expectedEvent := models.Event{
		ID:       "event-" + devID1,
		Device:   devID1,
		Readings: []models.Reading{models.Reading{Id: "123-abc", Name: "test-reading", Value: "123"}},
	}

	edgexEvent, ok := result.(models.Event)
	assert.NotNil(t, result)
	assert.True(t, ok)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedEvent.ID, edgexEvent.ID)
	assert.Equal(t, expectedEvent.Readings[0].Id, edgexEvent.Readings[0].Id)
	assert.Equal(t, expectedEvent.Readings[0].Name, edgexEvent.Readings[0].Name)
	assert.Equal(t, expectedEvent.Readings[0].Value, edgexEvent.Readings[0].Value)
}

func TestTransformFromCloudEventEmptyCloudEvent(t *testing.T) {
	var cloudEvent cloudevents.Event
	conv := NewConversion()
	continuePipeline, result := conv.TransformFromCloudEvent(context, cloudEvent)
	_, ok := result.(error)

	assert.True(t, ok)
	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
}

func TestTransformFromCloudEventEmptyReadins(t *testing.T) {
	cloudeventJSON := `{"data":{"id":"event-id1","device":"id1","readings":[]},"id":"event-id1","source":"id1","specversion":"1.0","type":"models.Event"}`
	var cloudEvent cloudevents.Event
	json.Unmarshal([]byte(cloudeventJSON), &cloudEvent)
	conv := NewConversion()
	continuePipeline, result := conv.TransformFromCloudEvent(context, cloudEvent)
	_, ok := result.(error)

	assert.NotNil(t, result)
	assert.True(t, ok)
	assert.False(t, continuePipeline)
}
