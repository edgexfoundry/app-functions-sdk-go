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

package app

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	"github.com/stretchr/testify/assert"
)

func TestFilterByProfileName(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameters", map[string]string{"": ""}, true},
		{"Empty Parameters", map[string]string{ProfileNames: ""}, false},
		{"Valid Parameters", map[string]string{ProfileNames: "GS1-AC-Drive, GS0-DC-Drive, GSX-ACDC-Drive"}, false},
		{"Empty FilterOut Parameters", map[string]string{ProfileNames: "GS1-AC-Drive, GS0-DC-Drive, GSX-ACDC-Drive", FilterOut: ""}, true},
		{"Valid FilterOut Parameters", map[string]string{ProfileNames: "GS1-AC-Drive, GS0-DC-Drive, GSX-ACDC-Drive", FilterOut: "true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.FilterByProfileName(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from FilterByProfileName should be nil")
			} else {
				assert.NotNil(t, trx, "return result from FilterByProfileName should not be nil")
			}
		})
	}
}

func TestFilterByDeviceName(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameters", map[string]string{"": ""}, true},
		{"Empty Parameters", map[string]string{DeviceNames: ""}, false},
		{"Valid Parameters", map[string]string{DeviceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03"}, false},
		{"Empty FilterOut Parameters", map[string]string{DeviceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: ""}, true},
		{"Valid FilterOut Parameters", map[string]string{DeviceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: "true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.FilterByDeviceName(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from FilterByDeviceName should be nil")
			} else {
				assert.NotNil(t, trx, "return result from FilterByDeviceName should not be nil")
			}
		})
	}
}

func TestFilterBySourceName(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameters", map[string]string{"": ""}, true},
		{"Empty Parameters", map[string]string{SourceNames: ""}, false},
		{"Valid Parameters", map[string]string{SourceNames: "R1, C2, R4"}, false},
		{"Empty FilterOut Parameters", map[string]string{SourceNames: "R1, C2, R4", FilterOut: ""}, true},
		{"Valid FilterOut Parameters", map[string]string{SourceNames: "R1, C2, R4", FilterOut: "true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.FilterBySourceName(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from FilterBySourceName should be nil")
			} else {
				assert.NotNil(t, trx, "return result from FilterBySourceName should not be nil")
			}
		})
	}
}

func TestFilterByResourceName(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameters", map[string]string{"": ""}, true},
		{"Empty Parameters", map[string]string{ResourceNames: ""}, false},
		{"Valid Parameters", map[string]string{ResourceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03"}, false},
		{"Empty FilterOut Parameters", map[string]string{ResourceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: ""}, true},
		{"Valid FilterOut Parameters", map[string]string{ResourceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: "true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.FilterByResourceName(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from FilterByResourceName should be nil")
			} else {
				assert.NotNil(t, trx, "return result from FilterByResourceName should not be nil")
			}
		})
	}
}

func TestTransform(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		Name          string
		TransformType string
		ExpectValid   bool
	}{
		{"Good - XML", "xMl", true},
		{"Good - JSON", "JsOn", true},
		{"Bad Type", "baDType", false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			params := make(map[string]string)
			params[TransformType] = test.TransformType
			transform := configurable.Transform(params)
			assert.Equal(t, test.ExpectValid, transform != nil)
		})
	}
}

