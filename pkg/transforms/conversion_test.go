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
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

var context *appcontext.Context
var lc logger.LoggingClient
var config *common.ConfigurationStruct

const (
	devID1 = "id1"
	devID2 = "id2"
)

func TestMain(m *testing.M) {
	lc := logger.NewMockClient()

	insecureSecrets := common.InsecureSecrets{
		"no_path": common.InsecureSecretsInfo{
			Path: "",
			Secrets: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		"db_secrets": common.InsecureSecretsInfo{
			Path: "db_secrets",
			Secrets: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	config = &common.ConfigurationStruct{
		Writable: common.WritableInfo{
			InsecureSecrets: insecureSecrets,
		},
	}
	eventClient := coredata.NewEventClient(local.New("http://test" + clients.ApiEventRoute))
	mockSP := newMockSecretProvider(lc, config)

	context = &appcontext.Context{
		LoggingClient:  lc,
		EventClient:    eventClient,
		SecretProvider: mockSP,
	}

	m.Run()
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
