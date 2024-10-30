//
// Copyright (c) 2020 Technotects
// Copyright (c) 2022 Intel Corporation
// Copyright (c) 2021 One Track Consulting
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

package http

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/trigger/http/mocks"
	triggerMocks "github.com/edgexfoundry/app-functions-sdk-go/v4/internal/trigger/mocks"
	interfaceMocks "github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces/mocks"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/stretchr/testify/assert"
)

func TestTriggerInitializeWithBackgroundChannel(t *testing.T) {
	background := make(chan interfaces.BackgroundMessage)

	bnd := &triggerMocks.ServiceBinding{}
	bnd.On("LoggingClient").Return(logger.NewMockClient())

	trigger := NewTrigger(bnd, nil, nil)

	deferred, err := trigger.Initialize(nil, nil, background)

	assert.Nil(t, deferred)
	assert.Error(t, err)
	assert.Equal(t, "background publishing not supported for services using HTTP trigger", err.Error())
}

func TestTriggerInitialize(t *testing.T) {
	bnd := &triggerMocks.ServiceBinding{}
	bnd.On("LoggingClient").Return(logger.NewMockClient())

	trm := &mocks.TriggerRouteManager{}
	trm.On("SetupTriggerRoute", internal.ApiTriggerRoute, mock.AnythingOfType("func(http.ResponseWriter, *http.Request)"))
	defer trm.AssertExpectations(t)

	trigger := NewTrigger(bnd, nil, trm)

	deferred, err := trigger.Initialize(nil, nil, nil)

	assert.Nil(t, deferred)
	assert.NoError(t, err)
}

func TestTriggerRequestHandler_BodyReadError(t *testing.T) {
	bnd := &triggerMocks.ServiceBinding{}
	bnd.On("LoggingClient").Return(logger.NewMockClient())

	trigger := Trigger{
		serviceBinding: bnd,
	}

	errorMsg := "fake error"

	writer := &mocks.TriggerResponseWriter{}
	writer.On("WriteHeader", http.StatusBadRequest)
	writer.On("Write", mock.Anything).Return(0, nil)
	defer writer.AssertExpectations(t)

	reqReader := &mocks.TriggerRequestReader{}
	reqReader.On("Read", mock.Anything).Return(0, errors.New(errorMsg))

	req, err := http.NewRequest("", "", reqReader)
	req.Header = http.Header{}

	require.NoError(t, err)

	trigger.requestHandler(writer, req)

	writer.AssertExpectations(t)
}

func TestTriggerRequestHandler_ProcessError(t *testing.T) {
	data := []byte("some data")
	contentType := "arbitrary string"
	correlationId := uuid.NewString()
	errCode := 47
	afc := appfunction.NewContext(correlationId, nil, contentType) // &interfaceMocks.AppFunctionContext{}
	pipeline := &interfaces.FunctionPipeline{}

	bnd := &triggerMocks.ServiceBinding{}
	bnd.On("LoggingClient").Return(logger.NewMockClient())
	bnd.On("BuildContext", mock.AnythingOfType("types.MessageEnvelope")).Return(afc)
	bnd.On("GetDefaultPipeline").Return(pipeline)
	bnd.On("DecodeMessage", mock.Anything, mock.Anything).Return(data, nil, false)
	bnd.On("ProcessMessage", afc, mock.Anything, pipeline).Return(func(ctx *appfunction.Context, messageData interface{}, p *interfaces.FunctionPipeline) *runtime.MessageError {
		assert.Equal(t, correlationId, ctx.CorrelationID())
		assert.Equal(t, afc, ctx)
		assert.Equal(t, data, messageData)
		assert.Equal(t, contentType, ctx.InputContentType())
		return &runtime.MessageError{
			Err:       fmt.Errorf("error"),
			ErrorCode: errCode,
		}
	})

	trigger := Trigger{
		serviceBinding:   bnd,
		messageProcessor: &triggerMocks.MessageProcessor{},
	}

	writer := &mocks.TriggerResponseWriter{}
	writer.On("WriteHeader", errCode)
	writer.On("Write", []byte("error")).Return(0, nil)

	req, err := http.NewRequest("", "", bytes.NewBuffer(data))
	req.Header = http.Header{}
	req.Header.Add(common.ContentType, contentType)
	req.Header.Add(common.CorrelationHeader, correlationId)

	require.NoError(t, err)

	trigger.requestHandler(writer, req)

	writer.AssertExpectations(t)
}

