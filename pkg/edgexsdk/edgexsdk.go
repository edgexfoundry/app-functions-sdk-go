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
	"flag"
	"fmt"

	"os"
	"os/signal"
	"syscall"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/common"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/configuration"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger/messagebus"
	registry "github.com/edgexfoundry/go-mod-registry"
	"github.com/edgexfoundry/go-mod-registry/pkg/factory"
)

// AppFunctionsSDK ...
type AppFunctionsSDK struct {
	transforms    []func(edgexcontext excontext.Context, params ...interface{}) (bool, interface{})
	ServiceKey    string
	configProfile string
	configDir     string
	useRegistry   bool
	config        common.ConfigurationStruct
}

// SetPipeline defines the order in which each function will be called as each event comes in.
func (sdk *AppFunctionsSDK) SetPipeline(transforms ...func(edgexcontext excontext.Context, params ...interface{}) (bool, interface{})) {
	sdk.transforms = transforms
}

// FilterByDeviceID ...
func (sdk *AppFunctionsSDK) FilterByDeviceID(deviceIDs []string) func(excontext.Context, ...interface{}) (bool, interface{}) {
	transforms := transforms.Filter{
		FilterValues: deviceIDs,
	}
	return transforms.FilterByDeviceID
}

// FilterByValueDescriptor ...
func (sdk *AppFunctionsSDK) FilterByValueDescriptor(valueIDs []string) func(excontext.Context, ...interface{}) (bool, interface{}) {
	transforms := transforms.Filter{
		FilterValues: valueIDs,
	}
	return transforms.FilterByValueDescriptor
}

// TransformToXML ...
func (sdk *AppFunctionsSDK) TransformToXML() func(excontext.Context, ...interface{}) (bool, interface{}) {
	transforms := transforms.Conversion{}
	return transforms.TransformToXML
}

// TransformToJSON ...
func (sdk *AppFunctionsSDK) TransformToJSON() func(excontext.Context, ...interface{}) (bool, interface{}) {
	transforms := transforms.Conversion{}
	return transforms.TransformToJSON
}

// HTTPPost ...
func (sdk *AppFunctionsSDK) HTTPPost(url string) func(excontext.Context, ...interface{}) (bool, interface{}) {
	transforms := transforms.HTTPSender{
		URL: url,
	}
	return transforms.HTTPPost
}

//MakeItRun the SDK
func (sdk *AppFunctionsSDK) MakeItRun() {
	// TODO: reconcile this with the new configuration.toml/Registry
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
	runtime := runtime.GolangRuntime{Transforms: sdk.transforms}

	// determine input type and create trigger for it
	trigger := sdk.setupTrigger(configuration, runtime)

	// Initialize the trigger (i.e. start a web server, or connect to message bus)
	trigger.Initialize()

}

func (sdk *AppFunctionsSDK) setupTrigger(configuration configuration.Configuration, runtime runtime.GolangRuntime) trigger.ITrigger {
	var trigger trigger.ITrigger
	// Need to make dynamic, search for the binding that is input
	switch configuration.Bindings[0].Type {
	case "http":
		println("Loading Http Trigger")
		trigger = &http.HTTPTrigger{Configuration: configuration, Runtime: runtime}
	case "messageBus":
		trigger = &messagebus.MessageBusTrigger{Configuration: configuration, Runtime: runtime}
	}
	return trigger
}

// Initialize the SDK
func (sdk *AppFunctionsSDK) Initialize() error {
	// Handles SIGINT/SIGTERM and exits gracefully
	listenForInterrupts()

	flag.BoolVar(&sdk.useRegistry, "registry", false, "Indicates the service should use the registry.")
	flag.BoolVar(&sdk.useRegistry, "r", false, "Indicates the service should use registry.")

	flag.StringVar(&sdk.configProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&sdk.configProfile, "p", "", "Specify a profile other than default.")

	flag.StringVar(&sdk.configDir, "config", "", "Specify an alternate configuration directory.")
	flag.StringVar(&sdk.configDir, "c", "", "Specify an alternate configuration directory.")

	flag.Parse()

	err := sdk.initializeRegistry()
	if err != nil {
		return fmt.Errorf("failed to initialize Registry: %v", err)
	}

	// TODO: Add additional initialization like logging, message bus, etc.
	return nil
}

func (sdk *AppFunctionsSDK) initializeRegistry() error {

	// Currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	configuration := &common.ConfigurationStruct{}
	err := common.LoadFromFile(sdk.configProfile, sdk.configDir, configuration)
	if err != nil {
		return err
	}

	if sdk.useRegistry {
		registryConfig := registry.Config{
			Host:          configuration.Registry.Host,
			Port:          configuration.Registry.Port,
			Type:          configuration.Registry.Type,
			Stem:          internal.ConfigRegistryStem,
			CheckInterval: "1s",
			CheckRoute:    internal.ApiPingRoute,
			ServiceHost:   configuration.Service.Host,
			ServicePort:   configuration.Service.Port,
		}

		registryClient, err := factory.NewRegistryClient(registryConfig, sdk.ServiceKey)
		if err != nil {
			return fmt.Errorf("connection to Registry could not be made: %v", err)
		}

		if !registryClient.IsRegistryRunning() {
			return fmt.Errorf("registry (%s) is not running", registryConfig.Type)
		}

		// Register the service with Registry
		err = registryClient.Register()
		if err != nil {
			return fmt.Errorf("could not register service with Registry: %v", err)
		}

		rawConfig, err := registryClient.GetConfiguration(configuration)
		if err != nil {
			return fmt.Errorf("could not get configuration from Registry: %v", err)
		}

		actual, ok := rawConfig.(*common.ConfigurationStruct)
		if !ok {
			return fmt.Errorf("configuration from Registry failed type check")
		}

		sdk.config = *actual

		go sdk.listenForConfigChanges(registryClient)
	}

	return nil
}

func (sdk *AppFunctionsSDK) listenForConfigChanges(registryClient registry.Client) {
	var errChannel chan error          //A channel for "config wait error" sourced from Registry
	var updateChannel chan interface{} //A channel for "config updates" sourced from Registry

	registryClient.WatchForChanges(updateChannel, errChannel, &common.WritableInfo{}, internal.WritableKey)

	for {
		select {
		case err := <-errChannel:
			// TODO: remove println and uncomment Logging once logging package is available
			//LoggingClient.Error(err.Error())
			fmt.Println(err.Error())

		case raw, ok := <-updateChannel:
			if !ok {
				return
			}

			actual, ok := raw.(*common.WritableInfo)
			if !ok {
				// TODO: remove println and uncomment Logging once logging package is available
				//LoggingClient.Error("listenForConfigChanges() type check failed")
				fmt.Println("listenForConfigChanges() type check failed")
				return
			}

			sdk.config.Writable = *actual

			// TODO: remove println and uncomment Logging once logging package is available
			//LoggingClient.Info("Writeable configuration has been updated. Setting log level to " + sdk.config.Writable.LogLevel)
			fmt.Println("Writeable configuration has been updated. Setting log level to " + sdk.config.Writable.LogLevel)
			//LoggingClient.SetLogLevel(sdk.config.Writable.LogLevel)

			// TODO: Deal with pub/sub topics may have changed. Save copy of writeable so that we can determine what if anything changed?
		}
	}
}

func listenForInterrupts() {
	go func() {
		signalChan := make(chan os.Signal)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		signalReceived := <-signalChan
		fmt.Printf("Terminating: %s\n", signalReceived)
		os.Exit(0)
	}()
}
