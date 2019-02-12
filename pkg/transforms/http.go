package transforms

import (
	"bytes"
	"net/http"
)

const mimeTypeJSON = "application/json"

// HTTPSender ...
type HTTPSender struct {
	URL    string
	Method string
}

// HTTPPost ...
func HTTPPost(url string, data []byte) bool {

	response, err := http.Post(url, mimeTypeJSON, bytes.NewReader(data))
	if err != nil {
		//LoggingClient.Error(err.Error())
		return false
	}
	defer response.Body.Close()
	// LoggingClient.Info(fmt.Sprintf("Response: %s", response.Status))

	// LoggingClient.Info(fmt.Sprintf("Sent data: %X", data))
	return true
}
