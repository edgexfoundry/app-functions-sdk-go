package messagebus

import (
    "fmt"
    "github.com/edgexfoundry/app-functions-sdk-go/internal/common"
    "github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
)

func NewMessageBusTrigger(configuration common.ConfigurationStruct, runtime *runtime.GolangRuntime, edgexClient common.EdgeXClients) *Trigger {
    switch configuration.MessageBus.Type {
    case "mqtt":
        return &Trigger{
            Configuration:      configuration,
            InitializeBehavior: &MqttInitBehavior{},
            Runtime:            runtime,
            EdgeXClients:       edgexClient,
        }
    case "zero":
        return &Trigger{
            Configuration:      configuration,
            InitializeBehavior: &ZeromqInitBehavior{},
            Runtime:            runtime,
            EdgeXClients:       edgexClient,
        }
    default:
        errMsg := fmt.Sprintf("%s messagebus type is not supported", configuration.MessageBus.Type)
        edgexClient.LoggingClient.Error(errMsg)
        panic(errMsg)
    }
}
