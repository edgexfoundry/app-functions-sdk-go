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

package runtime

import (
	"errors"
	"os"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/appfunction"
	mocks2 "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces/mocks"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	common2 "github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/transforms"
)

var dic *di.Container

func TestMain(m *testing.M) {
	config := common.ConfigurationStruct{
		Writable: common.WritableInfo{
			LogLevel:        "DEBUG",
			StoreAndForward: common.StoreAndForwardInfo{Enabled: true, MaxRetryCount: 10},
		},
	}

	mockMetricsManager := &mocks2.MetricsManager{}
	mockMetricsManager.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMetricsManager.On("Unregister", mock.Anything)

	dic = di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return mockMetricsManager
		},
	})

	os.Exit(m.Run())
}

func TestProcessRetryItems(t *testing.T) {

	targetTransformWasCalled := false
	expectedPayload := "This is a sample payload"
	contextData := map[string]string{"x": "y"}

	transformPassthru := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return true, data
	}

	successTransform := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		targetTransformWasCalled = true

		actualPayload, ok := data.([]byte)

		require.True(t, ok, "Expected []byte payload")
		require.Equal(t, expectedPayload, string(actualPayload))
		require.Equal(t, contextData, appContext.GetAllValues())
		return false, nil
	}

	failureTransform := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		targetTransformWasCalled = true
		require.Equal(t, contextData, appContext.GetAllValues())
		return false, errors.New("I failed")
	}

	tests := []struct {
		Name                     string
		TargetTransform          interfaces.AppFunction
		TargetTransformWasCalled bool
		ExpectedPayload          string
		RetryCount               int
		ExpectedRetryCount       int
		RemoveCount              int
		BadVersion               bool
		ContextData              map[string]string
		UsePerTopic              bool
	}{
		{"Happy Path - Default", successTransform, true, expectedPayload, 0, 0, 1, false, contextData, false},
		{"RetryCount Increased - Default", failureTransform, true, expectedPayload, 4, 5, 0, false, contextData, false},
		{"Max Retries - Default", failureTransform, true, expectedPayload, 9, 9, 1, false, contextData, false},
		{"Bad Version - Default", successTransform, false, expectedPayload, 0, 0, 1, true, contextData, false},
		{"Happy Path - Per Topics", successTransform, true, expectedPayload, 0, 0, 1, false, contextData, true},
		{"RetryCount Increased - Per Topics", failureTransform, true, expectedPayload, 4, 5, 0, false, contextData, true},
		{"Max Retries - Per Topics", failureTransform, true, expectedPayload, 9, 9, 1, false, contextData, true},
		{"Bad Version - Per Topics", successTransform, false, expectedPayload, 0, 0, 1, true, contextData, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			targetTransformWasCalled = false
			runtime := NewFunctionPipelineRuntime(serviceKey, nil, dic)

			var pipeline *interfaces.FunctionPipeline

			if test.UsePerTopic {
				err := runtime.AddFunctionsPipeline("per-topic", []string{"#"}, []interfaces.AppFunction{transformPassthru, transformPassthru, test.TargetTransform})
				require.NoError(t, err)
				pipeline = runtime.GetPipelineById("per-topic")
				require.NotNil(t, pipeline)
			} else {
				runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transformPassthru, transformPassthru, test.TargetTransform})
				pipeline = runtime.GetDefaultPipeline()
				require.NotNil(t, pipeline)
			}

			version := pipeline.Hash
			if test.BadVersion {
				version = "some bad version"
			}
			storedObject := interfaces.NewStoredObject("dummy", []byte(test.ExpectedPayload), pipeline.Id, 2, version, contextData)
			storedObject.RetryCount = test.RetryCount

			removes, updates := runtime.storeForward.processRetryItems([]interfaces.StoredObject{storedObject})
			assert.Equal(t, test.TargetTransformWasCalled, targetTransformWasCalled, "Target transform not called")
			if test.RetryCount != test.ExpectedRetryCount {
				if assert.True(t, len(updates) > 0, "Remove count not as expected") {
					assert.Equal(t, test.ExpectedRetryCount, updates[0].RetryCount, "Retry Count not as expected")
				}
			}
			assert.Equal(t, test.RemoveCount, len(removes), "Remove count not as expected")
		})
	}
}

