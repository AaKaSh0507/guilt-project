package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Streams struct {
	client *redis.Client
	stream string
}

func NewStreams(client *redis.Client, stream string) *Streams {
	return &Streams{client: client, stream: stream}
}

func (s *Streams) Publish(ctx context.Context, job EntryMLJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: s.stream,
		Values: map[string]interface{}{"job": data},
	}).Result()
	return err
}

func (s *Streams) Consume(ctx context.Context, group, consumer string, timeout time.Duration) ([]EntryMLJob, error) {
	res, err := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{s.stream, ">"},
		Count:    1,
		Block:    timeout,
	}).Result()

	if err != nil {
		return nil, err
	}

	var jobs []EntryMLJob
	for _, stream := range res {
		for _, msg := range stream.Messages {
			raw, ok := msg.Values["job"].(string)
			if !ok {
				continue
			}
			var job EntryMLJob
			if err := json.Unmarshal([]byte(raw), &job); err != nil {
				continue
			}
			jobs = append(jobs, job)

			// ack
			_, _ = s.client.XAck(ctx, s.stream, group, msg.ID).Result()
		}
	}
	return jobs, nil
}

func (s *Streams) EnsureGroup(ctx context.Context, group string) error {
	// create group if missing
	_, err := s.client.XGroupCreateMkStream(ctx, s.stream, group, "$").Result()
	if err != nil && !isBusyGroupError(err) {
		return err
	}
	return nil
}

func isBusyGroupError(err error) bool {
	return err != nil && (len(err.Error()) > 0 && ( // naive
	fmt.Sprintf("%v", err) == "BUSYGROUP Consumer Group name already exists"))
}
