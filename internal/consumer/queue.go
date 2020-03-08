package consumer

import (
	"context"
	"fmt"

	"github.com/go-stomp/stomp"
	"github.com/jeks313/activemq-archiver/internal/archive"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

// Queue represents an archive queue on a particular topic
type Queue struct {
	Hostname string // activemq hostname: localhost:61613
	Queue    string // queue: queue name, e.g. Archive
	Topic    string // topic: topic name, e.g. MySuperDataTopic
	Key      string // key: what key to partition data on, assumes payload is JSON
	Ctx      context.Context
	conn     *stomp.Conn
	sub      *stomp.Subscription
}

// Connect sets up the activemq connection
func (q *Queue) Connect(hostname string) error {
	conn, err := stomp.Dial("tcp", hostname)
	if err != nil {
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
		return err
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
				return fmt.Errorf("consume: received error")
			}
			if msg.Err != nil {
				log.Error().Err(msg.Err).Msg("consume: received error")
				return msg.Err
			}
			keyValue := fromJSON(msg.Body, q.Key)
			if keyValue == "" {
				keyValue = "undef"
			}
			err := arch.Write(q.Topic, keyValue, msg.Body)
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

// TODO: funcJSON this needs a little more meat on the bones I think to handle bad JSON and log
// this will just return empty strings if the data in is garbage, which doesn't give a lot of
// visibility into something drifting, like a schema change
func fromJSON(doc []byte, key string) string {
	return gjson.GetBytes(doc, key).String()
}
