// TODO: Change Copyright to your company if open sourcing or remove header
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

package config

// This file contains example of custom configuration that can be loaded from the service's configuration.toml
// and/or the Configuration Provider, aka Consul (if enabled).
// For more details see https://docs.edgexfoundry.org/2.0/microservices/application/GeneralAppServiceConfig/#custom-configuration
// TODO: Update this configuration as needed for you service's needs and remove this comment
//       or remove this file if not using custom configuration.

import (
	"errors"
	"reflect"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// TODO: Define your structured custom configuration types. Must be wrapped with an outer struct with
//       single element that matches the top level custom configuration element in your configuration.toml file,
//       'AppCustom' in this example. Replace this example with your configuration structure or
//       remove this file if not using structured custom configuration.
type ServiceConfig struct {
	AppCustom AppCustomConfig
}

// AppCustomConfig is example of service's custom structured configuration that is specified in the service's
// configuration.toml file and Configuration Provider (aka Consul), if enabled.
type AppCustomConfig struct {
	ResourceNames string
	SomeValue     int
	SomeService   HostInfo
}

// HostInfo is example struct for defining connection information for external service
type HostInfo struct {
	Host     string
	Port     int
	Protocol string
}

// TODO: Update using your Custom configuration type.
// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false //errors.New("unable to cast raw config to type 'ServiceConfig'")
	}

	*c = *configuration

	return true
}

// TODO: Update using your Custom configuration 'writeable' type or remove if not using ListenForCustomConfigChanges
// UpdateWritableFromRaw updates the service's writable configuration from raw data received from
// the Service Provider. Must implement if using ListenForCustomConfigChanges, otherwise this can be removed.
func (ac *AppCustomConfig) UpdateWritableFromRaw(rawWritableConfig interface{}) bool {
	appCustom, ok := rawWritableConfig.(*AppCustomConfig)
	if !ok {
		return false //errors.New("unable to cast raw writeable config to type 'AppCustomConfig'")
	}

	*ac = *appCustom

	return true
}

// WaitForCustomConfigChanges waits for indication that the custom configuration section has been updated and then process
// the changes as needed
// TODO: Update to use your custom configuration section that you want to be writable (i.e. runtime changes from Consul)
//       or remove if not using custom configuration section or writable custom configuration.
func (ac *AppCustomConfig) WaitForCustomConfigChanges(configChanged chan bool, lc logger.LoggingClient) {
	previous := *ac // Copy for change detection

	go func() {
		for {
			select {
			case <-configChanged:
				// TODO: Process the changed configuration.
				//       Must keep a previous copy of the configuration to determine what has changed.
				//		 Replace the examples below with your appropriate processing logic.
				switch {
				case previous.SomeValue != ac.SomeValue:
					lc.Infof("AppCustom.SomeValue changed to: %d", ac.SomeValue)
				case previous.ResourceNames != ac.ResourceNames:
					lc.Infof("AppCustom.ResourceNames changed to: %s", ac.ResourceNames)
				case !reflect.DeepEqual(previous.SomeService, ac.SomeService):
					lc.Infof("AppCustom.SomeService changed to: %v", ac.SomeService)
				default:
					lc.Info("No changes detected")
				}

				previous = *ac
			}
		}
	}()
}

// Validate ensures your custom configuration has proper values.
// TODO: Update to properly validate your custom configuration
func (ac *AppCustomConfig) Validate() error {
	if ac.SomeValue <= 0 {
		return errors.New("SomeValue must be greater than zero")
	}

	if reflect.DeepEqual(ac.SomeService, HostInfo{}) {
		return errors.New("SomeService is not set")
	}

	return nil
}