func TestHTTPExport(t *testing.T) {
	configurable := Configurable{lc: lc}

	testUrl := "http://url"
	testMimeType := common.ContentTypeJSON
	testPersistOnError := "false"
	testBadPersistOnError := "bogus"
	testContinueOnSendError := "true"
	testBadContinueOnSendError := "bogus"
	testReturnInputData := "true"
	testBadReturnInputData := "bogus"

	testHeaderName := "My-Header"
	testSecretName := "my-secret"
	testSecretValueKey := "header"

	testHTTPRequestHeaders := `{
		"Connection": "keep-alive",
		"From": "[user@example.com](mailto:user@example.com)"
	  }`

	testBadHTTPRequestHeaders := `{
		"Connection": "keep-alive", 
		"From": 
	  `

	tests := []struct {
		Name                string
		Method              string
		Url                 *string
		MimeType            *string
		PersistOnError      *string
		ContinueOnSendError *string
		ReturnInputData     *string
		HeaderName          *string
		SecretName          *string
		SecretValueKey      *string
		HTTPRequestHeaders  *string
		ExpectValid         bool
	}{
		{"Valid Post - ony required params", ExportMethodPost, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, nil, true},
		{"Valid Post - w/o secrets", http.MethodPost, &testUrl, &testMimeType, &testPersistOnError, nil, nil, nil, nil, nil, nil, true},
		{"Valid Post - with secrets", ExportMethodPost, &testUrl, &testMimeType, nil, nil, nil, &testHeaderName, &testSecretName, &testSecretValueKey, nil, true},
		{"Valid Post - with http requet headers", ExportMethodPost, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, &testHTTPRequestHeaders, true},
		{"Valid Post - with all params", ExportMethodPost, &testUrl, &testMimeType, &testPersistOnError, &testContinueOnSendError, &testReturnInputData, &testHeaderName, &testSecretName, &testSecretValueKey, &testHTTPRequestHeaders, true},
		{"Invalid Post - no url", ExportMethodPost, nil, &testMimeType, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid Post - no mimeType", ExportMethodPost, &testUrl, nil, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid Post - bad persistOnError", ExportMethodPost, &testUrl, &testMimeType, &testBadPersistOnError, nil, nil, nil, nil, nil, nil, false},
		{"Invalid Post - missing headerName", ExportMethodPost, &testUrl, &testMimeType, &testPersistOnError, nil, nil, nil, &testSecretName, &testSecretValueKey, nil, false},
		{"Invalid Post - missing secretName", ExportMethodPost, &testUrl, &testMimeType, &testPersistOnError, nil, nil, &testHeaderName, nil, &testSecretValueKey, nil, false},
		{"Invalid Post - missing secretValueKey", ExportMethodPost, &testUrl, &testMimeType, &testPersistOnError, nil, nil, &testHeaderName, &testSecretName, nil, nil, false},
		{"Invalid Post - unmarshal error for http requet headers", ExportMethodPost, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, &testBadHTTPRequestHeaders, false},
		{"Valid Put - ony required params", ExportMethodPut, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, nil, true},
		{"Valid Put - w/o secrets", ExportMethodPut, &testUrl, &testMimeType, &testPersistOnError, nil, nil, nil, nil, nil, nil, true},
		{"Valid Put - with secrets", http.MethodPut, &testUrl, &testMimeType, nil, nil, nil, &testHeaderName, &testSecretName, &testSecretValueKey, nil, true},
		{"Valid Put - with http request headers", ExportMethodPut, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, &testHTTPRequestHeaders, true},
		{"Valid Put - with all params", ExportMethodPut, &testUrl, &testMimeType, &testPersistOnError, nil, nil, &testHeaderName, &testSecretName, &testSecretValueKey, &testHTTPRequestHeaders, true},
		{"Invalid Put - no url", ExportMethodPut, nil, &testMimeType, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid Put - no mimeType", ExportMethodPut, &testUrl, nil, nil, nil, nil, nil, nil, nil, nil, false},
		{"Invalid Put - bad persistOnError", ExportMethodPut, &testUrl, &testMimeType, &testBadPersistOnError, nil, nil, nil, nil, nil, nil, false},
		{"Invalid Put - bad continueOnSendError", ExportMethodPut, &testUrl, &testMimeType, nil, &testBadContinueOnSendError, nil, nil, nil, nil, nil, false},
		{"Invalid Put - bad returnInputData", ExportMethodPut, &testUrl, &testMimeType, nil, nil, &testBadReturnInputData, nil, nil, nil, nil, false},
		{"Invalid Put - missing headerName", ExportMethodPut, &testUrl, &testMimeType, &testPersistOnError, nil, nil, nil, &testSecretName, &testSecretValueKey, nil, false},
		{"Invalid Put - missing secretName", ExportMethodPut, &testUrl, &testMimeType, &testPersistOnError, nil, nil, &testHeaderName, nil, &testSecretValueKey, nil, false},
		{"Invalid Put - missing secretValueKey", ExportMethodPut, &testUrl, &testMimeType, &testPersistOnError, nil, nil, &testHeaderName, &testSecretName, nil, nil, false},
		{"Invalid Put - unmarshal error for http requet headers", ExportMethodPut, &testUrl, &testMimeType, nil, nil, nil, nil, nil, nil, &testBadHTTPRequestHeaders, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			params := make(map[string]string)
			params[ExportMethod] = test.Method

			if test.Url != nil {
				params[Url] = *test.Url
			}

			if test.MimeType != nil {
				params[MimeType] = *test.MimeType
			}

			if test.PersistOnError != nil {
				params[PersistOnError] = *test.PersistOnError
			}

			if test.ContinueOnSendError != nil {
				params[ContinueOnSendError] = *test.ContinueOnSendError
			}

			if test.ReturnInputData != nil {
				params[ReturnInputData] = *test.ReturnInputData
			}

			if test.HeaderName != nil {
				params[HeaderName] = *test.HeaderName
			}

			if test.SecretName != nil {
				params[SecretName] = *test.SecretName
			}

			if test.SecretValueKey != nil {
				params[SecretValueKey] = *test.SecretValueKey
			}

			if test.HTTPRequestHeaders != nil {
				params[HttpRequestHeaders] = *test.HTTPRequestHeaders
			}

			transform := configurable.HTTPExport(params)
			assert.Equal(t, test.ExpectValid, transform != nil)
		})
	}
}

