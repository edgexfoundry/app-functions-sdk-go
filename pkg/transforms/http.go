package transforms

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
)

const mimeTypeJSON = "application/json"

// HTTPSender ...
type HTTPSender struct {
	URL    string
	Method string
}

// HTTPPost ...
func (sender HTTPSender) HTTPPost(edgexcontext excontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	if result, ok := params[0].(string); ok {
		response, err := http.Post(sender.URL, mimeTypeJSON, bytes.NewReader(([]byte)(result)))
		if err != nil {
			//LoggingClient.Error(err.Error())
			return false, err
		}
		defer response.Body.Close()

	}
	// LoggingClient.Info(fmt.Sprintf("Response: %s", response.Status))

	// LoggingClient.Info(fmt.Sprintf("Sent data: %X", data))
	return false, errors.New("Unexpected type received")
}
