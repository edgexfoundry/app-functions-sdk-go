// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	appfunction "github.com/edgexfoundry/app-functions-sdk-go/v4/internal/appfunction"
	common "github.com/edgexfoundry/app-functions-sdk-go/v4/internal/common"

	interfaces "github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"

	logger "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	messaging "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/messaging"

	mock "github.com/stretchr/testify/mock"

	runtime "github.com/edgexfoundry/app-functions-sdk-go/v4/internal/runtime"

	types "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// ServiceBinding is an autogenerated mock type for the ServiceBinding type
type ServiceBinding struct {
	mock.Mock
}

// BuildContext provides a mock function with given fields: env
func (_m *ServiceBinding) BuildContext(env types.MessageEnvelope) interfaces.AppFunctionContext {
	ret := _m.Called(env)

	var r0 interfaces.AppFunctionContext
	if rf, ok := ret.Get(0).(func(types.MessageEnvelope) interfaces.AppFunctionContext); ok {
		r0 = rf(env)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.AppFunctionContext)
		}
	}

	return r0
}

// Config provides a mock function with given fields:
func (_m *ServiceBinding) Config() *common.ConfigurationStruct {
	ret := _m.Called()

	var r0 *common.ConfigurationStruct
	if rf, ok := ret.Get(0).(func() *common.ConfigurationStruct); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*common.ConfigurationStruct)
		}
	}

	return r0
}

// DecodeMessage provides a mock function with given fields: appContext, envelope
func (_m *ServiceBinding) DecodeMessage(appContext *appfunction.Context, envelope types.MessageEnvelope) (interface{}, *runtime.MessageError, bool) {
	ret := _m.Called(appContext, envelope)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(*appfunction.Context, types.MessageEnvelope) interface{}); ok {
		r0 = rf(appContext, envelope)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 *runtime.MessageError
	if rf, ok := ret.Get(1).(func(*appfunction.Context, types.MessageEnvelope) *runtime.MessageError); ok {
		r1 = rf(appContext, envelope)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*runtime.MessageError)
		}
	}

	var r2 bool
	if rf, ok := ret.Get(2).(func(*appfunction.Context, types.MessageEnvelope) bool); ok {
		r2 = rf(appContext, envelope)
	} else {
		r2 = ret.Get(2).(bool)
	}

	return r0, r1, r2
}

// GetDefaultPipeline provides a mock function with given fields:
func (_m *ServiceBinding) GetDefaultPipeline() *interfaces.FunctionPipeline {
	ret := _m.Called()

	var r0 *interfaces.FunctionPipeline
	if rf, ok := ret.Get(0).(func() *interfaces.FunctionPipeline); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*interfaces.FunctionPipeline)
		}
	}

	return r0
}

// GetMatchingPipelines provides a mock function with given fields: incomingTopic
func (_m *ServiceBinding) GetMatchingPipelines(incomingTopic string) []*interfaces.FunctionPipeline {
	ret := _m.Called(incomingTopic)

	var r0 []*interfaces.FunctionPipeline
	if rf, ok := ret.Get(0).(func(string) []*interfaces.FunctionPipeline); ok {
		r0 = rf(incomingTopic)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*interfaces.FunctionPipeline)
		}
	}

	return r0
}

// LoadCustomConfig provides a mock function with given fields: config, sectionName
func (_m *ServiceBinding) LoadCustomConfig(config interfaces.UpdatableConfig, sectionName string) error {
	ret := _m.Called(config, sectionName)

	var r0 error
	if rf, ok := ret.Get(0).(func(interfaces.UpdatableConfig, string) error); ok {
		r0 = rf(config, sectionName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// LoggingClient provides a mock function with given fields:
func (_m *ServiceBinding) LoggingClient() logger.LoggingClient {
	ret := _m.Called()

	var r0 logger.LoggingClient
	if rf, ok := ret.Get(0).(func() logger.LoggingClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(logger.LoggingClient)
		}
	}

	return r0
}

// ProcessMessage provides a mock function with given fields: appContext, data, pipeline
func (_m *ServiceBinding) ProcessMessage(appContext *appfunction.Context, data interface{}, pipeline *interfaces.FunctionPipeline) *runtime.MessageError {
	ret := _m.Called(appContext, data, pipeline)

	var r0 *runtime.MessageError
	if rf, ok := ret.Get(0).(func(*appfunction.Context, interface{}, *interfaces.FunctionPipeline) *runtime.MessageError); ok {
		r0 = rf(appContext, data, pipeline)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*runtime.MessageError)
		}
	}

	return r0
}

// SecretProvider provides a mock function with given fields:
func (_m *ServiceBinding) SecretProvider() messaging.SecretDataProvider {
	ret := _m.Called()

	var r0 messaging.SecretDataProvider
	if rf, ok := ret.Get(0).(func() messaging.SecretDataProvider); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(messaging.SecretDataProvider)
		}
	}

	return r0
}

type mockConstructorTestingTNewServiceBinding interface {
	mock.TestingT
	Cleanup(func())
}

// NewServiceBinding creates a new instance of ServiceBinding. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewServiceBinding(t mockConstructorTestingTNewServiceBinding) *ServiceBinding {
	mock := &ServiceBinding{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
