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
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces/mocks"
	mocks2 "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger/mocks"
	gometrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterMetric(t *testing.T) {
	expectedName := "testCounter"
	expectedFullName := expectedName + "-https://somewhere.com"
	expectedUrl := "https://somewhere.com"
	expectedTags := map[string]string{"url": expectedUrl}

	tests := []struct {
		Name              string
		NilMetricsManager bool
		RegisterError     error
		alreadyRegistered bool
	}{
		{"Happy Path", false, nil, false},
		{"Happy Path - already registered", false, nil, false},
		{"Error - No Metrics Manager", true, nil, false},
		{"Error - Register error", false, errors.New("register failed"), false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockMetricsMgr := &mocks2.MetricsManager{}
			mockMetricsMgr.On("IsRegistered", mock.Anything).Return(test.alreadyRegistered)
			mockMetricsMgr.On("Register", expectedFullName, mock.Anything, expectedTags).Return(test.RegisterError).Once()
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debugf", mock.Anything, mock.Anything)
			mockLogger.On("Infof", mock.Anything, mock.Anything)
			mockCtx := &mocks.AppFunctionContext{}
			mockCtx.On("LoggingClient").Return(mockLogger)
			if test.NilMetricsManager {
				mockCtx.On("MetricsManager").Return(nil)
				mockLogger.On("Errorf", "Metrics manager not available. Unable to register %s metric", expectedFullName)
			} else {
				mockCtx.On("MetricsManager").Return(mockMetricsMgr)
			}

			if test.RegisterError != nil {
				mockLogger.On("Errorf", mock.Anything, expectedFullName, "register failed")
			}

			metric := gometrics.NewCounter()
			registerMetric(mockCtx,
				func() string { return expectedFullName },
				func() any { return metric },
				expectedTags)
			require.NotNil(t, metric)

		})
	}
}
