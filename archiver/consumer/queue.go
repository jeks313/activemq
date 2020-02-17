package consumer

import (
	"context"
	"fmt"

	"github.com/go-stomp/stomp"
	"github.com/jeks313/activemq-archiver/archive"
	"github.com/rs/zerolog/log"
)

// Queue represents an archive queue on a particular topic
type Queue struct {
	Hostname string // activemq hostname: localhost:61613
	Queue    string // queue: queue name, e.g. Archive
	Topic    string // topic: topic name, e.g. MySuperDataTopic
	Ctx      context.Context
	conn     *stomp.Conn
	sub      *stomp.Subscription
}

// Connect sets up the activemq connection
func (q *Queue) Connect(hostname string) error {
	conn, err := stomp.Dial("tcp", hostname)
	if err != nil {
		log.Error().Err(err).Str("hostname", hostname).Msg("failed to connect to activemq")
		return err
	}
	q.conn = conn
	return nil
}

// Subscribe subscribes to the queue
func (q *Queue) Subscribe(topic string) error {
	t := fmt.Sprintf("/queue/Consumer.Archive.VirtualTopic.%s", topic)
	sub, err := q.conn.Subscribe(t, stomp.AckClientIndividual)
	if err != nil {
		log.Error().Err(err).Str("topic", topic).Msg("failed to subscribe to topic")
	}
	q.sub = sub
	return nil
}

// Consume from the queue subscription
func (q *Queue) Consume(arch *archive.Archives) error {
	defer func() {
		log.Info().Msg("queue: unsubscribing")
		q.sub.Unsubscribe()
	}()
	defer func() {
		log.Info().Msg("queue: disconnecting")
		q.conn.Disconnect()
	}()
	for {
		select {
		case <-q.Ctx.Done():
			log.Info().Msg("queue: cancellation received, stopping")
			return nil
		case msg := <-q.sub.C:
			if msg == nil {
				log.Info().Msg("queue: received nil message, stopping")
			}
			err := arch.Write(q.Topic, "key", msg.Body)
			if err != nil {
				log.Error().Err(err).Msg("queue: failed to write document to archive")
				return err
			}
			err = q.conn.Ack(msg)
			if err != nil {
				log.Error().Err(err).Msg("queue: failed to ack message")
				return err
			}
		}
	}
}
