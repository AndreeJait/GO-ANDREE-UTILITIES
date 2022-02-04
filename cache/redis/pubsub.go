package redis

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

type (
	pubsub struct {
		r  *redis.Client
		p  *redis.PubSub
		cn string
	}
)

func (p *pubsub) Receive() error {
	if _, err := p.p.Receive(); err != nil {
		return errors.Wrap(err, "failed to receive")
	}

	return nil
}

func (p *pubsub) Publish(message string) error {
	r := p.r.Publish(p.cn, message)

	if err := r.Err(); err != nil {
		return errors.Wrapf(err, "failed to publish message to cn %s", p.cn)
	}

	return nil
}

func (p *pubsub) Channel() <-chan *redis.Message {
	return p.p.Channel()
}

func (p *pubsub) Close() error {
	if err := p.p.Close(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
