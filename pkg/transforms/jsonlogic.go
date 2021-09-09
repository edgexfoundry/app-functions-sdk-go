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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/diegoholiveira/jsonlogic"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
)

// JSONLogic ...
type JSONLogic struct {
	Rule string
}

// NewJSONLogic creates, initializes and returns a new instance of HTTPSender
func NewJSONLogic(rule string) JSONLogic {
	return JSONLogic{
		Rule: rule,
	}
}

// Evaluate ...
func (logic JSONLogic) Evaluate(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		// We didn't receive a result
		return false, fmt.Errorf("function Evaluate in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	coercedData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}

	reader := strings.NewReader(string(coercedData))
	rule := strings.NewReader(logic.Rule)

	var logicResult bytes.Buffer
	ctx.LoggingClient().Debugf("Applying JSONLogic Rule in pipeline '%s'", ctx.PipelineId())
	err = jsonlogic.Apply(rule, reader, &logicResult)
	if err != nil {
		return false, fmt.Errorf("unable to apply JSONLogic rule in pipeline '%s': %s", ctx.PipelineId(), err.Error())
	}

	var result bool
	decoder := json.NewDecoder(&logicResult)
	err = decoder.Decode(&result)
	if err != nil {
		return false, fmt.Errorf("unable to decode JSONLogic result in pipeline '%s': %s", ctx.PipelineId(), err.Error())
	}

	ctx.LoggingClient().Debugf("Condition met in pipeline '%s': %s", ctx.PipelineId(), strconv.FormatBool(result))

	return result, data
}
