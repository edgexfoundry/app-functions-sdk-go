package messagebus

import (
    "context"
    "fmt"
    "github.com/edgexfoundry/go-mod-messaging/messaging"
    "github.com/edgexfoundry/go-mod-messaging/pkg/types"
    "sync"
)

type ZeromqInitBehavior struct{}

func (z *ZeromqInitBehavior) Initialize(trigger *Trigger, appWg *sync.WaitGroup, appCtx context.Context) error {
    var err error
    logger := trigger.EdgeXClients.LoggingClient

    logger.Info(fmt.Sprintf("Initializing Message Bus Trigger. Subscribing to topic: %s on port %d , Publish Topic: %s on port %d", trigger.Configuration.Binding.SubscribeTopic, trigger.Configuration.MessageBus.SubscribeHost.Port, trigger.Configuration.Binding.PublishTopic, trigger.Configuration.MessageBus.PublishHost.Port))

    trigger.client, err = messaging.NewMessageClient(trigger.Configuration.MessageBus)
    if err != nil {
        return err
    }
    trigger.topics = []types.TopicChannel{{Topic: trigger.Configuration.Binding.SubscribeTopic, Messages: make(chan types.MessageEnvelope)}}
    messageErrors := make(chan error)

    err = trigger.client.Connect()
    if err != nil {
        return err
    }

    trigger.client.Subscribe(trigger.topics, messageErrors)
    receiveMessage := true

    appWg.Add(1)

    go func() {
        defer appWg.Done()

        for receiveMessage {
            select {
            case <-appCtx.Done():
                return

            case msgErr := <-messageErrors:
                logger.Error(fmt.Sprintf("Failed to receive ZMQ Message, %v", msgErr))

            case msgs := <-trigger.topics[0].Messages:
                HandleMessage(trigger, msgs, logger)
            }
        }
    }()

    return nil
}
