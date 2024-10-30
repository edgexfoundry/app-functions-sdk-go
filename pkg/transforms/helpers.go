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
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"
)

func registerMetric(ctx interfaces.AppFunctionContext, fullNameFunc func() string, getMetric func() any, tags map[string]string) {
	lc := ctx.LoggingClient()
	fullName := fullNameFunc()

	metricsManger := ctx.MetricsManager()
	if metricsManger == nil {
		lc.Errorf("Metrics manager not available. Unable to register %s metric", fullName)
		return
	}

	// Only register the metric if it hasn't been registered yet.
	if !metricsManger.IsRegistered(fullName) {
		var err error
		lc.Debugf("Registering metric %s.", fullName)
		err = metricsManger.Register(fullName, getMetric(), tags)

		if err != nil {
			// In case of race condition, check again if metric was registered by another thread
			if !metricsManger.IsRegistered(fullName) {
				lc.Errorf("Unable to register metric %s. Collection will continue, but metric will not be reported: %s", fullName, err.Error())
			}
			return
		}

		lc.Infof("%s metric has been registered and will be reported (if enabled)", fullName)
	}
}
