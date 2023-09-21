//
// Copyright (c) 2023 Intel Corporation
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
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
)

func createRegisterMetric(ctx interfaces.AppFunctionContext,
	fullNameFunc func() string, getMetric func() any, setMetric func(),
	tags map[string]string) {
	// Only need to create and register the metric if it hasn't been created yet.
	if getMetric() == nil {
		lc := ctx.LoggingClient()
		var err error
		fullName := fullNameFunc()
		lc.Debugf("Initializing metric %s.", fullName)
		setMetric()
		metricsManger := ctx.MetricsManager()
		if metricsManger != nil {
			err = metricsManger.Register(fullName, getMetric(), tags)
		} else {
			err = errors.New("metrics manager not available")
		}

		if err != nil {
			lc.Errorf("Unable to register metric %s. Collection will continue, but metric will not be reported: %s", fullName, err.Error())
			return
		}

		lc.Infof("%s metric has been registered and will be reported (if enabled)", fullName)
	}
}
