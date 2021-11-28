// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// TriggerRouteManager is an autogenerated mock type for the TriggerRouteManager type
type TriggerRouteManager struct {
	mock.Mock
}

// SetupTriggerRoute provides a mock function with given fields: path, handlerForTrigger
func (_m *TriggerRouteManager) SetupTriggerRoute(path string, handlerForTrigger func(http.ResponseWriter, *http.Request)) {
	_m.Called(path, handlerForTrigger)
}
