package kafka

import (
	"fmt"

	"github.com/optiopay/kafka/proto"
)

func ExampleConsumer() {
	// connect to kafka cluster
	addresses := []string{"localhost:9092", "localhost:9093"}
	broker, err := Dial(addresses, NewBrokerConf("test"))
	if err != nil {
		panic(err)
	}
	defer broker.Close()

	// create new consumer
	conf := NewConsumerConf("my-messages", 0)
	conf.StartOffset = StartOffsetNewest
	consumer, err := broker.Consumer(conf)
	if err != nil {
		panic(err)
	}

	// read all messages
	for {
		msg, err := consumer.Consume()
		if err != nil {
			if err == ErrNoData {
				break
			}
			panic(err)
		}

		fmt.Printf("message: %#v", msg)
	}
}

func ExampleOffsetCoordinator() {
	// connect to kafka cluster
	addresses := []string{"localhost:9092", "localhost:9093"}
	broker, err := Dial(addresses, NewBrokerConf("test"))
	if err != nil {
		panic(err)
	}
	defer broker.Close()

	// create offset coordinator and customize configuration
	conf := NewOffsetCoordinatorConf("my-consumer-group")
	conf.RetryErrLimit = 20
	coordinator, err := broker.OffsetCoordinator(conf)
	if err != nil {
		panic(err)
	}

	// write consumed message offset for topic/partition
	if err := coordinator.Commit("my-topic", 0, 12); err != nil {
		panic(err)
	}

	// get latest consumed offset for given topic/partition
	off, _, err := coordinator.Offset("my-topic", 0)
	if err != nil {
		panic(err)
	}

	if off != 12 {
		panic(fmt.Sprintf("offset is %d, not 12", off))
	}
}

func ExampleProducer() {
	// connect to kafka cluster
	addresses := []string{"localhost:9092", "localhost:9093"}
	broker, err := Dial(addresses, NewBrokerConf("test"))
	if err != nil {
		panic(err)
	}
	defer broker.Close()

	// create new producer
	conf := NewProducerConf()
	conf.RequiredAcks = proto.RequiredAcksLocal

	// write two messages to kafka using single call to make it atomic
	producer := broker.Producer(conf)
	messages := []*proto.Message{
		{Value: []byte("first")},
		{Value: []byte("second")},
	}
	if _, err := producer.Produce("my-messages", 0, messages...); err != nil {
		panic(err)
	}
}

func ExampleMerge() {
	// connect to kafka cluster
	addresses := []string{"localhost:9092", "localhost:9093"}
	broker, err := Dial(addresses, NewBrokerConf("test"))
	if err != nil {
		panic(err)
	}
	defer broker.Close()

	topics := []string{"fruits", "vegetables"}
	fetchers := make([]Consumer, len(topics))

	// create consumers for different topics
	for i, topic := range topics {
		conf := NewConsumerConf(topic, 0)
		conf.RetryLimit = 20
		conf.StartOffset = StartOffsetNewest
		consumer, err := broker.Consumer(conf)
		if err != nil {
			panic(err)
		}
		fetchers[i] = consumer
	}

	// merge all created consumers (they don't even have to belong to the same broker!)
	mx := Merge(fetchers...)
	defer mx.Close()

	// consume messages from all sources
	for {
		msg, err := mx.Consume()
		if err != nil {
			panic(err)
		}
		fmt.Printf("message: %#v", msg)
	}
}
