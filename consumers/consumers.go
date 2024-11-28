package consumers

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"quiz/configs"
)

var groupId = uuid.New().String() // each instance has a unique group ID

func Consume(topic string, handler sarama.ConsumerGroupHandler) {
	config := sarama.NewConfig()
	config.Version = sarama.V1_1_0_0
	config.Consumer.Return.Errors = true

	group, err := sarama.NewConsumerGroup(configs.KafkaBrokerAddress, groupId, config)
	if err != nil {
		log.Fatalln(err)
	}

	// Track errors
	go func() {
		for err := range group.Errors() {
			fmt.Println("ERROR", err)
		}
	}()

	// Iterate over consumer sessions.
	ctx := context.Background()
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			err := group.Consume(ctx, []string{topic}, handler)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}()
}
