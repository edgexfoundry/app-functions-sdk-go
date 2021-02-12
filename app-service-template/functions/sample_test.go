// TODO: Change Copyright to your company if open sourcing or remove header
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

package functions

import (
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// This file contains example of how to unit test pipeline functions
// TODO: Change these sample unit tests to test your custom type and function(s)

var TestEventId = uuid.New().String()

func TestSample_LogEventDetails(t *testing.T) {
	expectedEvent := createTestEvent()
	expectedContinuePipeline := true

	target := NewSample()
	actualContinuePipeline, actualEvent := target.LogEventDetails(creatTestAppSdkContext(), expectedEvent)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Equal(t, expectedEvent, actualEvent)
}

func TestSample_ConvertEventToXML(t *testing.T) {
	event := createTestEvent()
	expectedXml, _ := event.ToXML()
	expectedContinuePipeline := true

	target := NewSample()
	actualContinuePipeline, actualXml := target.ConvertEventToXML(creatTestAppSdkContext(), event)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Equal(t, expectedXml, actualXml)

}

func TestSample_OutputXML(t *testing.T) {
	expectedXml, _ := createTestEvent().ToXML()
	expectedContinuePipeline := false
	appContext := creatTestAppSdkContext()

	target := NewSample()
	actualContinuePipeline, result := target.OutputXML(appContext, expectedXml)
	actualXml := string(appContext.OutputData)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Nil(t, result)
	assert.Equal(t, expectedXml, actualXml)

}

func createTestEvent() models.Event {
	deviceName := "MyDevice"

	event := models.Event{
		ID:     TestEventId,
		Device: deviceName,
		Origin: time.Now().UnixNano(),
		Tags:   make(map[string]string),
	}

	event.Readings = append(event.Readings, models.Reading{
		Id:        uuid.New().String(),
		Origin:    time.Now().UnixNano(),
		Device:    deviceName,
		Name:      "MyReading",
		Value:     "1234",
		ValueType: models.ValueTypeInt32,
	})

	event.Tags["WhereAmI"] = "NotKansas"

	return event
}

func creatTestAppSdkContext() *appcontext.Context {
	return &appcontext.Context{
		EventID:       TestEventId,
		CorrelationID: uuid.New().String(),
		LoggingClient: logger.NewMockClient(),
	}
}
