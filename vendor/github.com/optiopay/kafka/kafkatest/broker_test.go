package kafkatest

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/optiopay/kafka"
	"github.com/optiopay/kafka/proto"
)

func TestBrokerProducer(t *testing.T) {
	broker := NewBroker()

	var wg sync.WaitGroup
	wg.Add(1)
	go readTestMessages(broker, t, &wg)

	producer := broker.Producer(kafka.NewProducerConf())
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go produceTestMessage(producer, t, &wg)
	}
	wg.Wait()
}

func readTestMessages(b *Broker, t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	var i int64
	for i = 1; i <= 20; i++ {
		msg := <-b.produced
		if got := len(msg.Messages); got != 1 {
			t.Fatalf("expected 1 message, got: %d", got)
		}
		m := msg.Messages[0]
		if m.Offset != i {
			t.Errorf("expected offset to be larger: prev: %d, got: %d", i, m.Offset)
		}
	}
}

func produceTestMessage(p kafka.Producer, t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 5; i++ {
		now := time.Now().UnixNano()
		msg := &proto.Message{Value: []byte(fmt.Sprintf("%d", now))}
		_, err := p.Produce("my-topic", 0, msg)
		if err != nil {
			t.Errorf("cannot produce: %s", err)
		}
	}
}
