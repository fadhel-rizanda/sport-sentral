package messaging

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log/slog"
	apperr "microservice-golang/shared/pkg/errors"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	cfg    Config
	logger *zap.Logger
}

type Config struct {
	URL           string
	MaxReconnects int
	ReconnectWait time.Duration

	// Stream Settings
	StreamName      string
	StreamSubjects  []string
	RetentionMaxAge time.Duration

	// Publish retry settings
	PublishMaxAttempts int
	PublishBaseDelay   time.Duration

	// Consumer Default Settings
	DefaultAckWait       time.Duration
	DefaultMaxDeliver    int
	DefaultMaxAckPending int
}

func Connect(cfg Config, logger *zap.Logger) (*Client, error) {
	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			slog.Warn("NATS disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			slog.Info("NATS reconnected", "url", nc.ConnectedUrl())
		}),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Drain()
		return nil, apperr.Internal(fmt.Errorf("jetstream context: %w", err))
	}

	_, err = js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name:      cfg.StreamName,
		Subjects:  cfg.StreamSubjects,
		MaxAge:    cfg.RetentionMaxAge,
		Retention: jetstream.LimitsPolicy,
		Storage:   jetstream.FileStorage,
		Replicas:  1,
	})
	if err != nil {
		nc.Drain()
		return nil, fmt.Errorf("create/update stream: %w", err)
	}

	slog.Info("NATS JetStream ready", "stream", cfg.StreamName)
	return &Client{
		nc:     nc,
		js:     js,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (c *Client) Publish(ctx context.Context, subject string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var lastErr error
	delay := c.cfg.PublishBaseDelay
	for attempt := 1; attempt <= c.cfg.PublishMaxAttempts; attempt++ {
		var ack *jetstream.PubAck
		ack, lastErr = c.js.Publish(ctx, subject, data)
		if lastErr == nil {
			slog.Debug("published", "subject", subject, "seq", ack.Sequence)
			return nil
		}

		slog.Warn("nats publish failed, retrying",
			"subject", subject,
			"attempt", attempt,
			"max", c.cfg.PublishMaxAttempts,
			"error", lastErr,
		)

		if attempt < c.cfg.PublishMaxAttempts {
			select {
			case <-ctx.Done():
				return fmt.Errorf("publish cancelled after %d attempt(s): %w", attempt, ctx.Err())
			case <-time.After(delay):
				delay *= 2 // 100ms → 200ms → give up
			}
		}
	}

	return fmt.Errorf("publish to %s failed after %d attempts: %w",
		subject, c.cfg.PublishMaxAttempts, lastErr)
}

func (c *Client) Subscribe(
	ctx context.Context,
	streamName, durableName string,
	filterSubjects []string,
	handler func(ctx context.Context, subject string, data []byte) error,
) error {
	cons, err := c.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:        durableName,
		FilterSubjects: filterSubjects,
		AckPolicy:      jetstream.AckExplicitPolicy,
		AckWait:        c.cfg.DefaultAckWait,
		MaxDeliver:     c.cfg.DefaultMaxDeliver,
		DeliverPolicy:  jetstream.DeliverAllPolicy,
		MaxAckPending:  c.cfg.DefaultMaxAckPending,
	})
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		if err := handler(ctx, msg.Subject(), msg.Data()); err != nil {
			slog.Error("handler error – nak", "subject", msg.Subject(), "error", err)
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("start consume: %w", err)
	}
	defer cc.Stop()

	<-ctx.Done()
	return nil
}

func (c *Client) Drain() {
	if c.nc != nil {
		_ = c.nc.Drain()
	}
}
