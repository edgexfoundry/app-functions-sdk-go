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
	"strconv"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/stretchr/testify/require"
)

func TestEventWrapper_Wrap(t *testing.T) {
	obj := struct {
		myInt int
		myStr string
	}{
		myInt: 3,
		myStr: "hello world",
	}
	tests := []struct {
		Name          string
		ProfileName   string
		DeviceName    string
		ResourceName  string
		ValueType     string
		MediaType     string
		Data          interface{}
		ExpectedError interface{}
	}{
		{"Successful Binary Reading", "MyProfile", "MyDevice", "BinaryEvent", common.ValueTypeBinary, "stuff", true, ""},
		{"Successful Object Reading", "MyProfile", "MyDevice", "ObjectEvent", common.ValueTypeObject, "", obj, ""},
		{"Successful Simple Reading", "MyProfile", "MyDevice", "ObjectEvent", common.ValueTypeString, "", "hello there", ""},
	}

	for _, test := range tests {
		var transform *EventWrapper
		switch test.ValueType {
		case common.ValueTypeBinary:
			transform = NewEventWrapperBinaryReading(test.ProfileName, test.DeviceName, test.ResourceName, test.MediaType)
		case common.ValueTypeObject:
			transform = NewEventWrapperObjectReading(test.ProfileName, test.DeviceName, test.ResourceName)
		default:
			transform = NewEventWrapperSimpleReading(test.ProfileName, test.DeviceName, test.ResourceName, test.ValueType)
		}
		actualBool, actualInterface := transform.Wrap(ctx, test.Data)
		if test.ExpectedError == "" {
			require.True(t, actualBool)
			require.Equal(t, "", ctx.ResponseContentType())
			ctxValues := ctx.GetAllValues()
			require.Equal(t, test.ProfileName, ctxValues[interfaces.PROFILENAME])
			require.Equal(t, test.DeviceName, ctxValues[interfaces.DEVICENAME])
			require.Equal(t, test.ResourceName, ctxValues[interfaces.SOURCENAME])
			eventRequest := actualInterface.(requests.AddEventRequest)
			require.Equal(t, test.DeviceName, eventRequest.Event.DeviceName)
			require.Equal(t, test.ProfileName, eventRequest.Event.ProfileName)
			require.Equal(t, test.ResourceName, eventRequest.Event.SourceName)
			require.Equal(t, test.DeviceName, eventRequest.Event.Readings[0].DeviceName)
			require.Equal(t, test.ProfileName, eventRequest.Event.Readings[0].ProfileName)
			require.Equal(t, test.ResourceName, eventRequest.Event.Readings[0].ResourceName)
			switch test.ValueType {
			case common.ValueTypeBinary:
				value, err := strconv.ParseBool(string(eventRequest.Event.Readings[0].BinaryReading.BinaryValue))
				require.NoError(t, err)
				require.Equal(t, test.Data, value)
			case common.ValueTypeObject:
				require.Equal(t, test.Data, eventRequest.Event.Readings[0].ObjectReading.ObjectValue)
			default:
				require.Equal(t, test.Data, eventRequest.Event.Readings[0].SimpleReading.Value)
			}
			return
		}
		require.False(t, actualBool)
		require.Equal(t, test.ExpectedError, actualInterface)
	}
}
