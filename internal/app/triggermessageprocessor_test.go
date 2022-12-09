//
// Copyright (c) 2021 One Track Consulting
// Copyright (c) 2022 Intel Corporation
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

package app

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	gometrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/runtime"
	triggerMocks "github.com/edgexfoundry/app-functions-sdk-go/v3/internal/trigger/mocks"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
)

func Test_simpleTriggerServiceBinding_BuildContext(t *testing.T) {
	container := &di.Container{}
	correlationId := uuid.NewString()
	contentType := uuid.NewString()

	bnd := &simpleTriggerServiceBinding{&Service{dic: container}, nil}

	got := bnd.BuildContext(types.MessageEnvelope{CorrelationID: correlationId, ContentType: contentType})

	require.NotNil(t, got)

	assert.Equal(t, correlationId, got.CorrelationID())
	assert.Equal(t, contentType, got.InputContentType())

	ctx, ok := got.(*appfunction.Context)
	require.True(t, ok)
	assert.Equal(t, container, ctx.Dic)
}

func Test_triggerMessageProcessor_MessageReceived(t *testing.T) {
	type returns struct {
		runtimeProcessor interface{}
		pipelineMatcher  interface{}
	}
	type args struct {
		ctx      interfaces.AppFunctionContext
		envelope types.MessageEnvelope
	}
	errorPipeline := testPipeline()
	errorPipeline.Id = "errorid"

	tests := []struct {
		name    string
		setup   returns
		args    args
		nilRh   bool
		wantErr int
	}{
		{
			name:    "no pipelines",
			setup:   returns{},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 0,
		},
		{
			name: "single pipeline",
			setup: returns{
				pipelineMatcher:  []*interfaces.FunctionPipeline{testPipeline()},
				runtimeProcessor: nil,
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 0,
		},
		{
			name: "single pipeline error",
			setup: returns{
				pipelineMatcher:  []*interfaces.FunctionPipeline{testPipeline()},
				runtimeProcessor: &runtime.MessageError{Err: fmt.Errorf("some error")},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 1,
		},
		{
			name: "multi pipeline",
			setup: returns{
				pipelineMatcher: []*interfaces.FunctionPipeline{testPipeline(), testPipeline(), testPipeline()},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 0,
		},
		{
			name: "multi pipeline single err",
			setup: returns{
				pipelineMatcher: []*interfaces.FunctionPipeline{testPipeline(), errorPipeline, testPipeline()},
				runtimeProcessor: func(appContext *appfunction.Context, data interface{}, pipeline *interfaces.FunctionPipeline) *runtime.MessageError {
					if pipeline.Id == "errorid" {
						return &runtime.MessageError{Err: fmt.Errorf("new error")}
					}
					return nil
				},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 1,
		},
		{
			name: "multi pipeline multi err",
			setup: returns{
				pipelineMatcher:  []*interfaces.FunctionPipeline{testPipeline(), testPipeline(), testPipeline()},
				runtimeProcessor: &runtime.MessageError{Err: fmt.Errorf("new error")},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tsb := triggerMocks.ServiceBinding{}

			tsb.On("ProcessMessage", mock.Anything, mock.Anything, mock.Anything).Return(tt.setup.runtimeProcessor)
			tsb.On("DecodeMessage", mock.Anything, mock.Anything).Return(nil, nil, false)
			tsb.On("GetMatchingPipelines", tt.args.envelope.ReceivedTopic).Return(tt.setup.pipelineMatcher)
			tsb.On("LoggingClient").Return(lc)

			mmMock := mocks.MetricsManager{}
			mmMock.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			bnd := NewTriggerMessageProcessor(&tsb, &mmMock)

			var rh interfaces.PipelineResponseHandler

			if !tt.nilRh {
				rh = func(ctx interfaces.AppFunctionContext, pipeline *interfaces.FunctionPipeline) error {
					assert.NotEqual(t, tt.args.ctx, ctx)
					assert.Equal(t, tt.args.ctx.CorrelationID(), ctx.CorrelationID()) //hmm
					return nil
				}
			}

			err := bnd.MessageReceived(tt.args.ctx, tt.args.envelope, rh)

			require.Equal(t, err == nil, tt.wantErr == 0)

			if err != nil {
				if merr, ok := err.(*multierror.Error); ok {
					assert.Equal(t, tt.wantErr, merr.Len())
				}
			}
		})
	}
}

func testPipeline() *interfaces.FunctionPipeline {
	return &interfaces.FunctionPipeline{
		MessagesProcessed:     gometrics.NewCounter(),
		MessageProcessingTime: gometrics.NewTimer(),
		ProcessingErrors:      gometrics.NewCounter(),
	}
}
