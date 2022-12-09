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

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/internal/common"
)

// ConfigurationName contains the name of data's common.ConfigurationStruct implementation in the DIC.
var ConfigurationName = di.TypeInstanceToName(common.ConfigurationStruct{})

// ConfigurationFrom helper function queries the DIC and returns service's common.ConfigurationStruct implementation.
func ConfigurationFrom(get di.Get) *common.ConfigurationStruct {
	return get(ConfigurationName).(*common.ConfigurationStruct)
}
