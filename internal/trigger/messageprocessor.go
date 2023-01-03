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

package trigger

import (
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
)

type MessageProcessor interface {
	// MessageReceived provides runtime orchestration to pass the envelope / context to configured pipeline(s)
	MessageReceived(ctx interfaces.AppFunctionContext, envelope types.MessageEnvelope, outputHandler interfaces.PipelineResponseHandler) error
	// ReceivedInvalidMessage is called when an invalid message is received so the metrics counter can be incremented.
	ReceivedInvalidMessage()
}
