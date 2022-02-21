//
// Copyright (c) 2022 One Track Consulting
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

package app

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/redis"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func Test_service_RegisterCustomStoreFactory_Inbuilt_Name(t *testing.T) {
	sut := &Service{}

	err := sut.RegisterCustomStoreFactory(db.RedisDB, nil)

	require.Error(t, err)
}

func Test_service_RegisterCustomStoreFactory(t *testing.T) {
	sut := &Service{}

	f := func(info interfaces.DatabaseInfo, credentials config.Credentials) (interfaces.StoreClient, error) {
		return nil, nil
	}

	name := uuid.NewString()

	err := sut.RegisterCustomStoreFactory(name, f)

	assert.NoError(t, err)

	_, found := sut.customStoreClientFactories[strings.ToUpper(name)]
	assert.True(t, found)
}

func Test_service_NewStoreClient_Invalid(t *testing.T) {
	sut := &Service{}

	_, err := sut.createStoreClient(interfaces.DatabaseInfo{}, config.Credentials{})

	assert.Error(t, err)
}

func Test_service_NewStoreClient_Redis(t *testing.T) {
	sut := &Service{customStoreClientFactories: map[string]func(db interfaces.DatabaseInfo, cred config.Credentials) (interfaces.StoreClient, error){}}

	result, err := sut.createStoreClient(interfaces.DatabaseInfo{Type: db.RedisDB, Timeout: "5s"}, config.Credentials{})

	assert.NoError(t, err)
	c, ok := result.(*redis.Client)
	assert.NotNil(t, c)
	assert.True(t, ok)
}

func Test_service_NewStoreClient_Custom(t *testing.T) {
	var x interfaces.StoreClient

	sut := &Service{}

	f := func(info interfaces.DatabaseInfo, credentials config.Credentials) (interfaces.StoreClient, error) {
		return x, nil
	}

	name := uuid.NewString()

	err := sut.RegisterCustomStoreFactory(name, f)

	require.NoError(t, err)

	result, err := sut.createStoreClient(interfaces.DatabaseInfo{Type: name}, config.Credentials{})

	assert.NoError(t, err)
	assert.Equal(t, x, result)
}