func TestSetOutputData(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameter", map[string]string{}, false},
		{"Valid Parameter With Value", map[string]string{ResponseContentType: "application/json"}, false},
		{"Valid Parameter Without Value", map[string]string{ResponseContentType: ""}, false},
		{"Unknown Parameter", map[string]string{"Unknown": "scary/text"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.SetResponseData(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from SetResponseData should be nil")
			} else {
				assert.NotNil(t, trx, "return result from SetResponseData should not be nil")
			}
		})
	}
}

func TestBatchByCount(t *testing.T) {
	configurable := Configurable{lc: lc}

	params := make(map[string]string)
	params[Mode] = BatchByCount
	params[BatchThreshold] = "30"
	params[IsEventData] = "true"

	transform := configurable.Batch(params)
	assert.NotNil(t, transform, "return result for BatchByCount should not be nil")
}

func TestBatchByTime(t *testing.T) {
	configurable := Configurable{lc: lc}

	params := make(map[string]string)
	params[Mode] = BatchByTime
	params[TimeInterval] = "10s"
	params[IsEventData] = "false"

	transform := configurable.Batch(params)
	assert.NotNil(t, transform, "return result for BatchByTime should not be nil")
}

func TestBatchByTimeAndCount(t *testing.T) {
	configurable := Configurable{lc: lc}

	params := make(map[string]string)
	params[Mode] = BatchByTimeAndCount
	params[BatchThreshold] = "30"
	params[TimeInterval] = "10s"

	trx := configurable.Batch(params)
	assert.NotNil(t, trx, "return result for BatchByTimeAndCount should not be nil")
}

func TestJSONLogic(t *testing.T) {
	params := make(map[string]string)
	params[Rule] = "{}"

	configurable := Configurable{lc: lc}

	trx := configurable.JSONLogic(params)
	assert.NotNil(t, trx, "return result from JSONLogic should not be nil")

}

