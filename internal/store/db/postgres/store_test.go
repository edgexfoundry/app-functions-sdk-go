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

// This test will only be executed if the tag redisRunning is added when running
// the tests with a command like:
// go test -tags postgresRunning

package postgres

import (
	"context"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var TestPayload = []byte("brandon was here")

var TestContractBase = interfaces.StoredObject{
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
}

var TestContractNoPayload = interfaces.StoredObject{
	AppServiceKey:    uuid.New().String(),
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
}

var TestContractNoVersion = interfaces.StoredObject{
	AppServiceKey:    uuid.New().String(),
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	CorrelationID:    TestCorrelationID,
}

var TestContractBadID = interfaces.StoredObject{
	ID:               "brandon!",
	AppServiceKey:    "brandon!",
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
}

var TestContractNoAppServiceKey = interfaces.StoredObject{
	ID:               uuid.New().String(),
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelineId:       TestPipelineId,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
}
var postgresClient *Client

// setupPostgresClient is called before the tests run
func setupPostgresClient() {
	lc := logger.NewMockClient()
	client, _ := NewClient(context.Background(), TestValidNoAuthConfig, TestCredential, "", "", lc)
	postgresClient = client
}

func TestMain(m *testing.M) {
	setupPostgresClient()
	defer postgresClient.Disconnect()

	m.Run()
}

func TestClient_Store(t *testing.T) {
	TestContractUUID := TestContractBase
	TestContractUUID.ID = uuid.New().String()
	TestContractUUID.AppServiceKey = uuid.New().String()

	TestContractValid := TestContractBase
	TestContractValid.AppServiceKey = uuid.New().String()

	TestContractNoAppServiceKey := TestContractBase
	TestContractUUID.ID = uuid.New().String()

	tests := []struct {
		name          string
		toStore       interfaces.StoredObject
		expectedError bool
	}{
		{
			"Success, no ID",
			TestContractValid,
			false,
		},
		{
			"Success, no ID double store",
			TestContractValid,
			false,
		},
		{
			"Failure, no app service key",
			TestContractNoAppServiceKey,
			true,
		},
		{
			"Failure, no app service key double store",
			TestContractNoAppServiceKey,
			true,
		},
		{
			"Success, object with UUID",
			TestContractUUID,
			false,
		},
		{
			"Failure, no payload",
			TestContractNoPayload,
			true,
		},
		{
			"Failure, no version",
			TestContractNoVersion,
			true,
		},
		{
			"Failure, bad ID",
			TestContractBadID,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			returnVal, err := postgresClient.Store(test.toStore)

			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			require.NotEqual(t, "", returnVal, "Function did not error but did not return a valid ID")
			storedObj := test.toStore
			storedObj.ID = returnVal
			err = postgresClient.RemoveFromStore(storedObj)
			require.NoError(t, err)
		})
	}
}

func TestClient_RetrieveFromStore(t *testing.T) {
	UUIDAppServiceKey := uuid.New().String()

	UUIDContract0 := TestContractBase
	UUIDContract0.ID = uuid.New().String()
	UUIDContract0.AppServiceKey = UUIDAppServiceKey

	UUIDContract1 := TestContractBase
	UUIDContract1.ID = uuid.New().String()
	UUIDContract1.AppServiceKey = UUIDAppServiceKey

	tests := []struct {
		name          string
		toStore       []interfaces.StoredObject
		key           string
		expectedError bool
	}{
		{
			"Success, single object",
			[]interfaces.StoredObject{UUIDContract0},
			UUIDAppServiceKey,
			false,
		},
		{
			"Success, multiple object",
			[]interfaces.StoredObject{UUIDContract0, UUIDContract1},
			UUIDAppServiceKey,
			false,
		},
		{
			"Failure, no app service key",
			[]interfaces.StoredObject{},
			"",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, object := range test.toStore {
				_, _ = postgresClient.Store(object)
			}

			actual, err := postgresClient.RetrieveFromStore(test.key)

			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, len(actual), len(test.toStore), "Returned slice length doesn't match expected")

			for _, object := range test.toStore {
				_ = postgresClient.RemoveFromStore(object)
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	TestContractValid := TestContractBase
	TestContractValid.AppServiceKey = uuid.New().String()
	TestContractValid.Version = uuid.New().String()

	// add the objects we're going to update in the database now so we have a known state
	TestContractValid.ID, _ = postgresClient.Store(TestContractValid)

	tests := []struct {
		name          string
		expectedVal   interfaces.StoredObject
		expectedError bool
	}{
		{
			"Success",
			TestContractValid,
			false,
		},
		{
			"Failure, no UUID",
			TestContractBase,
			true,
		},
		{
			"Failure, no app service key",
			TestContractNoAppServiceKey,
			true,
		},
		{
			"Failure, no payload",
			TestContractNoPayload,
			true,
		},
		{
			"Failure, no version",
			TestContractNoVersion,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := postgresClient.Update(test.expectedVal)

			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			// only do a lookup on tests that we aren't expecting errors
			actual, _ := postgresClient.RetrieveFromStore(test.expectedVal.AppServiceKey)
			require.NotNil(t, actual, "No objects retrieved from store")
			require.Equal(t, test.expectedVal, actual[0], "Return value doesn't match expected")

			_ = postgresClient.RemoveFromStore(test.expectedVal)
		})
	}
}

func TestClient_RemoveFromStore(t *testing.T) {
	TestContractValid := TestContractBase
	TestContractValid.AppServiceKey = uuid.New().String()

	// add the objects we're going to update in the database now so we have a known state
	var storeErr error
	TestContractValid.ID, storeErr = postgresClient.Store(TestContractValid)
	require.NoError(t, storeErr)

	tests := []struct {
		name          string
		testObject    interfaces.StoredObject
		expectedError bool
	}{
		{
			"Success",
			TestContractValid,
			false,
		},
		{
			"Failure, no app service key",
			TestContractNoAppServiceKey,
			true,
		},
		{
			"Failure, no payload",
			TestContractNoPayload,
			true,
		},
		{
			"Failure, no version",
			TestContractNoVersion,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := postgresClient.RemoveFromStore(test.testObject)

			if test.expectedError {
				require.Error(t, err)
				return // test complete
			} else {
				require.NoError(t, err)
			}

			// only do a lookup on tests that we aren't expecting errors
			actual, _ := postgresClient.RetrieveFromStore(test.testObject.AppServiceKey)
			require.Len(t, actual, 0)
		})
	}
}
