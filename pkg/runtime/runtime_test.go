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
package runtime

import (
	"errors"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/context"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	devID1        = "id1"
	devID2        = "id2"
	readingName1  = "sensor1"
	readingValue1 = "123.45"
)

func TestProcessEventNoTransforms(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	context := context.Context{}
	runtime := GolangRuntime{}
	result := runtime.ProcessEvent(context, eventIn)
	if result != nil {
		t.Fatal("result should be nil since no transforms have been passed")
	}
}
func TestProcessEventOneCustomTransform(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	context := context.Context{}
	transform1WasCalled := false
	transform1 := func(params ...interface{}) (bool, interface{}) {
		if len(params) != 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(*models.Event); ok {
			if ok == false {
				t.Fatal("Should have receieved CoreData event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected CoreData event")
			}
		}
		transform1WasCalled = true
		return true, "Hello"
	}
	runtime := GolangRuntime{
		Transforms: []func(params ...interface{}) (bool, interface{}){transform1},
	}
	result := runtime.ProcessEvent(context, eventIn)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
}
func TestProcessEventTwoCustomTransforms(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	context := context.Context{}
	transform1WasCalled := false
	transform2WasCalled := false

	transform1 := func(params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true
		if len(params) != 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(*models.Event); ok {
			if ok == false {
				t.Fatal("Should have receieved CoreData event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected CoreData event")
			}
		}

		return true, "Transform1Result"
	}
	transform2 := func(params ...interface{}) (bool, interface{}) {
		transform2WasCalled = true

		if params[0] != "Transform1Result" {
			t.Fatal("Did not recieve result from previous transform")
		}
		return true, "Hello"
	}
	runtime := GolangRuntime{
		Transforms: []func(params ...interface{}) (bool, interface{}){transform1, transform2},
	}
	result := runtime.ProcessEvent(context, eventIn)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
	if transform2WasCalled == false {
		t.Fatal("transform2 should have been called")
	}
}
func TestProcessEventThreeCustomTransformsOneFail(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	context := context.Context{}
	transform1WasCalled := false
	transform2WasCalled := false
	transform3WasCalled := false

	transform1 := func(params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true
		if len(params) != 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(*models.Event); ok {
			if ok == false {
				t.Fatal("Should have receieved CoreData event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected CoreData event")
			}
		}

		return false, errors.New("Transform1Result")
	}
	transform2 := func(params ...interface{}) (bool, interface{}) {
		transform2WasCalled = true

		if params[0] != "Transform1Result" {
			t.Fatal("Did not recieve result from previous transform")
		}
		return true, "Hello"
	}
	transform3 := func(params ...interface{}) (bool, interface{}) {
		transform3WasCalled = true

		if params[0] != "Transform1Result" {
			t.Fatal("Did not recieve result from previous transform")
		}
		return true, "Hello"
	}
	runtime := GolangRuntime{
		Transforms: []func(params ...interface{}) (bool, interface{}){transform1, transform2, transform3},
	}
	result := runtime.ProcessEvent(context, eventIn)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
	if transform2WasCalled == true {
		t.Fatal("transform2 should NOT have been called")
	}
	if transform3WasCalled == true {
		t.Fatal("transform3 should NOT have been called")
	}
}