func TestMQTTExport(t *testing.T) {
	configurable := Configurable{lc: lc}

	params := make(map[string]string)
	params[BrokerAddress] = "mqtt://broker:8883"
	params[Topic] = "topic"
	params[SecretName] = "my-secret"
	params[ClientID] = "clientid"
	params[Qos] = "0"
	params[Retain] = "true"
	params[AutoReconnect] = "true"
	params[SkipVerify] = "true"
	params[PersistOnError] = "false"
	params[AuthMode] = "none"
	params[ConnectTimeout] = "5s"
	params[KeepAlive] = "6s"
	params[WillEnabled] = "true"
	params[WillTopic] = "will"
	params[WillPayload] = "goodbye"
	params[PreConnectRetryCount] = "10"
	params[PreConnectRetryInterval] = "6s"
	params[MaxReconnectInterval] = "10s"

	trx := configurable.MQTTExport(params)
	assert.NotNil(t, trx, "return result from MQTTSecretSend should not be nil")
}

func TestMQTTExportWillOptions(t *testing.T) {
	configurable := Configurable{lc: lc}

	goodWillOptions := make(map[string]string)
	goodWillOptions[WillEnabled] = "true"
	goodWillOptions[WillTopic] = "will"
	goodWillOptions[WillPayload] = "goodbye"
	goodWillOptions[WillRetained] = "true"
	goodWillOptions[WillQos] = "2"
	emptyWill := make(map[string]string)

	willDisabled := make(map[string]string)
	willDisabled[WillEnabled] = "false"

	badEnabled := make(map[string]string)
	badEnabled[WillEnabled] = "junk"

	missingTopic := copyMap(goodWillOptions)
	delete(missingTopic, WillTopic)
	emptyTopic := copyMap(goodWillOptions)
	emptyTopic[WillTopic] = ""

	missingPayload := copyMap(goodWillOptions)
	delete(missingPayload, WillPayload)
	emptyPayload := copyMap(goodWillOptions)
	emptyPayload[WillTopic] = ""

	badRetained := copyMap(goodWillOptions)
	badRetained[WillRetained] = "junk"

	badQos := copyMap(goodWillOptions)
	badQos[WillQos] = "junk"

	tests := []struct {
		Name        string
		Params      map[string]string
		ExpectError bool
	}{
		{"Happy Path all options", goodWillOptions, false},
		{"Happy Path no options", emptyWill, false},
		{"Happy Path will disabled", emptyWill, false},
		{"Bad WillEnabled", badEnabled, true},
		{"Bad WillRetained", badRetained, true},
		{"Bad WillQos", badQos, true},
		{"Missing Topic", missingTopic, true},
		{"Empty Topic", emptyTopic, true},
		{"Missing Payload", missingPayload, true},
		{"Empty Payload", emptyPayload, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			test.Params[BrokerAddress] = "mqtt://broker:8883"
			test.Params[Topic] = "topic"
			test.Params[SecretName] = "my-secret"
			test.Params[ClientID] = "clientid"
			test.Params[AuthMode] = "none"

			trx := configurable.MQTTExport(test.Params)

			if test.ExpectError {
				assert.Nil(t, trx)
				return
			}

			assert.NotNil(t, trx)
		})
	}
}

func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string)
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

func TestAddTags(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		Name      string
		ParamName string
		TagsSpec  string
		ExpectNil bool
	}{
		{"Good - non-empty list", Tags, "GatewayId:HoustonStore000123,Latitude:29.630771,Longitude:-95.377603", false},
		{"Good - empty list", Tags, "", false},
		{"Bad - No : separator", Tags, "GatewayId HoustonStore000123, Latitude:29.630771,Longitude:-95.377603", true},
		{"Bad - Missing value", Tags, "GatewayId:,Latitude:29.630771,Longitude:-95.377603", true},
		{"Bad - Missing key", Tags, "GatewayId:HoustonStore000123,:29.630771,Longitude:-95.377603", true},
		{"Bad - Missing key & value", Tags, ":,:,:", true},
		{"Bad - No Tags parameter", "NotTags", ":,:,:", true},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			params := make(map[string]string)
			params[testCase.ParamName] = testCase.TagsSpec

			transform := configurable.AddTags(params)
			assert.Equal(t, testCase.ExpectNil, transform == nil)
		})
	}
}

