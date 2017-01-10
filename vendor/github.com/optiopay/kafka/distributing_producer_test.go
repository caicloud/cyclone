package kafka

import (
	"fmt"
	"sync"
	"testing"

	"github.com/optiopay/kafka/proto"
)

type recordingProducer struct {
	sync.Mutex
	msgs []*proto.Message
}

func newRecordingProducer() *recordingProducer {
	return &recordingProducer{msgs: make([]*proto.Message, 0)}
}

func (p *recordingProducer) Produce(topic string, part int32, msgs ...*proto.Message) (int64, error) {
	p.Lock()
	defer p.Unlock()

	offset := len(p.msgs)
	p.msgs = append(p.msgs, msgs...)
	for i, msg := range msgs {
		msg.Offset = int64(offset + i)
		msg.Topic = topic
		msg.Partition = part
	}
	return int64(len(p.msgs)), nil
}

func TestRoundRobinProducer(t *testing.T) {
	rec := newRecordingProducer()
	p := NewRoundRobinProducer(rec, 3)

	data := [][][]byte{
		{
			[]byte("a 1"),
			[]byte("a 2"),
		},
		{
			[]byte("b 1"),
		},
		{
			[]byte("c 1"),
			[]byte("c 2"),
			[]byte("c 3"),
		},
		{
			[]byte("d 1"),
		},
	}

	for i, values := range data {
		msgs := make([]*proto.Message, len(values))
		for i, value := range values {
			msgs[i] = &proto.Message{Value: value}
		}
		if _, err := p.Distribute("test-topic", msgs...); err != nil {
			t.Errorf("cannot distribute %d message: %s", i, err)
		}
	}

	// a, [0, 1]
	if rec.msgs[0].Partition != 0 || rec.msgs[1].Partition != 0 {
		t.Fatalf("expected partition 0, got %d and %d", rec.msgs[0].Partition, rec.msgs[1].Partition)
	}

	// b, [2]
	if rec.msgs[2].Partition != 1 {
		t.Fatalf("expected partition 1, got %d", rec.msgs[2].Partition)
	}

	// c, [3, 4, 5]
	if rec.msgs[3].Partition != 2 || rec.msgs[4].Partition != 2 {
		t.Fatalf("expected partition 2, got %d and %d", rec.msgs[3].Partition, rec.msgs[3].Partition)
	}

	// d, [6]
	if rec.msgs[6].Partition != 0 {
		t.Fatalf("expected partition 0, got %d", rec.msgs[6].Partition)
	}
}

func TestHashProducer(t *testing.T) {
	const parts = 3
	rec := newRecordingProducer()
	p := NewHashProducer(rec, parts)

	var keys [][]byte
	for i := 0; i < 30; i++ {
		keys = append(keys, []byte(fmt.Sprintf("key-%d", i)))
	}
	for i, key := range keys {
		msg := &proto.Message{Key: key}
		if _, err := p.Distribute("test-topic", msg); err != nil {
			t.Errorf("cannot distribute %d message: %s", i, err)
		}
	}

	if len(rec.msgs) != len(keys) {
		t.Fatalf("expected %d messages, got %d", len(keys), len(rec.msgs))
	}

	for i, key := range keys {
		want, err := messageHashPartition(key, parts)
		if err != nil {
			t.Errorf("cannot compute hash: %s", err)
			continue
		}
		if got := rec.msgs[i].Partition; want != got {
			t.Errorf("expected partition %d, got %d", want, got)
		} else if got > parts-1 {
			t.Errorf("number of partitions is %d, but message written to %d", parts, got)
		}
	}
}

func TestRandomProducerIsConcurrencySafe(t *testing.T) {
	const workers = 100

	p := NewRandomProducer(nullproducer{}, 4)
	start := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(workers)

	// spawn worker, each starting to produce at the same time
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			msg := &proto.Message{Value: []byte("value")}

			<-start

			for n := 0; n < 1000; n++ {
				if _, err := p.Distribute("x", msg); err != nil {
					t.Errorf("cannot distribute: %s", err)
				}
			}
		}()
	}

	close(start)
	wg.Wait()
}

type nullproducer struct{}

func (nullproducer) Produce(topic string, part int32, msgs ...*proto.Message) (int64, error) {
	return 0, nil
}
