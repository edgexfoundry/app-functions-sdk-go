// Copyright (c) 2022 Intel Corporation
// Copyright (c) 2021 One Track Consulting
// Copyright (C) 2023 IOTech Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/google/uuid"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/transforms"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceKey = "AppService-UnitTest"
)

var testAddEventRequest = createAddEventRequest()
var testEvent = testAddEventRequest.Event

func createAddEventRequest() requests.AddEventRequest {
	event := dtos.NewEvent("Thermostat", "FamilyRoomThermostat", "Temperature")
	_ = event.AddSimpleReading("Temperature", common.ValueTypeInt64, int64(72))
	request := requests.NewAddEventRequest(event)
	return request
}

func TestProcessMessageBusRequest(t *testing.T) {
	expected := http.StatusBadRequest

	badRequest := testAddEventRequest
	badRequest.Event.ProfileName = ""
	badRequest.Event.DeviceName = ""
	payload, err := json.Marshal(badRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
	}
	context := appfunction.NewContext("testId", dic, "")

	dummyTransform := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return true, "Hello"
	}

	runtime := NewFunctionPipelineRuntime("", nil, dic)
	runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{dummyTransform})

	_, decodeErr, _ := runtime.DecodeMessage(context, envelope)
	require.NotNil(t, decodeErr)

	assert.Equal(t, expected, decodeErr.ErrorCode)
}

func TestProcessMessageNoTransforms(t *testing.T) {
	expected := http.StatusInternalServerError

	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
	}
	context := appfunction.NewContext("testId", dic, "")

	runtime := NewFunctionPipelineRuntime("", nil, dic)

	messageData, decodeError, _ := runtime.DecodeMessage(context, envelope)
	require.Nil(t, decodeError)
	require.NotNil(t, messageData)

	result := runtime.ProcessMessage(context, messageData, runtime.GetDefaultPipeline())
	require.NotNil(t, result)
	assert.Equal(t, expected, result.ErrorCode)
}

func TestProcessMessageOneCustomTransform(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
		ReceivedTopic: uuid.NewString(),
	}
	context := appfunction.NewContext("testId", dic, "")

	transform1WasCalled := false
	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		require.NotNil(t, data, "should have been passed the first event from CoreData")
		if result, ok := data.(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			require.Equal(t, testEvent.DeviceName, result.DeviceName, "Did not receive expected EdgeX event")
		}
		transform1WasCalled = true
		return true, "Hello"
	}
	runtime := NewFunctionPipelineRuntime("", nil, dic)
	runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transform1})

	messageData, decodeError, _ := runtime.DecodeMessage(context, envelope)
	require.Nil(t, decodeError)
	require.NotNil(t, messageData)

	result := runtime.ProcessMessage(context, messageData, runtime.GetDefaultPipeline())
	require.Nil(t, result)
	require.True(t, transform1WasCalled, "transform1 should have been called")

	assertEventMetadataSet(t, context, envelope)
}

func TestProcessMessageTwoCustomTransforms(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
		ReceivedTopic: uuid.NewString(),
	}
	context := appfunction.NewContext("testId", dic, "")
	transform1WasCalled := false
	transform2WasCalled := false

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform1WasCalled = true
		require.NotNil(t, data, "should have been passed the first event from CoreData")
		if result, ok := data.(dtos.Event); ok {
			require.True(t, ok, "Should have received Event")
			assert.Equal(t, testEvent.DeviceName, result.DeviceName, "Did not receive expected Event")
		}

		return true, "Transform1Result"
	}
	transform2 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform2WasCalled = true

		require.Equal(t, "Transform1Result", data, "Did not receive result from previous transform")

		return true, "Hello"
	}
	runtime := NewFunctionPipelineRuntime("", nil, dic)
	runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transform1, transform2})

	messageData, decodeError, _ := runtime.DecodeMessage(context, envelope)
	require.Nil(t, decodeError)
	require.NotNil(t, messageData)

	result := runtime.ProcessMessage(context, messageData, runtime.GetDefaultPipeline())
	require.Nil(t, result)
	assert.True(t, transform1WasCalled, "transform1 should have been called")
	assert.True(t, transform2WasCalled, "transform2 should have been called")

	assertEventMetadataSet(t, context, envelope)
}

