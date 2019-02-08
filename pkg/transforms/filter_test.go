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
	"testing"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func TestFilterByDeviceIDFound(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	filter := Filter{
		FilterValues: []string{"id1"},
	}
	continuePipeline, result := filter.FilterByDeviceID(eventIn)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
	}
	if eventOut, ok := result.(*models.Event); ok {
		if eventOut.Device != "id1" {
			t.Fatal("device id does not match filter")
		}
	}
}
func TestFilterByDeviceIDNotFound(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	filter := Filter{
		FilterValues: []string{"id2"},
	}
	continuePipeline, result := filter.FilterByDeviceID(eventIn)
	if result != nil {
		t.Fatal("result should be nil")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}

func TestFilterByDeviceIDNoParameters(t *testing.T) {
	filter := Filter{
		FilterValues: []string{"id2"},
	}
	continuePipeline, result := filter.FilterByDeviceID()
	if result != nil {
		t.Fatal("result should be nil")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}