func TestTriggerRequestHandler(t *testing.T) {
	data := []byte("some data")
	contentType := "arbitrary string"
	correlationId := uuid.NewString()
	afc := appfunction.NewContext(correlationId, nil, contentType) // &interfaceMocks.AppFunctionContext{}
	pipeline := &interfaces.FunctionPipeline{}

	bnd := &triggerMocks.ServiceBinding{}
	bnd.On("LoggingClient").Return(logger.NewMockClient())
	bnd.On("BuildContext", mock.AnythingOfType("types.MessageEnvelope")).Return(afc)
	bnd.On("GetDefaultPipeline").Return(pipeline)
	bnd.On("DecodeMessage", afc, mock.Anything).Return(data, nil, false)
	bnd.On("ProcessMessage", afc, mock.Anything, pipeline).Return(func(ctx *appfunction.Context, messageData interface{}, p *interfaces.FunctionPipeline) *runtime.MessageError {
		assert.Equal(t, correlationId, ctx.CorrelationID())
		assert.Equal(t, afc, ctx)
		assert.Equal(t, data, messageData)
		assert.Equal(t, contentType, ctx.InputContentType())
		ctx.SetResponseData(data)
		return nil
	})

	trigger := Trigger{
		serviceBinding:   bnd,
		messageProcessor: &triggerMocks.MessageProcessor{},
	}

	writer := &mocks.TriggerResponseWriter{}

	writer.On("Write", data).Return(0, nil)

	req, err := http.NewRequest("", "", bytes.NewBuffer(data))
	req.Header = http.Header{}
	req.Header.Add(common.ContentType, contentType)
	req.Header.Add(common.CorrelationHeader, correlationId)

	require.NoError(t, err)

	trigger.requestHandler(writer, req)
}

func Test_getResponseHandler(t *testing.T) {
	data := []byte("some data in response")
	correlationId := uuid.NewString()

	type inputs struct {
		correlationId string
		contentType   string
		data          []byte
		pipeline      *interfaces.FunctionPipeline
		writerHeader  http.Header
	}
	tests := []struct {
		name    string
		inputs  inputs
		setup   func(writer *mocks.TriggerResponseWriter, ctx *interfaceMocks.AppFunctionContext, i inputs)
		wantErr bool
	}{
		{name: "write error", inputs: inputs{pipeline: &interfaces.FunctionPipeline{}, correlationId: uuid.NewString(), data: []byte("some data in response")}, setup: func(writer *mocks.TriggerResponseWriter, ctx *interfaceMocks.AppFunctionContext, ip inputs) {
			ctx.On("ResponseContentType").Return(ip.contentType)
			ctx.On("ResponseData").Return(data)
			writer.On("Write", data).Return(0, fmt.Errorf("write error"))
		}, wantErr: true},
		{name: "happy no content type", inputs: inputs{pipeline: &interfaces.FunctionPipeline{}, correlationId: uuid.NewString(), data: []byte("some data in response")}, setup: func(writer *mocks.TriggerResponseWriter, ctx *interfaceMocks.AppFunctionContext, ip inputs) {
			ctx.On("CorrelationID").Return(correlationId)
			ctx.On("ResponseContentType").Return(ip.contentType)
			ctx.On("ResponseData").Return(data)
			writer.On("Write", data).Return(5, nil)
		}, wantErr: false},
		{name: "happy", inputs: inputs{pipeline: &interfaces.FunctionPipeline{}, correlationId: uuid.NewString(), data: []byte("some data in response")}, setup: func(writer *mocks.TriggerResponseWriter, ctx *interfaceMocks.AppFunctionContext, ip inputs) {
			ctx.On("CorrelationID").Return(correlationId)
			ctx.On("ResponseContentType").Return(ip.contentType)
			ctx.On("ResponseData").Return(data)
			writer.On("Write", data).Return(5, nil)
			writer.On("Header").Return(ip.writerHeader)
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &mocks.TriggerResponseWriter{}
			ctx := &interfaceMocks.AppFunctionContext{}

			tt.setup(writer, ctx, tt.inputs)

			err := getResponseHandler(writer, logger.NewMockClient())(ctx, tt.inputs.pipeline)

			assert.Equal(t, tt.wantErr, err != nil)

			assert.Equal(t, tt.inputs.contentType, tt.inputs.writerHeader.Get(common.ContentType))
		})
	}
}
