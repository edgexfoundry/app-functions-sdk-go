//
// Copyright (c) 2020 Technotects
// Copyright (c) 2022-2023 Intel Corporation
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

package app

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/internal/trigger/mqtt"
	"github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces"
)

const (
	// Valid types of App Service triggers
	TriggerTypeMessageBus = "EDGEX-MESSAGEBUS"
	TriggerTypeMQTT       = "EXTERNAL-MQTT"
	TriggerTypeHTTP       = "HTTP"
)

func (svc *Service) setupTrigger(configuration *common.ConfigurationStruct) interfaces.Trigger {
	var t interfaces.Trigger

	serviceBinding := NewTriggerServiceBinding(svc)
	messageProcessor := NewTriggerMessageProcessor(serviceBinding, svc.MetricsManager())

	switch triggerType := strings.ToUpper(configuration.Trigger.Type); triggerType {
	case TriggerTypeHTTP:
		svc.LoggingClient().Info("HTTP trigger selected")
		t = http.NewTrigger(serviceBinding, messageProcessor, svc.webserver)

	case TriggerTypeMessageBus:
		svc.LoggingClient().Info("EdgeX MessageBus trigger selected")
		t = messagebus.NewTrigger(serviceBinding, messageProcessor, svc.dic)

	case TriggerTypeMQTT:
		svc.LoggingClient().Info("External MQTT trigger selected")
		t = mqtt.NewTrigger(serviceBinding, messageProcessor)

	default:
		if factory, found := svc.customTriggerFactories[triggerType]; found {
			var err error
			t, err = factory(svc)
			if err != nil {
				svc.LoggingClient().Errorf("failed to initialize custom trigger [%s]: %s", triggerType, err.Error())
				return nil
			}
		} else if len(configuration.Trigger.Type) == 0 {
			svc.LoggingClient().Error("Trigger type not found, missing common config? Use -cp or -cc flags for common config")
		} else {
			svc.LoggingClient().Errorf("Invalid Trigger type of '%s' specified", configuration.Trigger.Type)
		}
	}

	return t
}

// RegisterCustomTriggerFactory allows users to register builders for custom trigger types
func (svc *Service) RegisterCustomTriggerFactory(name string,
	factory func(interfaces.TriggerConfig) (interfaces.Trigger, error)) error {
	nu := strings.ToUpper(name)

	if nu == TriggerTypeMessageBus ||
		nu == TriggerTypeHTTP ||
		nu == TriggerTypeMQTT {
		return fmt.Errorf("cannot register custom trigger for builtin type (%s)", name)
	}

	if svc.customTriggerFactories == nil {
		svc.customTriggerFactories = make(map[string]func(sdk *Service) (interfaces.Trigger, error), 1)
	}

	svc.customTriggerFactories[nu] = func(sdk *Service) (interfaces.Trigger, error) {
		serviceBinding := NewTriggerServiceBinding(sdk)
		messageProcessor := NewTriggerMessageProcessor(serviceBinding, svc.MetricsManager())

		cfg := interfaces.TriggerConfig{
			Logger:          sdk.lc,
			ContextBuilder:  serviceBinding.BuildContext,
			MessageReceived: messageProcessor.MessageReceived,
			ConfigLoader:    serviceBinding.LoadCustomConfig,
		}

		return factory(cfg)
	}

	return nil
}
