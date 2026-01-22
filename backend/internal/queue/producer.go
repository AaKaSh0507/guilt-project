package queue

import "context"

type Producer struct {
	streams *Streams
}

func NewProducer(s *Streams) *Producer {
	return &Producer{streams: s}
}

func (p *Producer) Enqueue(ctx context.Context, job EntryMLJob) error {
	return p.streams.Publish(ctx, job)
}