func TestDoStoreAndForwardRetry(t *testing.T) {
	payload := []byte("My Payload")

	httpPost := transforms.NewHTTPSender("http://nowhere", "", true).HTTPPost
	successTransform := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return false, nil
	}
	transformPassthru := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return true, data
	}

	tests := []struct {
		Name                string
		TargetTransform     interfaces.AppFunction
		RetryCount          int
		ExpectedRetryCount  int
		ExpectedObjectCount int
		UsePerTopic         bool
	}{
		{"RetryCount Increased - Default", httpPost, 1, 2, 1, false},
		{"Max Retries - Default", httpPost, 9, 0, 0, false},
		{"Retry Success - Default", successTransform, 1, 0, 0, false},
		{"RetryCount Increased - Per Topics", httpPost, 1, 2, 1, true},
		{"Max Retries - Per Topics", httpPost, 9, 0, 0, true},
		{"Retry Success - Per Topics", successTransform, 1, 0, 0, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runtime := NewFunctionPipelineRuntime(serviceKey, nil, updateDicWithMockStoreClient())
			runtime.storeForward.dataCount.Inc(1)

			var pipeline *interfaces.FunctionPipeline

			if test.UsePerTopic {
				err := runtime.AddFunctionsPipeline("per-topic", []string{"#"}, []interfaces.AppFunction{transformPassthru, test.TargetTransform})
				require.NoError(t, err)
				pipeline = runtime.GetPipelineById("per-topic")
				require.NotNil(t, pipeline)
			} else {
				runtime.SetDefaultFunctionsPipeline([]interfaces.AppFunction{transformPassthru, test.TargetTransform})
				pipeline = runtime.GetDefaultPipeline()
				require.NotNil(t, pipeline)
			}

			object := interfaces.NewStoredObject(serviceKey, payload, pipeline.Id, 1, pipeline.Hash, nil)
			object.CorrelationID = "CorrelationID"
			object.RetryCount = test.RetryCount

			_, _ = mockStoreObject(object)

			// Target of this test
			runtime.storeForward.retryStoredData(serviceKey)

			objects := mockRetrieveObjects(serviceKey)
			assert.Equal(t, int64(len(objects)), runtime.storeForward.dataCount.Count())

			if assert.Equal(t, test.ExpectedObjectCount, len(objects)) && test.ExpectedObjectCount > 0 {
				assert.Equal(t, test.ExpectedRetryCount, objects[0].RetryCount)
				assert.Equal(t, serviceKey, objects[0].AppServiceKey, "AppServiceKey not as expected")
				assert.Equal(t, object.CorrelationID, objects[0].CorrelationID, "CorrelationID not as expected")
			}

		})
	}
}

func TestStoreForLaterRetry(t *testing.T) {
	payload := []byte("My Payload")

	pipeline := &interfaces.FunctionPipeline{
		Id:   "pipeline.Id",
		Hash: "pipeline.Hash",
	}

	updateDicWithMockStoreClient()
	ctx := appfunction.NewContext(uuid.NewString(), dic, common2.ContentTypeJSON)
	runtime := NewFunctionPipelineRuntime(serviceKey, nil, dic)
	assert.Equal(t, int64(0), runtime.storeForward.dataCount.Count())
	runtime.storeForward.storeForLaterRetry(payload, ctx, pipeline, 0)
	assert.Equal(t, int64(1), runtime.storeForward.dataCount.Count())
}

func TestTriggerRetry(t *testing.T) {
	mockLogger := &loggerMocks.LoggingClient{}
	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return mockLogger
		},
	})
	updateDicWithMockStoreClient()
	runtime := NewFunctionPipelineRuntime(serviceKey, nil, dic)

	// Run test with data count at default of 0
	runtime.storeForward.triggerRetry()
	mockLogger.AssertNotCalled(t, "Debugf", mock.Anything, mock.Anything)

	// Run test with data count at 1 to verify retry code is executed
	runtime.storeForward.dataCount.Inc(1)
	mockLogger.On("Debugf", mock.Anything, mock.Anything)
	mockLogger.On("Debug", "Triggering Store and Forward retry of failed data")
	runtime.storeForward.triggerRetry()
	mockLogger.AssertCalled(t, "Debugf", mock.Anything, mock.Anything)
}

var mockObjectStore map[string]interfaces.StoredObject

func updateDicWithMockStoreClient() *di.Container {
	mockObjectStore = make(map[string]interfaces.StoredObject)
	storeClient := &mocks.StoreClient{}
	storeClient.Mock.On("Store", mock.Anything).Return(mockStoreObject)
	storeClient.Mock.On("RemoveFromStore", mock.Anything).Return(mockRemoveObject)
	storeClient.Mock.On("Update", mock.Anything).Return(mockUpdateObject)
	storeClient.Mock.On("RetrieveFromStore", mock.Anything).Return(mockRetrieveObjects, nil)

	dic.Update(di.ServiceConstructorMap{
		container.StoreClientName: func(get di.Get) interface{} {
			return storeClient
		},
	})

	return dic
}

func mockStoreObject(object interfaces.StoredObject) (string, error) {
	if err := validateContract(false, object); err != nil {
		return "", err
	}

	if object.ID == "" {
		object.ID = uuid.New().String()
	}

	mockObjectStore[object.ID] = object

	return object.ID, nil
}

func mockUpdateObject(object interfaces.StoredObject) error {

	if err := validateContract(true, object); err != nil {
		return err
	}

	mockObjectStore[object.ID] = object
	return nil
}

func mockRemoveObject(object interfaces.StoredObject) error {
	if err := validateContract(true, object); err != nil {
		return err
	}

	delete(mockObjectStore, object.ID)
	return nil
}

func mockRetrieveObjects(serviceKey string) []interfaces.StoredObject {
	var objects []interfaces.StoredObject
	for _, item := range mockObjectStore {
		if item.AppServiceKey == serviceKey {
			objects = append(objects, item)
		}
	}

	return objects
}

// TODO remove this and use verify func on StoredObject when it is available
func validateContract(IDRequired bool, o interfaces.StoredObject) error {
	if IDRequired {
		if o.ID == "" {
			return errors.New("invalid contract, ID cannot be empty")
		}
	}
	if o.AppServiceKey == "" {
		return errors.New("invalid contract, app service key cannot be empty")
	}
	if len(o.Payload) == 0 {
		return errors.New("invalid contract, payload cannot be empty")
	}
	if o.Version == "" {
		return errors.New("invalid contract, version cannot be empty")
	}

	return nil
}
