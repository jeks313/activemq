package consumer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	messagesWritten = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "absolute",
			Subsystem: "activemq_archiver",
			Name:      "messages_written_count",
			Help:      "Number of messages written from the topic to the archive files",
		},
		[]string{
			"topic", // what topic this is for
			"key",   // what key we are splitting the files on
		},
	)
	messagesWrittenBytes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "absolute",
			Subsystem: "activemq_archiver",
			Name:      "messages_written_bytes",
			Help:      "Number of bytes written from the topic to the archive files",
		},
		[]string{
			"topic", // what topic this is for
			"key",   // what key we are splitting the files on
		},
	)
)
