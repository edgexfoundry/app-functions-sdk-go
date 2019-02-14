package edgexsdk

import (
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
)

func TestSetPipelineNoTransforms(t *testing.T) {
	sdk := AppFunctionsSDK{}
	err := sdk.SetPipeline()
	if err == nil {
		t.Fatal("Should return error")
	}
	if err.Error() != "No transforms provided to pipeline" {
		t.Fatal("Incorrect error message received")
	}
}
func TestSetPipelineNoTransformsNil(t *testing.T) {
	sdk := AppFunctionsSDK{}
	transform1 := func(edgexcontext excontext.Context, params ...interface{}) (bool, interface{}) {
		return false, nil
	}
	err := sdk.SetPipeline(transform1)
	if err != nil {
		t.Fatal("Error should be nil")
	}
	if len(sdk.transforms) != 1 {
		t.Fatal("sdk.Transforms should have 1 transform")
	}

}

func TestFilterByDeviceID(t *testing.T) {
	sdk := AppFunctionsSDK{}
	deviceIDs := []string{"GS1-AC-Drive01"}

	trx := sdk.FilterByDeviceID(deviceIDs)
	if trx == nil {
		t.Fatal("return result from FilterByDeviceID should not be nil")
	}
}
