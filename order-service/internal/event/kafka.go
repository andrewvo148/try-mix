package event

import (
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
	"order-service/internal/model"
)

type KafkaHandler struct {
	producer sarama.SyncProducer
}

func NewKafkaHandler(brokers []string) (*KafkaHandler, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaHandler{producer: producer}, nil
}

func (h *KafkaHandler) PublishOrderEvent(order *model.Order, eventType string) error {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "orders",
		Key:   sarama.StringEncoder(order.ID.String()),
		Value: sarama.StringEncoder(orderJSON),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(eventType),
			},
		},
	}

	_, _, err = h.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Published order event: %s, OrderID: %s", eventType, order.ID)
	return nil
}

func (h *KafkaHandler) Close() error {
	return h.producer.Close()
} 