package kafka

import "time"

type Message struct {
	Key      []byte
	Value    []byte
	Headers  map[string][]byte
	Metadata any

	//
	Topic     string
	Partition int32
	Offset    int64
	Timestamp time.Time
}
