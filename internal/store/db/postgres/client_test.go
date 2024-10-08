//go:build postgresRunning
// +build postgresRunning

/*******************************************************************************
 * Copyright (C) 2024 IOTech Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

// This test will only be executed if the tag postgresRunning is added when running
// the tests with a command like:
// go test -tags postgresRunning

package postgres

import (
	"context"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/store/db"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/stretchr/testify/require"
)

const (
	TestHost      = "localhost"
	TestPort      = 5432
	TestTimeout   = "5s"
	TestMaxIdle   = 5000
	TestBatchSize = 1337

	TestRetryCount       = 100
	TestPipelinePosition = 1337
	TestVersion          = "your"
	TestCorrelationID    = "test"
	TestPipelineId       = "test-pipeline"
)

var TestValidNoAuthConfig = bootstrapConfig.Database{
	Type:    db.Postgres,
	Host:    TestHost,
	Port:    TestPort,
	Timeout: TestTimeout,
}

var TestCredential = bootstrapConfig.Credentials{Username: "postgres", Password: "mysecretpassword"}

func TestClient_NewClient(t *testing.T) {
	tests := []struct {
		name          string
		config        bootstrapConfig.Database
		expectedError bool
	}{
		{"Success, no auth", TestValidNoAuthConfig, false},
	}

	lc := logger.NewMockClient()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewClient(context.Background(), test.config, TestCredential, "", "", lc, "svcKey")

			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