func TestProcessMessageThreeCustomTransformsOneFail(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
		ReceivedTopic: uuid.NewString(),
	}
	context := appfunction.NewContext("testId", dic, "")

	transform1WasCalled := false
	transform2WasCalled := false
	transform3WasCalled := false

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform1WasCalled = true
		require.NotNil(t, data, "should have been passed the first event from CoreData")

		if result, ok := data.(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			require.Equal(t, testEvent.DeviceName, result.DeviceName, "Did not receive expected EdgeX event")
		}

		return false, "Transform1Result"
	}
	transform2 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform2WasCalled = true
		require.Equal(t, "Transform1Result", data, "Did not receive result from previous transform")
		return true, "Hello"
	}
	transform3 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform3WasCalled = true
		require.Equal(t, "Transform1Result", data, "Did not receive result from previous transform")
		return true, "Hello"
	}
	runtime := NewFunctionPipelineRuntime("", nil, dic)
	runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transform1, transform2, transform3})

	messageData, decodeError, _ := runtime.DecodeMessage(context, envelope)
	require.Nil(t, decodeError)
	require.NotNil(t, messageData)

	result := runtime.ProcessMessage(context, messageData, runtime.GetDefaultPipeline())
	require.Nil(t, result)
	assert.True(t, transform1WasCalled, "transform1 should have been called")
	assert.False(t, transform2WasCalled, "transform2 should NOT have been called")
	assert.False(t, transform3WasCalled, "transform3 should NOT have been called")

	assertEventMetadataSet(t, context, envelope)
}

func TestProcessMessageTransformError(t *testing.T) {
	// Error expected from FilterByDeviceName
	expectedError := "FilterByDeviceName: type received is not an Event"
	expectedErrorCode := http.StatusUnprocessableEntity

	// Send a RegistryInfo to the pipeline, instead of an Event
	registryInfo := config.RegistryInfo{
		Host: testEvent.DeviceName,
	}
	payload, _ := json.Marshal(registryInfo)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
		ReceivedTopic: uuid.NewString(),
	}
	context := appfunction.NewContext("testId", dic, "")

	// Let the Runtime know we are sending a RegistryInfo, so it passes it to the first function
	runtime := NewFunctionPipelineRuntime("", &config.RegistryInfo{}, dic)
	// FilterByDeviceName with return an error if it doesn't receive and Event
	runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transforms.NewFilterFor([]string{"SomeDevice"}).FilterByDeviceName})

	messageData, decodeError, _ := runtime.DecodeMessage(context, envelope)
	require.Nil(t, decodeError)
	require.NotNil(t, messageData)

	msgErr := runtime.ProcessMessage(context, messageData, runtime.GetDefaultPipeline())

	require.Error(t, msgErr, "Expected an error")
	assert.Contains(t, msgErr.Error(), expectedError)
	assert.Equal(t, expectedErrorCode, msgErr.ErrorCode)

	assertReceivedTopicSet(t, context, envelope)
}

func assertEventMetadataSet(t *testing.T, context *appfunction.Context, envelope types.MessageEnvelope) {
	assertReceivedTopicSet(t, context, envelope)

	v, f := context.GetValue(interfaces.DEVICENAME)
	require.True(t, f)
	assert.Equal(t, testAddEventRequest.Event.DeviceName, v)

	v, f = context.GetValue(interfaces.PROFILENAME)
	require.True(t, f)
	assert.Equal(t, testAddEventRequest.Event.ProfileName, v)

	v, f = context.GetValue(interfaces.SOURCENAME)
	require.True(t, f)
	assert.Equal(t, testAddEventRequest.Event.SourceName, v)
}

func assertReceivedTopicSet(t *testing.T, context *appfunction.Context, envelope types.MessageEnvelope) {
	v, f := context.GetValue(interfaces.RECEIVEDTOPIC)
	require.True(t, f)
	assert.Equal(t, envelope.ReceivedTopic, v)
}

