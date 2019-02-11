//
// Copyright (c) 2019 Intel Corporation
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

package edgexsdk

import (
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/common"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/configuration"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger"
	httptrigger "github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger/http"
	messagebustrigger "github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger/messagebus"
	"github.com/edgexfoundry/go-mod-registry"
)

// AppFunctionsSDK ...
type AppFunctionsSDK struct {
	transforms []func(params ...interface{}) (bool, interface{})
	config common.ConfigurationStruct
}

// SetPipeline defines the order in which each function will be called as each event comes in.
func (afsdk *AppFunctionsSDK) SetPipeline(transforms ...func(params ...interface{}) (bool, interface{})) {
	afsdk.transforms = transforms
}

// FilterByDeviceID ...
func (afsdk *AppFunctionsSDK) FilterByDeviceID(deviceIDs []string) func(...interface{}) (bool, interface{}) {
	transforms := transforms.Filter{
		FilterValues: deviceIDs,
	}
	return transforms.FilterByDeviceID
}

// FilterByValueDescriptor ...
func (afsdk *AppFunctionsSDK) FilterByValueDescriptor(valueIDs []string) func(...interface{}) (bool, interface{}) {
	transforms := transforms.Filter{
		FilterValues: valueIDs,
	}
	return transforms.FilterByValueDescriptor
}

// TransformToXML ...
func (afsdk *AppFunctionsSDK) TransformToXML() func(...interface{}) (bool, interface{}) {
	transforms := transforms.Conversion{}
	return transforms.TransformToXML
}

//MakeItRun the SDK
func (afsdk *AppFunctionsSDK) MakeItRun() {

	// load the configuration
	configuration := configuration.Configuration{
		Bindings: []configuration.Binding{
			configuration.Binding{
				Type: "http",
			},
		},
	} //configuration.LoadConfiguration()

	// a little telemetry where?

	//determine which runtime to load
	runtime := runtime.GolangRuntime{Transforms: afsdk.transforms}

	// determine input type and create trigger for it
	trigger := afsdk.setupTrigger(configuration, runtime)

	// Initialize the trigger (i.e. start a web server, or connect to message bus)
	trigger.Initialize()

}

func (afsdk *AppFunctionsSDK) setupTrigger(configuration configuration.Configuration, runtime runtime.GolangRuntime) trigger.ITrigger {
	var trigger trigger.ITrigger
	// Need to make dynamic, search for the binding that is input
	switch configuration.Bindings[0].Type {
	case "http":
		println("Loading Http Trigger")
		trigger = &httptrigger.HTTPTrigger{Configuration: configuration, Runtime: runtime}
	case "messageBus":
		trigger = &messagebustrigger.MessageBusTrigger{Configuration: configuration, Runtime: runtime}
	}
	return trigger
}
func (afsdk *AppFunctionsSDK) Initialize(useRegistry bool, useProfile string, serviceKey string ) error {
	//We currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	config := &common.ConfigurationStruct{}
	err := common.LoadFromFile(useProfile, config)
	if err != nil {
		return err
	}

	if useRegistry {
		registryConfig := registry.Config{
			Host:          config.Registry.Host,
			Port:          config.Registry.Port,
			Type:          "consul",
			Stem:          "edgex/core/1.0/",
			CheckInterval: "1s",
			CheckRoute:    "/api/v1/ping",
			ServiceHost: 	config.Service.Host,
			ServicePort: 	config.Service.Port,
		}

		registryClient, err := registry.NewRegistryClient(registryConfig, serviceKey)
		if err != nil {
			return fmt.Errorf("connection to Registry could not be made: %v", err)
		}

		// Register the service with Registry
		err = registryClient.Register()
		if err != nil {
			return fmt.Errorf("could not register service with Registry: %v", err)
		}

		rawConfig, err := registry.Client.GetConfiguration(config)
		if err != nil {
			return fmt.Errorf("could not get configuration from Registry: %v", err)
		}

		actual, ok := rawConfig.(*common.ConfigurationStruct)
		if !ok {
			return fmt.Errorf("configuration from Registry failed type check")
		}

		afsdk.config = *actual

		go func() {
			var errChannel chan error //A channel for "config wait error" sourced from Registry
			var updateChannel chan interface{} //A channel for "config updates" sourced from Registry


			registry.Client.WatchForChanges(updateChannel, errChannel, &common.WritableInfo{}, internal.WritableKey)

			for {
				select {
				case ex := <-errChannel:
					LoggingClient.Error(ex.Error())

				case raw, ok := <-updateChannel:
					if !ok {
						return
					}

					actual, ok := raw.(*common.WritableInfo)
					if !ok {
						LoggingClient.Error("listenForConfigChanges() type check failed")
						return
					}

					afsdk.config.Writable = *actual

					LoggingClient.Info("Writeable configuration has been updated. Setting log level to " + afsdk.config.Writable.LogLevel)
					LoggingClient.SetLogLevel(afsdk.config.Writable.LogLevel)
				}
			}
		}()

	}

	return nil
}
