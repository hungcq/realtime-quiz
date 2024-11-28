package event_publisher

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"quiz/configs"
)

var producer sarama.SyncProducer

func init() {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	var err error
	producer, err = sarama.NewSyncProducer(configs.KafkaBrokerAddress, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}
}

func Publish(topic string, key string, data any) error {
	msgValue, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(msgValue),
	})
	return err
}