func TestProcessMessageJSON(t *testing.T) {
	expectedCorrelationID := "123-234-345-456"

	transform1WasCalled := false

	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
	}

	context := appfunction.NewContext("testing", dic, "")

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform1WasCalled = true

		require.Equal(t, expectedCorrelationID, appContext.CorrelationID(), "Context doesn't contain expected CorrelationID")

		if result, ok := data.(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			assert.Equal(t, testEvent.DeviceName, result.DeviceName, "Did not receive expected EdgeX event, wrong device")
			assert.Equal(t, testEvent.Id, result.Id, "Did not receive expected EdgeX event, wrong ID")
		}

		return false, nil
	}

	runtime := NewFunctionPipelineRuntime("", nil, dic)
	runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transform1})

	messageData, decodeError, _ := runtime.DecodeMessage(context, envelope)
	require.Nil(t, decodeError)
	require.NotNil(t, messageData)

	result := runtime.ProcessMessage(context, messageData, runtime.GetDefaultPipeline())
	assert.Nilf(t, result, "result should be null. Got %v", result)
	assert.True(t, transform1WasCalled, "transform1 should have been called")
}

func TestProcessMessageCBOR(t *testing.T) {
	expectedCorrelationID := "123-234-345-456"

	transform1WasCalled := false

	payload, err := cbor.Marshal(testAddEventRequest)
	assert.NoError(t, err, "expected no error when marshalling data")

	envelope := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   common.ContentTypeCBOR,
	}

	context := appfunction.NewContext("testing", dic, "")

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform1WasCalled = true

		require.Equal(t, expectedCorrelationID, appContext.CorrelationID(), "Context doesn't contain expected CorrelationID")

		if result, ok := data.(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			assert.Equal(t, testEvent.DeviceName, result.DeviceName, "Did not receive expected EdgeX event, wrong device")
			assert.Equal(t, testEvent.Id, result.Id, "Did not receive expected EdgeX event, wrong ID")
		}

		return false, nil
	}

	runtime := NewFunctionPipelineRuntime("", nil, dic)
	runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transform1})

	messageData, decodeError, _ := runtime.DecodeMessage(context, envelope)
	require.Nil(t, decodeError)
	require.NotNil(t, messageData)

	result := runtime.ProcessMessage(context, messageData, runtime.GetDefaultPipeline())
	assert.Nil(t, result, "result should be null")
	assert.True(t, transform1WasCalled, "transform1 should have been called")
}

type CustomType struct {
	ID string `json:"id"`
}

// Must implement the Marshaller interface so SetResponseData will marshal it to JSON
func (custom CustomType) MarshalJSON() ([]byte, error) {
	test := struct {
		ID string `json:"id"`
	}{
		ID: custom.ID,
	}

	return json.Marshal(test)
}

