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
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/etm"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/util"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
)

type AESProtection struct {
	SecretName     string
	SecretValueKey string
	EncryptionKey  string
}

// NewAESProtection creates, initializes and returns a new instance of AESProtection configured
// to retrieve the encryption key from the Secret Store
func NewAESProtection(secretName string, secretValueKey string) *AESProtection {
	return &AESProtection{
		SecretName:     secretName,
		SecretValueKey: secretValueKey,
	}
}

// Encrypt encrypts a string, []byte, or json.Marshaller type using AES 256 encryption.
// It also signs the data using a SHA512 hash.
// It will return a Base64 encode []byte of the encrypted data.
func (protection *AESProtection) Encrypt(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, fmt.Errorf("function Encrypt in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	ctx.LoggingClient().Debugf("Encrypting with AES256 in pipeline '%s'", ctx.PipelineId())

	byteData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}

	key, err := protection.getKey(ctx)

	if err != nil {
		return false, err
	}

	if len(key) == 0 {
		return false, fmt.Errorf("AES256 encryption key not set in pipeline '%s'", ctx.PipelineId())
	}

	aead, err := etm.NewAES256SHA512(key)

	if err != nil {
		return false, err
	}

	nonce := make([]byte, aead.NonceSize())
	_, err = rand.Read(nonce)

	if err != nil {
		return false, err
	}

	dst := make([]byte, 0)

	encrypted := aead.Seal(dst, nonce, byteData, nil)

	clearKey(key)

	encodedData := []byte(base64.StdEncoding.EncodeToString(encrypted))

	// Set response "content-type" header to "text/plain"
	ctx.SetResponseContentType(common.ContentTypeText)

	return true, encodedData
}

func (protection *AESProtection) getKey(ctx interfaces.AppFunctionContext) ([]byte, error) {
	// If using Secret Store for the encryption key
	if len(protection.SecretName) != 0 && len(protection.SecretValueKey) != 0 {
		// Note secrets are cached so this call doesn't result in unneeded calls to SecretStore Service and
		// the cache is invalidated when StoreSecrets is used.
		secretData, err := ctx.SecretProvider().GetSecret(protection.SecretName, protection.SecretValueKey)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to retieve encryption key at SecretName=%s and SecretValueKey=%s in pipeline '%s'",
				protection.SecretName,
				protection.SecretValueKey,
				ctx.PipelineId())
		}

		key, ok := secretData[protection.SecretValueKey]
		if !ok {
			return nil, fmt.Errorf(
				"unable find encryption key in secret data for name=%s in pipeline '%s'",
				protection.SecretValueKey,
				ctx.PipelineId())
		}

		ctx.LoggingClient().Debugf(
			"Using encryption key from Secret Store at SecretName=%s & SecretValueKey=%s in pipeline '%s'",
			protection.SecretName,
			protection.SecretValueKey,
			ctx.PipelineId())

		return hex.DecodeString(key)
	}
	return nil, fmt.Errorf("no key configured")
}

func clearKey(key []byte) {
	for i := range key {
		key[i] = 0
	}
}
