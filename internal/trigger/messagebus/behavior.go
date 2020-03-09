package messagebus

import (
    "context"
    "fmt"
    "github.com/edgexfoundry/app-functions-sdk-go/appcontext"
    "github.com/edgexfoundry/go-mod-core-contracts/clients"
    "github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
    "github.com/edgexfoundry/go-mod-messaging/pkg/types"
    "sync"
)

type InitializeBehavior interface {
    Initialize(trigger *Trigger, appWg *sync.WaitGroup, appCtx context.Context) error
}

func HandleMessage(trigger *Trigger, msgs types.MessageEnvelope, logger logger.LoggingClient) {
    go func() {
        logger.Trace("Received message from bus", "topic", trigger.Configuration.Binding.SubscribeTopic, clients.CorrelationHeader, msgs.CorrelationID)

        edgexContext := &appcontext.Context{
            CorrelationID:         msgs.CorrelationID,
            Configuration:         trigger.Configuration,
            LoggingClient:         trigger.EdgeXClients.LoggingClient,
            EventClient:           trigger.EdgeXClients.EventClient,
            ValueDescriptorClient: trigger.EdgeXClients.ValueDescriptorClient,
            CommandClient:         trigger.EdgeXClients.CommandClient,
            NotificationsClient:   trigger.EdgeXClients.NotificationsClient,
        }

        messageError := trigger.Runtime.ProcessMessage(edgexContext, msgs)
        if messageError != nil {
            // ProcessMessage logs the error, so no need to log it here.
            return
        }

        if edgexContext.OutputData != nil {
            outputEnvelope := types.MessageEnvelope{
                CorrelationID: edgexContext.CorrelationID,
                Payload:       edgexContext.OutputData,
                ContentType:   clients.ContentTypeJSON,
            }
            err := trigger.client.Publish(outputEnvelope, trigger.Configuration.Binding.PublishTopic)
            if err != nil {
                logger.Error(fmt.Sprintf("Failed to publish Message to bus, %v", err))
            }

            logger.Trace("Published message to bus", "topic", trigger.Configuration.Binding.PublishTopic, clients.CorrelationHeader, msgs.CorrelationID)
        }
    }()
}
