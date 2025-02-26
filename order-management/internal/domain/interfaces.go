package domain

// EventProducer defines an interface for any messaging system
type EventProducer interface {
	Publish(topic string, key []byte, vslue []byte) error
}

type ConsumerEvent interface {
	
}
