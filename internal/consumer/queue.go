package consumer

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-stomp/stomp"
	"github.com/jeks313/activemq-archiver/internal/archive"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
)

// Queue represents an archive queue on a particular topic
type Queue struct {
	Hostname          string                        // activemq hostname: localhost:61613
	Handlers          map[string]ContentTypeHandler // translates data types to JSON byte arrays
	ContentTypeHeader string                        // header to use to control content type
	Queue             string                        // queue: queue name, e.g. Archive
	Topic             string                        // topic: topic name, e.g. MySuperDataTopic
	Headers           []string                      // headers to include into the payload from the message
	Key               string                        // key: what key to partition data on, assumes payload is JSON
	Ctx               context.Context
	conn              *stomp.Conn
	sub               *stomp.Subscription
}

func New() *Queue {
	q := &Queue{}
	q.Handlers = make(map[string]ContentTypeHandler)
	return q
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
	var documentID int64
	for {
		select {
		case <-q.Ctx.Done():
			log.Info().Msg("queue: cancellation received, stopping")
			return nil
		case msg := <-q.sub.C:
			documentID = documentID + 1
			if msg == nil {
				log.Info().Msg("queue: received nil message, stopping")
				return fmt.Errorf("consume: received error")
			}
			if msg.Err != nil {
				log.Error().Err(msg.Err).Msg("consume: received error")
				return msg.Err
			}
			headers := headersFromMessage(q.Headers, msg)
			data, err := q.handleContentType(headers, msg.Body)
			if err != nil {
				log.Error().Err(err).Msg("failed to convert message type")
				err = q.conn.Ack(msg)
				if err != nil {
					log.Error().Err(err).Msg("queue: failed to ack message")
					return err
				}
				continue
			}
			headersAndData, err := sjson.SetBytes(data, "headers", headers)
			headersAndDataNonPretty := pretty.UglyInPlace(headersAndData) // yes some of our payloads are pretty printed, sigh
			if err != nil {
				log.Error().Err(err).Msg("queue: failed to merge headers to payload")
				return err
			}
			keyValue := fromJSON(headersAndDataNonPretty, q.Key)
			if keyValue == "" {
				keyValue = "undef"
			}
			err = arch.Write(q.Topic, keyValue, headersAndDataNonPretty)
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

func (q *Queue) handleContentType(headers map[string]string, data []byte) ([]byte, error) {
	var contentType string
	var ok bool
	if contentType, ok = headers[q.ContentTypeHeader]; ok {
		if handler, ok := q.Handlers[contentType]; ok {
			log.Debug().Str("content_type", contentType).Msg("handling data conversion")
			return handler(data)
		}
		return data, fmt.Errorf("unhandled content type: %v", contentType)
	}
	log.Debug().Str("content_type_header", q.ContentTypeHeader).Msg("no content type header found")
	return data, nil
}

func headersFromMessage(headersWanted []string, msg *stomp.Message) map[string]string {
	headers := make(map[string]string)
	headersToMerge := make(map[string]string)
	for i := 0; i < msg.Header.Len(); i++ {
		header, value := msg.Header.GetAt(i)
		headers[strings.ToLower(header)] = value
	}
	for _, header := range headersWanted {
		if value, ok := headers[strings.ToLower(header)]; ok {
			headersToMerge[strings.ToLower(header)] = value
		}
	}
	return headersToMerge
}

// TODO: funcJSON this needs a little more meat on the bones I think to handle bad JSON and log
// this will just return empty strings if the data in is garbage, which doesn't give a lot of
// visibility into something drifting, like a schema change
func fromJSON(doc []byte, key string) string {
	return gjson.GetBytes(doc, key).String()
}