func TestDecode_Process_MessageTargetType(t *testing.T) {
	jsonPayload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	eventJsonPayload, err := json.Marshal(testEvent)
	require.NoError(t, err)

	cborPayload, err := cbor.Marshal(testAddEventRequest)
	assert.NoError(t, err)

	eventCborPayload, err := cbor.Marshal(testEvent)
	require.NoError(t, err)

	expected := CustomType{
		ID: "Id1",
	}
	customJsonPayload, _ := expected.MarshalJSON()
	byteData := []byte("This is my bytes")

	targetTypeTests := []struct {
		Name               string
		TargetType         interface{}
		Payload            []byte
		ContentType        string
		ExpectedOutputData []byte
		ErrorExpected      bool
	}{
		{"JSON default Target Type", nil, jsonPayload, common.ContentTypeJSON, eventJsonPayload, false},
		{"CBOR default Target Type", nil, cborPayload, common.ContentTypeCBOR, eventJsonPayload, false},
		{"JSON Event Event DTO", &dtos.Event{}, eventJsonPayload, common.ContentTypeJSON, eventJsonPayload, false},
		{"CBOR Event Event DTO", &dtos.Event{}, eventCborPayload, common.ContentTypeCBOR, eventJsonPayload, false}, // Not re-encoding as CBOR
		{"Custom Type Json", &CustomType{}, customJsonPayload, common.ContentTypeJSON, customJsonPayload, false},
		{"Byte Slice", &[]byte{}, byteData, "application/binary", byteData, false},
		{"Target Type Not a pointer", dtos.Event{}, nil, "", nil, true},
	}

	for _, currentTest := range targetTypeTests {
		t.Run(currentTest.Name, func(t *testing.T) {
			envelope := types.MessageEnvelope{
				CorrelationID: "123-234-345-456",
				Payload:       currentTest.Payload,
				ContentType:   currentTest.ContentType,
			}

			context := appfunction.NewContext("testing", dic, "")

			runtime := NewFunctionPipelineRuntime("", currentTest.TargetType, dic)
			runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transforms.NewResponseData().SetResponseData})

			targetData, err, _ := runtime.DecodeMessage(context, envelope)
			if currentTest.ErrorExpected {
				assert.Error(t, err, fmt.Sprintf("expected an error for test '%s'", currentTest.Name))
				return
			} else {
				require.Nil(t, err, fmt.Sprintf("unexpected error for test '%s'", currentTest.Name))
				require.NotNil(t, targetData)
			}

			err = runtime.ProcessMessage(context, targetData, runtime.GetDefaultPipeline())
			require.Nil(t, err)

			// ResponseData will be nil if an error occurred in the pipeline processing the data
			assert.Equal(t, currentTest.ExpectedOutputData, context.ResponseData(), fmt.Sprintf("'%s' test failed", currentTest.Name))

			switch currentTest.TargetType.(type) {
			case nil:
				assertEventMetadataSet(t, context, envelope)
			case *dtos.Event:
				assertEventMetadataSet(t, context, envelope)
			default:
				assertReceivedTopicSet(t, context, envelope)
			}
		})
	}
}

func TestExecutePipelinePersist(t *testing.T) {
	expectedItemCount := 1

	context := appfunction.NewContext("testing", dic, "")
	transformPassThru := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return true, data
	}

	runtime := NewFunctionPipelineRuntime(serviceKey, nil, updateDicWithMockStoreClient())

	httpPost := transforms.NewHTTPSender("http://nowhere", "", true).HTTPPost
	runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transformPassThru, httpPost})

	payload := []byte("My Payload")

	pipeline := runtime.GetDefaultPipeline()
	// Target of this test
	actual := runtime.ExecutePipeline(payload, context, pipeline, 0, false)

	require.NotNil(t, actual)
	require.Error(t, actual.Err, "Error expected from export function")
	storedObjects := mockRetrieveObjects(serviceKey)
	require.Equal(t, expectedItemCount, len(storedObjects), "unexpected item count")
	assert.Equal(t, serviceKey, storedObjects[0].AppServiceKey, "AppServiceKey not as expected")
	assert.Equal(t, context.CorrelationID(), storedObjects[0].CorrelationID, "CorrelationID not as expected")
}

func TestGolangRuntime_processEventPayload(t *testing.T) {
	jsonAddEventPayload, _ := json.Marshal(testAddEventRequest)
	cborAddEventPayload, _ := cbor.Marshal(testAddEventRequest)
	jsonEventPayload, _ := json.Marshal(testAddEventRequest.Event)
	cborEventPayload, _ := cbor.Marshal(testAddEventRequest.Event)

	notAnEvent := dtos.DeviceResource{
		Description: "Not An Event",
		Name:        "SomeResource",
	}
	jsonInvalidPayload, _ := json.Marshal(notAnEvent)
	cborInvalidPayload, _ := cbor.Marshal(notAnEvent)

	expectedEvent := testEvent

	tests := []struct {
		Name        string
		Payload     []byte
		ContentType string
		Expected    *dtos.Event
		ExpectError bool
	}{
		{"JSON Add Event DTO", jsonAddEventPayload, common.ContentTypeJSON, &expectedEvent, false},
		{"CBOR Add Event DTO", cborAddEventPayload, common.ContentTypeCBOR, &expectedEvent, false},
		{"JSON Event DTO", jsonEventPayload, common.ContentTypeJSON, &expectedEvent, false},
		{"CBOR Event DTO", cborEventPayload, common.ContentTypeCBOR, &expectedEvent, false},
		{"invalid JSON", jsonInvalidPayload, common.ContentTypeJSON, nil, true},
		{"invalid CBOR", cborInvalidPayload, common.ContentTypeCBOR, nil, true},
	}

	target := FunctionsPipelineRuntime{lc: logger.NewMockClient()}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			envelope := types.MessageEnvelope{}
			envelope.Payload = testCase.Payload
			envelope.ContentType = testCase.ContentType

			actual, err := target.processEventPayload(envelope)
			if testCase.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.Expected, actual)
		})
	}
}

