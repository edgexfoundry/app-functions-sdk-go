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
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsProcessor(t *testing.T) {
	actual, err := NewMetricsProcessor(nil)
	require.NoError(t, err)
	require.NotNil(t, actual)
	assert.Empty(t, actual.additionalTags)

	inputTags := map[string]interface{}{
		"Tag1": "str1",
		"Tag2": 123,
		"Tag3": 12.34,
	}
	expectedTags := []dtos.MetricTag{
		{Name: "Tag1", Value: "str1"},
		{Name: "Tag2", Value: "123"},
		{Name: "Tag3", Value: "12.34"},
	}

	actual, err = NewMetricsProcessor(inputTags)
	require.NoError(t, err)
	require.NotNil(t, actual)
	require.Len(t, actual.additionalTags, len(expectedTags))

	// Convert the slices to maps to make comparison easier since the contents of the two slices are not always in the same order.
	expectedTagMap := metricTagSliceToMap(expectedTags)
	actualTagMap := metricTagSliceToMap(actual.additionalTags)
	for expectedName, expectedValue := range expectedTagMap {
		assert.Equal(t, expectedValue, actualTagMap[expectedName])
	}
}

func metricTagSliceToMap(tags []dtos.MetricTag) map[string]string {
	tagMap := make(map[string]string)
	for _, item := range tags {
		tagMap[item.Name] = item.Value
	}

	return tagMap
}
func TestMetricsProcessor_ToLineProtocol(t *testing.T) {
	target, err := NewMetricsProcessor(map[string]interface{}{"Tag1": "value1"})
	require.NoError(t, err)
	expectedTimestamp := time.Now().UnixMilli()
	expectedContinue := true
	expectedResult := fmt.Sprintf("UnitTestMetric,ServiceName=UnitTestService,SomeTag=SomeValue,Tag1=value1 int=12i,float=12.35,uint=99u %d\n", expectedTimestamp)
	source := dtos.Metric{
		Versionable: common.NewVersionable(),
		Name:        "UnitTestMetric",
		Fields: []dtos.MetricField{
			{
				Name:  "int",
				Value: 12,
			},
			{
				Name:  "float",
				Value: 12.35,
			},
			{
				Name:  "uint",
				Value: uint(99),
			},
		},
		Tags: []dtos.MetricTag{
			{
				Name:  "ServiceName",
				Value: "UnitTestService",
			},
			{
				Name:  "SomeTag",
				Value: "SomeValue",
			},
		},
		Timestamp: expectedTimestamp,
	}

	expectedOrigTagCount := len(source.Tags)
	actualContinue, actualResult := target.ToLineProtocol(ctx, source)
	assert.Equal(t, expectedContinue, actualContinue)
	assert.Equal(t, expectedResult, actualResult)
	assert.Equal(t, expectedOrigTagCount, len(source.Tags))
}
