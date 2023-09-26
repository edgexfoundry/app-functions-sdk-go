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

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	mocks2 "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces/mocks"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	gometrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateRegisterMetric(t *testing.T) {
	expectedName := "testCounter"
	expectedFullName := expectedName + "-https://somewhere.com"
	expectedUrl := "https://somewhere.com"
	expectedTags := map[string]string{"url": expectedUrl}

	tests := []struct {
		Name              string
		NilMetricsManager bool
		RegisterError     error
	}{
		{"Happy Path", false, nil},
		{"Error - No Metrics Manager", true, nil},
		{"Error - Register error", false, errors.New("register failed")},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockMetricsMgr := &mocks2.MetricsManager{}
			mockMetricsMgr.On("Register", expectedFullName, mock.Anything, expectedTags).Return(test.RegisterError).Once()
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debugf", mock.Anything, mock.Anything)
			mockLogger.On("Infof", mock.Anything, mock.Anything)
			mockCtx := &mocks.AppFunctionContext{}
			mockCtx.On("LoggingClient").Return(mockLogger)
			if test.NilMetricsManager {
				mockCtx.On("MetricsManager").Return(nil)
				mockLogger.On("Errorf", mock.Anything, expectedFullName, "metrics manager not available")
			} else {
				mockCtx.On("MetricsManager").Return(mockMetricsMgr)
			}

			if test.RegisterError != nil {
				mockLogger.On("Errorf", mock.Anything, expectedFullName, "register failed")
			}

			var metric gometrics.Counter
			createRegisterMetric(mockCtx,
				func() string { return expectedFullName },
				func() any { return metric },
				func() { metric = gometrics.NewCounter() },
				expectedTags)
			require.NotNil(t, metric)

		})
	}
}
