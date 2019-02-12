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

package main

import (
	"flag"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/edgexsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
)

const (
	serviceKey = "sampleFilterXml"
)

func main() {
	var configProfile string
	var useRegistry bool
	// need to move to boostrap outside of main and into SDK portion
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use the registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry.")
	flag.StringVar(&configProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&configProfile, "p", "", "Specify a profile other than default.")

	flag.Parse()
	// 1) First thing to do is to create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &edgexsdk.AppFunctionsSDK{}
	edgexSdk.Initialize(useRegistry, configProfile, serviceKey)

	// 2) Since our FilterByDeviceID Function requires the list of DeviceID's we would
	// like to search for, we'll go ahead and define that now.
	deviceIDs := []string{"GS1-AC-Drive01"}
	// 3) This is our pipeline configuration, the collection of functions to
	// execute every time an event is triggered.
	edgexSdk.SetPipeline(
		edgexSdk.FilterByDeviceID(deviceIDs),
		edgexSdk.TransformToJSON(),
		printXMLToConsole,
	)

	// 4) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	edgexSdk.MakeItRun()
}

func printXMLToConsole(edgexcontext excontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	println(params[0].(string))
	edgexcontext.Complete(params[0].(string))
	return false, nil
}
