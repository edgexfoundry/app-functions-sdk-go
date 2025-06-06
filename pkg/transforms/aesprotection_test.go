//
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

package transforms

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/etm"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces/mocks"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/util"
	bootstrapMocks "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	plainString = "This is the test string used for testing"
)

func TestNewAESProtection(t *testing.T) {
	secretName := uuid.NewString()
	secretValueKey := uuid.NewString()

	sut := NewAESProtection(secretName, secretValueKey)

	assert.Equal(t, secretName, sut.SecretName)
	assert.Equal(t, secretValueKey, sut.SecretValueKey)
}

func TestAESProtection_clearKey(t *testing.T) {
	key := []byte(uuid.NewString())

	clearKey(key)

	for _, v := range key {
		assert.Equal(t, byte(0), v)
	}
}

func TestAESProtection_getKey(t *testing.T) {
	secretName := uuid.NewString()
	secretValueKey := uuid.NewString()
	pipelineId := uuid.NewString()
	key := "217A24432646294A404E635266556A586E3272357538782F413F442A472D4B6150645367566B59703373367639792442264529482B4D6251655468576D5A7134"

	type fields struct {
		SecretName     string
		SecretValueKey string
		EncryptionKey  string
	}
	tests := []struct {
		name     string
		fields   fields
		ctxSetup func(ctx *mocks.AppFunctionContext, secretProvider *bootstrapMocks.SecretProvider)
		wantErr  bool
	}{
		{name: "no key", wantErr: true},
		{
			name:   "secret error",
			fields: fields{SecretName: secretName, SecretValueKey: secretValueKey},
			ctxSetup: func(ctx *mocks.AppFunctionContext, secretProvider *bootstrapMocks.SecretProvider) {
				secretProvider.On("GetSecret", secretName, secretValueKey).Return(nil, fmt.Errorf("secret error"))
				ctx.On("SecretProvider").Return(secretProvider)
			},
			wantErr: true,
		},
		{
			name:   "secret not in map",
			fields: fields{SecretName: secretName, SecretValueKey: secretValueKey},
			ctxSetup: func(ctx *mocks.AppFunctionContext, secretProvider *bootstrapMocks.SecretProvider) {
				secretProvider.On("GetSecret", secretName, secretValueKey).Return(map[string]string{}, nil)
				ctx.On("SecretProvider").Return(secretProvider)
			},
			wantErr: true,
		},
		{
			name:   "happy",
			fields: fields{SecretName: secretName, SecretValueKey: secretValueKey},
			ctxSetup: func(ctx *mocks.AppFunctionContext, secretProvider *bootstrapMocks.SecretProvider) {
				ctx.On("SetResponsesContentType", common.ContentTypeText).Return()
				secretProvider.On("GetSecret", secretName, secretValueKey).Return(map[string]string{secretValueKey: key}, nil)
				ctx.On("SecretProvider").Return(secretProvider)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aesData := &AESProtection{
				SecretName:     tt.fields.SecretName,
				SecretValueKey: tt.fields.SecretValueKey,
				EncryptionKey:  tt.fields.EncryptionKey,
			}

			mockSecretProvider := &bootstrapMocks.SecretProvider{}
			ctx := &mocks.AppFunctionContext{}
			ctx.On("PipelineId").Return(pipelineId)
			ctx.On("LoggingClient").Return(logger.NewMockClient())

			if tt.ctxSetup != nil {
				tt.ctxSetup(ctx, mockSecretProvider)
			}

			if k, err := aesData.getKey(ctx); (err != nil) != tt.wantErr {
				t.Errorf("getKey() error = %v, wantErr %v", err, tt.wantErr)

				if !tt.wantErr {
					assert.Equal(t, key, k)
				}
			}
		})
	}
}

func TestAESProtection_Encrypt(t *testing.T) {
	secretName := uuid.NewString()
	secretValueKey := uuid.NewString()
	key := "217A24432646294A404E635266556A586E3272357538782F413F442A472D4B6150645367566B59703373367639792442264529482B4D6251655468576D5A7134"

	mockSecretProvider := &bootstrapMocks.SecretProvider{}
	mockSecretProvider.On("GetSecret", secretName, secretValueKey).Return(map[string]string{secretValueKey: key}, nil)
	ctx := &mocks.AppFunctionContext{}
	ctx.On("SetResponseContentType", common.ContentTypeText).Return()
	ctx.On("PipelineId").Return("pipeline-id")
	ctx.On("LoggingClient").Return(logger.NewMockClient())
	ctx.On("SecretProvider").Return(mockSecretProvider)

	enc := NewAESProtection(secretName, secretValueKey)

	continuePipeline, encrypted := enc.Encrypt(ctx, []byte(plainString))
	assert.True(t, continuePipeline)

	ebytes, err := util.CoerceType(encrypted)

	require.NoError(t, err)

	//output is base64 encoded
	dbytes, err := base64.StdEncoding.DecodeString(string(ebytes))

	if err != nil {
		panic(err)
	}

	decrypted := aes256Decrypt(t, dbytes, key)

	assert.Equal(t, plainString, string(decrypted))
}

func aes256Decrypt(t *testing.T, dbytes []byte, key string) []byte {
	k, err := hex.DecodeString(key)

	if err != nil {
		panic(err)
	}

	//internally we are leaning heavily on ETM logic
	//do not want to re-implement here
	etm, err := etm.NewAES256SHA512(k)

	require.NoError(t, err)

	dst := make([]byte, 0)

	res, err := etm.Open(dst, nil, dbytes, nil)

	require.NoError(t, err)

	return res
}
