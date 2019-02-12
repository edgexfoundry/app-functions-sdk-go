package transforms

import (
	"bytes"
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

	response, err := http.Post(sender.URL, mimeTypeJSON, bytes.NewReader(params[0].([]byte)))
	if err != nil {
		//LoggingClient.Error(err.Error())
		return false, err
	}
	defer response.Body.Close()
	// LoggingClient.Info(fmt.Sprintf("Response: %s", response.Status))

	// LoggingClient.Info(fmt.Sprintf("Sent data: %X", data))
	return false, nil
}
