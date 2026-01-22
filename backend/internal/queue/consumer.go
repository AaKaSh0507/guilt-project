package queue

import (
	"context"
	"time"
)

type Consumer struct {
	streams *Streams
	group   string
	name    string
	timeout time.Duration
}

func NewConsumer(s *Streams, group, name string, timeout time.Duration) *Consumer {
	return &Consumer{
		streams: s,
		group:   group,
		name:    name,
		timeout: timeout,
	}
}

func (c *Consumer) Poll(ctx context.Context) ([]EntryMLJob, error) {
	return c.streams.Consume(ctx, c.group, c.name, c.timeout)
}
