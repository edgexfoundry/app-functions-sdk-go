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

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"
)

// MetricsProcessor contains functions to process the Metric DTO
type MetricsProcessor struct {
	additionalTags []dtos.MetricTag
}

// NewMetricsProcessor creates a new MetricsProcessor with additional tags to add to the Metrics that are processed
func NewMetricsProcessor(additionalTags map[string]interface{}) (*MetricsProcessor, error) {
	mp := &MetricsProcessor{}

	for name, value := range additionalTags {
		if err := dtos.ValidateMetricName(name, "Tag"); err != nil {
			return nil, err
		}

		mp.additionalTags = append(mp.additionalTags, dtos.MetricTag{Name: name, Value: fmt.Sprintf("%v", value)})
	}

	return mp, nil
}

// ToLineProtocol transforms a Metric DTO to a string conforming to Line Protocol syntax which is most commonly used with InfluxDB
// For more information on Line Protocol see: https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/
func (mp *MetricsProcessor) ToLineProtocol(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()
	lc.Debugf("ToLineProtocol called in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		// Go here for details on Error Handle: https://docs.edgexfoundry.org/1.3/microservices/application/ErrorHandling/
		return false, fmt.Errorf("function ToLineProtocol in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	metric, ok := data.(dtos.Metric)
	if !ok {
		return false, fmt.Errorf("function ToLineProtocol in pipeline '%s', type received is not an Metric", ctx.PipelineId())
	}

	if len(mp.additionalTags) > 0 {
		metric.Tags = append(metric.Tags, mp.additionalTags...)
	}

	// New line is needed if the resulting metric data is batched and sent in chunks to service like InfluxDB
	result := fmt.Sprintln(metric.ToLineProtocol())

	lc.Debugf("Transformed Metric to '%s' in pipeline '%s", result, ctx.PipelineId())

	return true, result
}