func TestEncrypt(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		Name           string
		Algorithm      string
		SecretName     string
		SecretValueKey string
		ExpectNil      bool
	}{
		{"AES256 - Bad - No secrets ", EncryptAES256, "", "", true},
		{"AES256 - good - secrets", EncryptAES256, uuid.NewString(), uuid.NewString(), false},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			params := make(map[string]string)
			if len(testCase.Algorithm) > 0 {
				params[Algorithm] = testCase.Algorithm
			}
			if len(testCase.SecretName) > 0 {
				params[SecretName] = testCase.SecretName
			}
			if len(testCase.SecretValueKey) > 0 {
				params[SecretValueKey] = testCase.SecretValueKey
			}

			transform := configurable.Encrypt(params)
			assert.Equal(t, testCase.ExpectNil, transform == nil)
		})
	}
}

func TestConfigurable_WrapIntoEvent(t *testing.T) {
	configurable := Configurable{lc: lc}

	profileName := "MyProfile"
	deviceName := "MyDevice"
	resourceName := "MyResource"
	simpleValueType := "int64"
	binaryValueType := "binary"
	objectValueType := "object"
	badValueType := "bogus"
	mediaType := "application/mxl"
	emptyMediaType := ""

	tests := []struct {
		Name         string
		ProfileName  *string
		DeviceName   *string
		ResourceName *string
		ValueType    *string
		MediaType    *string
		ExpectNil    bool
	}{
		{"Valid simple", &profileName, &deviceName, &resourceName, &simpleValueType, nil, false},
		{"Invalid simple - missing profile", nil, &deviceName, &resourceName, &simpleValueType, nil, true},
		{"Invalid simple - missing device", &profileName, nil, &resourceName, &simpleValueType, nil, true},
		{"Invalid simple - missing resource", &profileName, &deviceName, nil, &simpleValueType, nil, true},
		{"Invalid simple - missing value type", &profileName, &deviceName, &resourceName, nil, nil, true},
		{"Invalid - bad value type", &profileName, &deviceName, &resourceName, &badValueType, nil, true},
		{"Valid binary", &profileName, &deviceName, &resourceName, &binaryValueType, &mediaType, false},
		{"Invalid binary - empty MediaType", &profileName, &deviceName, &resourceName, &binaryValueType, &emptyMediaType, true},
		{"Invalid binary - missing MediaType", &profileName, &deviceName, &resourceName, &binaryValueType, nil, true},
		{"Valid object", &profileName, &deviceName, &resourceName, &objectValueType, nil, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			params := make(map[string]string)
			if testCase.ProfileName != nil {
				params[ProfileName] = *testCase.ProfileName
			}
			if testCase.DeviceName != nil {
				params[DeviceName] = *testCase.DeviceName
			}
			if testCase.ResourceName != nil {
				params[ResourceName] = *testCase.ResourceName
			}
			if testCase.ValueType != nil {
				params[ValueType] = *testCase.ValueType
			}
			if testCase.MediaType != nil {
				params[MediaType] = *testCase.MediaType
			}

			transform := configurable.WrapIntoEvent(params)
			assert.Equal(t, testCase.ExpectNil, transform == nil)
		})
	}
}

func TestConfigurable_ToLineProtocol(t *testing.T) {
	configurable := Configurable{lc: lc}

	tests := []struct {
		Name      string
		Params    map[string]string
		ExpectNil bool
	}{
		{"Valid, empty tags parameter", map[string]string{Tags: ""}, false},
		{"Valid, some tags", map[string]string{Tags: "tag1:value1, tag2:value2"}, false},
		{"Invalid, no tags parameter", nil, true},
		{"Invalid, bad tags", map[string]string{Tags: "tag1 = value1, tag2 =value2"}, true},
	}

	for _, test := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			actual := configurable.ToLineProtocol(test.Params)
			assert.Equal(t, test.ExpectNil, actual == nil)
		})
	}
}
