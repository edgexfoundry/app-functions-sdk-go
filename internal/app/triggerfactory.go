//
// Copyright (c) 2020 Technotects
// Copyright (c) 2021 Intel Corporation
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

	gometrics "github.com/rcrowley/go-metrics"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mqtt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

const (
	// Valid types of App Service triggers
	TriggerTypeMessageBus = "EDGEX-MESSAGEBUS"
	TriggerTypeMQTT       = "EXTERNAL-MQTT"
	TriggerTypeHTTP       = "HTTP"
)

func (svc *Service) setupTrigger(configuration *common.ConfigurationStruct) interfaces.Trigger {
	var t interfaces.Trigger

	bnd := NewTriggerServiceBinding(svc)
	mp := &triggerMessageProcessor{
		bnd:              bnd,
		messagesReceived: gometrics.NewCounter()}

	if err := svc.MetricsManager().Register(internal.MessagesReceivedName, mp.messagesReceived, nil); err != nil {
		svc.lc.Warnf("%s metric failed to register and will not be reported: %s", internal.MessagesReceivedName, err.Error())
	} else {
		svc.lc.Infof("%s metric has been registered and will be reported", internal.MessagesReceivedName)
	}

	switch triggerType := strings.ToUpper(configuration.Trigger.Type); triggerType {
	case TriggerTypeHTTP:
		svc.LoggingClient().Info("HTTP trigger selected")
		t = http.NewTrigger(bnd, mp, svc.webserver)

	case TriggerTypeMessageBus:
		svc.LoggingClient().Info("EdgeX MessageBus trigger selected")
		t = messagebus.NewTrigger(bnd, mp, svc.dic)

	case TriggerTypeMQTT:
		svc.LoggingClient().Info("External MQTT trigger selected")
		t = mqtt.NewTrigger(bnd, mp)

	default:
		if factory, found := svc.customTriggerFactories[triggerType]; found {
			var err error
			t, err = factory(svc)
			if err != nil {
				svc.LoggingClient().Errorf("failed to initialize custom trigger [%s]: %s", triggerType, err.Error())
				return nil
			}
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
		binding := NewTriggerServiceBinding(sdk)
		messageProcessor := &triggerMessageProcessor{
			bnd:              binding,
			messagesReceived: gometrics.NewCounter()}

		if err := svc.MetricsManager().Register(internal.MessagesReceivedName, messageProcessor.messagesReceived, nil); err != nil {
			svc.lc.Warnf("%s metric failed to register and will not be reported: %s", internal.MessagesReceivedName, err.Error())
		} else {
			svc.lc.Infof("%s metric has been registered and will be reported (if enabled)", internal.MessagesReceivedName)
		}

		cfg := interfaces.TriggerConfig{
			Logger:           sdk.lc,
			ContextBuilder:   binding.BuildContext,
			MessageProcessor: messageProcessor.Process,
			MessageReceived:  messageProcessor.MessageReceived,
			ConfigLoader:     binding.LoadCustomConfig,
		}

		return factory(cfg)
	}

	return nil
}
