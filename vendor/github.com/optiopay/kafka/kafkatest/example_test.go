package kafkatest

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/optiopay/kafka"
	"github.com/optiopay/kafka/proto"
)

func ExampleBroker_Producer() {
	broker := NewBroker()
	msg := &proto.Message{Value: []byte("first")}

	producer := broker.Producer(kafka.NewProducerConf())

	// mock server actions, handling any produce call
	go func() {
		resp, err := broker.ReadProducers(time.Millisecond * 20)
		if err != nil {
			panic(fmt.Sprintf("failed reading producers: %s", err))
		}
		if len(resp.Messages) != 1 {
			panic("expected single message")
		}
		if !reflect.DeepEqual(resp.Messages[0], msg) {
			panic("expected different message")
		}
	}()

	// provide data for above goroutine
	_, err := producer.Produce("my-topic", 0, msg)
	if err != nil {
		panic(fmt.Sprintf("cannot produce message: %s", err))
	}

	mockProducer := producer.(*Producer)

	// test error handling by forcing producer to return error,
	//
	// it is possible to manipulate produce result by changing producer's
	// ResponseOffset and ResponseError attributes
	mockProducer.ResponseError = errors.New("my spoon is too big!")
	_, err = producer.Produce("my-topic", 0, msg)
	fmt.Printf("Error: %s\n", err)

	// output:
	//
	// Error: my spoon is too big!
}

func ExampleBroker_Consumer() {
	broker := NewBroker()
	msg := &proto.Message{Value: []byte("first")}

	// mock server actions, pushing data through consumer
	go func() {
		consumer, _ := broker.Consumer(kafka.NewConsumerConf("my-topic", 0))
		c := consumer.(*Consumer)
		// it is possible to send messages through consumer...
		c.Messages <- msg

		// every consumer fetch call is blocking untill there is either message
		// or error ready to return, this way we can test slow consumers
		time.Sleep(time.Millisecond * 20)

		// ...as well as push errors to mock failure
		c.Errors <- errors.New("expected error is expected")
	}()

	// test broker never fails creating consumer
	consumer, _ := broker.Consumer(kafka.NewConsumerConf("my-topic", 0))

	m, err := consumer.Consume()
	if err == nil {
		fmt.Printf("Value: %q\n", m.Value)
	}
	if _, err = consumer.Consume(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	// output:
	//
	// Value: "first"
	// Error: expected error is expected
}

func ExampleServer() {
	// symulate server latency for all fetch requests
	delayFetch := func(nodeID int32, reqKind int16, content []byte) Response {
		if reqKind != proto.FetchReqKind {
			return nil
		}
		time.Sleep(time.Millisecond * 500)
		return nil
	}

	server := NewServer(delayFetch)
	server.MustSpawn()
	defer func() {
		_ = server.Close()
	}()
	fmt.Printf("running server: %s", server.Addr())

	server.AddMessages("my-topic", 0,
		&proto.Message{Value: []byte("first")},
		&proto.Message{Value: []byte("second")})

	// connect to server using broker and fetch/write messages
}