func TestTopicMatches(t *testing.T) {
	incomingTopic := "edgex/events/P/D/S"

	tests := []struct {
		name           string
		incomingTopic  string
		pipelineTopics []string
		expected       bool
	}{
		{"Match - Default all", incomingTopic, []string{TopicWildCard}, true},
		{"Match - Not First Topic", incomingTopic, []string{"not-edgex/#", TopicWildCard}, true},
		{"Match - Exact", incomingTopic, []string{incomingTopic}, true},
		{"Match - Any Profile for Device and Source", incomingTopic, []string{"edgex/events/+/D/S"}, true},
		{"Match - Any Profile for Device and Source", incomingTopic, []string{"edgex/events/+/D/S"}, true},
		{"Match - Any Device for Profile and Source", incomingTopic, []string{"edgex/events/P/+/S"}, true},
		{"Match - Any Source for Profile and Device", incomingTopic, []string{"edgex/events/P/D/#"}, true},
		{"Match - All Events ", incomingTopic, []string{"edgex/events/#"}, true},
		{"Match - First Topic Deeper ", incomingTopic, []string{"edgex/events/P/D/S/Z", "edgex/events/#"}, true},
		{"Match - All Devices and Sources for Profile ", incomingTopic, []string{"edgex/events/P/#"}, true},
		{"Match - All Sources for Profile and Device ", incomingTopic, []string{"edgex/events/P/D/#"}, true},
		{"Match - All Sources for a Device for any Profile ", incomingTopic, []string{"edgex/events/+/D/#"}, true},
		{"Match - Source for any Profile and any Device ", incomingTopic, []string{"edgex/events/+/+/S"}, true},
		{"NoMatch - SourceX for any Profile and any Device ", incomingTopic, []string{"edgex/events/+/+/Sx"}, false},
		{"NoMatch - All Sources for DeviceX and any Profile ", incomingTopic, []string{"edgex/events/+/Dx/#"}, false},
		{"NoMatch - All Sources for ProfileX and Device ", incomingTopic, []string{"edgex/events/Px/D/#"}, false},
		{"NoMatch - All Sources for Profile and DeviceX ", incomingTopic, []string{"edgex/events/P/Dx/#"}, false},
		{"NoMatch - All Sources for ProfileX and DeviceX ", incomingTopic, []string{"edgex/events/Px/Dx/#"}, false},
		{"NoMatch - All Devices and Sources for ProfileX ", incomingTopic, []string{"edgex/events/Px/#"}, false},
		{"NoMatch - Any Profile for DeviceX and Source", incomingTopic, []string{"edgex/events/+/Dx/S"}, false},
		{"NoMatch - Any Profile for DeviceX and Source", incomingTopic, []string{"edgex/events/+/Dx/S"}, false},
		{"NoMatch - Any Profile for Device and SourceX", incomingTopic, []string{"edgex/events/+/D/Sx"}, false},
		{"NoMatch - Any Profile for DeviceX and SourceX", incomingTopic, []string{"edgex/events/+/Dx/Sx"}, false},
		{"NoMatch - Any Device for Profile and SourceX", incomingTopic, []string{"edgex/events/P/+/Sx"}, false},
		{"NoMatch - Any Device for ProfileX and Source", incomingTopic, []string{"edgex/events/Px/+/S"}, false},
		{"NoMatch - Any Device for ProfileX and SourceX", incomingTopic, []string{"edgex/events/Px/+/Sx"}, false},
		{"NoMatch - Any Source for ProfileX and Device", incomingTopic, []string{"edgex/events/Px/D/#"}, false},
		{"NoMatch - Any Source for Profile and DeviceX", incomingTopic, []string{"edgex/events/P/Dx/#"}, false},
		{"NoMatch - Any Source for ProfileX and DeviceX", incomingTopic, []string{"edgex/events/Px/Dx/#"}, false},
		{"NoMatch - Pipeline Topic Deeper", incomingTopic, []string{"edgex/events/P/D/S/Z"}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := topicMatches(test.incomingTopic, test.pipelineTopics)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPipelineById(t *testing.T) {
	target := NewFunctionPipelineRuntime(serviceKey, nil, dic)

	expectedId := "my-pipeline"
	expectedTopics := []string{"edgex/events/#"}
	expectedTransforms := []interfaces.AppFunction{
		transforms.NewResponseData().SetResponseData,
	}
	badId := "bogus"

	target.SetDefaultFunctionsPipeline(expectedTransforms)

	err := target.AddFunctionsPipeline(expectedId, expectedTopics, expectedTransforms)
	require.NoError(t, err)

	actual := target.GetPipelineById(interfaces.DefaultPipelineId)
	require.NotNil(t, actual)
	assert.Equal(t, interfaces.DefaultPipelineId, actual.Id)
	assert.Equal(t, []string{TopicWildCard}, actual.Topics)
	assert.Equal(t, expectedTransforms, actual.Transforms)
	assert.NotEmpty(t, actual.Hash)

	actual = target.GetPipelineById(expectedId)
	require.NotNil(t, actual)
	assert.Equal(t, expectedId, actual.Id)
	assert.Equal(t, expectedTopics, actual.Topics)
	assert.Equal(t, expectedTransforms, actual.Transforms)
	assert.NotEmpty(t, actual.Hash)

	actual = target.GetPipelineById(badId)
	require.Nil(t, actual)
}

func TestGetMatchingPipelines(t *testing.T) {
	target := NewFunctionPipelineRuntime(serviceKey, nil, dic)

	expectedTransforms := []interfaces.AppFunction{
		transforms.NewResponseData().SetResponseData,
	}

	err := target.AddFunctionsPipeline("one", []string{"edgex/events/+/D1/#"}, expectedTransforms)
	require.NoError(t, err)
	err = target.AddFunctionsPipeline("two", []string{"edgex/events/P1/#"}, expectedTransforms)
	require.NoError(t, err)
	err = target.AddFunctionsPipeline("three", []string{"edgex/events/P1/D1/S1"}, expectedTransforms)
	require.NoError(t, err)
	err = target.AddFunctionsPipeline("four", []string{"edgex/events/device/device-virtual/+/Random-Float-Device/#"}, expectedTransforms)
	require.NoError(t, err)
	err = target.AddFunctionsPipeline("five", []string{"edgex/events/device/device-virtual/+/+/Int32"}, expectedTransforms)
	require.NoError(t, err)

	tests := []struct {
		name          string
		incomingTopic string
		expected      int
	}{
		{"Match 3", "edgex/events/P1/D1/S1", 3},
		{"Match 2", "edgex/events/P1/D1/S2", 2},
		{"Match 1", "edgex/events/P2/D1/S2", 1},
		{"Match 0", "edgex/events/P2/D2/S2", 0},
		{"Match 1", "edgex/events/device/device-virtual/Random-Float-Device/Random-Float-Device/Float32", 1},
		{"Match 1", "edgex/events/device/device-virtual/Random-Integer-Device/Random-Integer-Device/Int32", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := target.GetMatchingPipelines(test.incomingTopic)
			assert.Equal(t, test.expected, len(actual))
		})
	}
}

func TestGolangRuntime_GetDefaultPipeline(t *testing.T) {
	target := NewFunctionPipelineRuntime(serviceKey, nil, dic)

	expectedNilTransformsHash := "Pipeline-functions: "
	expectedTransforms := []interfaces.AppFunction{
		transforms.NewResponseData().SetResponseData,
	}

	// Returns dummy default pipeline with nil transforms if default never set.
	actual := target.GetDefaultPipeline()
	require.NotNil(t, actual)
	assert.Equal(t, interfaces.DefaultPipelineId, actual.Id)
	assert.Equal(t, []string{TopicWildCard}, actual.Topics)
	assert.Nil(t, actual.Transforms)
	assert.Equal(t, expectedNilTransformsHash, actual.Hash)

	target.SetDefaultFunctionsPipeline(expectedTransforms)

	actual = target.GetDefaultPipeline()
	require.NotNil(t, actual)
	assert.Equal(t, interfaces.DefaultPipelineId, actual.Id)
	assert.Equal(t, []string{TopicWildCard}, actual.Topics)
	assert.Equal(t, expectedTransforms, actual.Transforms)
	assert.NotEmpty(t, actual.Hash)
	assert.NotEqual(t, expectedNilTransformsHash, actual.Hash)

}

func TestGolangRuntime_SetFunctionsPipelineTransforms(t *testing.T) {
	target := NewFunctionPipelineRuntime(serviceKey, nil, dic)

	id := "my-pipeline"
	topics := []string{"edgex/events/#"}
	initialTransforms := []interfaces.AppFunction{
		transforms.NewResponseData().SetResponseData,
	}

	compress := transforms.NewCompression()

	expectedTransforms := []interfaces.AppFunction{
		compress.CompressWithGZIP,
		transforms.NewResponseData().SetResponseData,
	}

	err := target.AddFunctionsPipeline(id, topics, initialTransforms)
	require.NoError(t, err)
	pipeline := target.GetPipelineById(id)
	require.NotEqual(t, expectedTransforms, pipeline.Transforms)

	target.SetFunctionsPipelineTransforms(id, expectedTransforms)
	pipeline = target.GetPipelineById(id)
	require.Equal(t, expectedTransforms, pipeline.Transforms)
}

func TestGolangRuntime_ClearAllFunctionsPipelineTransforms(t *testing.T) {
	target := NewFunctionPipelineRuntime(serviceKey, nil, dic)

	id1 := "pipeline1"
	id2 := "pipeline2"
	topics := []string{"edgex/events/#"}
	transforms := []interfaces.AppFunction{
		transforms.NewResponseData().SetResponseData,
	}

	target.SetDefaultFunctionsPipeline(transforms)
	err := target.AddFunctionsPipeline(id1, topics, transforms)
	require.NoError(t, err)
	err = target.AddFunctionsPipeline(id2, topics, transforms)
	require.NoError(t, err)
	pipeline := target.GetDefaultPipeline()
	require.NotNil(t, pipeline.Transforms)
	pipeline = target.GetPipelineById(id1)
	require.NotNil(t, pipeline.Transforms)
	pipeline = target.GetPipelineById(id2)
	require.NotNil(t, pipeline.Transforms)

	target.ClearAllFunctionsPipelineTransforms()
	pipeline = target.GetDefaultPipeline()
	require.Nil(t, pipeline.Transforms)
	pipeline = target.GetPipelineById(id1)
	assert.Nil(t, pipeline.Transforms)
	pipeline = target.GetPipelineById(id2)
	assert.Nil(t, pipeline.Transforms)
}

func TestFunctionPipelineRuntime_RemoveAllFunctionPipelines(t *testing.T) {
	target := NewFunctionPipelineRuntime(serviceKey, nil, dic)

	id1 := "pipeline1"
	id2 := "pipeline2"
	topics := []string{"edgex/events/#"}
	transforms := []interfaces.AppFunction{
		transforms.NewResponseData().SetResponseData,
	}

	target.SetDefaultFunctionsPipeline(transforms)
	err := target.AddFunctionsPipeline(id1, topics, transforms)
	require.NoError(t, err)
	err = target.AddFunctionsPipeline(id2, topics, transforms)
	require.NoError(t, err)

	pipeline := target.GetDefaultPipeline()
	require.NotNil(t, pipeline)
	pipeline = target.GetPipelineById(id1)
	require.NotNil(t, pipeline)
	pipeline = target.GetPipelineById(id2)
	require.NotNil(t, pipeline)

	target.RemoveAllFunctionPipelines()

	pipeline = target.GetDefaultPipeline()
	require.Nil(t, pipeline.Transforms)
	pipeline = target.GetPipelineById(id1)
	assert.Nil(t, pipeline)
	pipeline = target.GetPipelineById(id2)
	assert.Nil(t, pipeline)
}
