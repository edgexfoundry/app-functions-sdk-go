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
//

package transforms

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tagsToAdd = map[string]string{
	"GatewayId": "HoustonStore000123",
	"Latitude":  "29.630771",
	"Longitude": "-95.377603",
}

var eventWithExistingTags = dtos.Event{
	Tags: map[string]string{
		"Tag1": "Value1",
		"Tag2": "Value2",
	},
}

var allTagsAdded = map[string]string{
	"Tag1":      "Value1",
	"Tag2":      "Value2",
	"GatewayId": "HoustonStore000123",
	"Latitude":  "29.630771",
	"Longitude": "-95.377603",
}

func TestTags_AddTags(t *testing.T) {
	tests := []struct {
		Name          string
		FunctionInput interface{}
		TagsToAdd     map[string]string
		Expected      map[string]string
		ErrorExpected bool
		ErrorContains string
	}{
		{"Happy path - no existing Event tags", dtos.Event{}, tagsToAdd, tagsToAdd, false, ""},
		{"Happy path - Event has existing tags", eventWithExistingTags, tagsToAdd, allTagsAdded, false, ""},
		{"Happy path - No tags added", eventWithExistingTags, map[string]string{}, eventWithExistingTags.Tags, false, ""},
		{"Error - No data", nil, nil, nil, true, "No Data Received"},
		{"Error - Input not event", "Not an Event", nil, nil, true, "type received is not an Event"},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			var continuePipeline bool
			var result interface{}

			target := NewTags(testCase.TagsToAdd)

			if testCase.FunctionInput != nil {
				continuePipeline, result = target.AddTags(ctx, testCase.FunctionInput)
			} else {
				continuePipeline, result = target.AddTags(ctx, nil)
			}

			if testCase.ErrorExpected {
				err := result.(error)
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ErrorContains)
				require.False(t, continuePipeline)
				return // Test completed
			}

			assert.True(t, continuePipeline)
			actual, ok := result.(dtos.Event)
			require.True(t, ok, "Result not an Event")
			assert.Equal(t, testCase.Expected, actual.Tags)
		})
	}
}
