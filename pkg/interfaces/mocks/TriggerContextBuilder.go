// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	interfaces "github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	mock "github.com/stretchr/testify/mock"

	types "github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// TriggerContextBuilder is an autogenerated mock type for the TriggerContextBuilder type
type TriggerContextBuilder struct {
	mock.Mock
}

// Execute provides a mock function with given fields: env
func (_m *TriggerContextBuilder) Execute(env types.MessageEnvelope) interfaces.AppFunctionContext {
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
