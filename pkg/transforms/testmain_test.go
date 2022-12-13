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

package transforms

import (
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces/mocks"
	commonDtos "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

var lc logger.LoggingClient
var dic *di.Container
var ctx *appfunction.Context
var mockEventClient *mocks.EventClient

func TestMain(m *testing.M) {
	lc = logger.NewMockClient()

	mockEventClient = &mocks.EventClient{}
	mockEventClient.On("Add", mock.Anything, mock.Anything).Return(commonDtos.BaseWithIdResponse{}, nil)

	dic = di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &common.ConfigurationStruct{}
		},
		bootstrapContainer.EventClientName: func(get di.Get) interface{} {
			return mockEventClient
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	ctx = appfunction.NewContext("123", dic, "")

	os.Exit(m.Run())
}
