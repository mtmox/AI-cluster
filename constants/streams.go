
package constants

import (
	"time"

	"github.com/nats-io/nats.go"
)

// StreamConfig represents the configuration for a stream
type StreamConfig struct {
	Name         string
	Subjects     []string
	Retention    nats.RetentionPolicy
	Discard		 nats.DiscardPolicy
	MaxAge       time.Duration
	MaxMsgs		 int64
}

// Streams contains the configurations for all streams
var Streams = []StreamConfig{
	{
		Name: "incoming",
		Subjects: []string{
			"chat.>",
			"generate.>",
		},
		Retention: nats.WorkQueuePolicy,
	},




	{
		Name: "outgoing",
		Subjects: []string{
			"chat.>",
			"generate.>",
		},
		Retention: nats.WorkQueuePolicy,
	},




	{
		Name: "nodes",
		Subjects: []string{
			"config.>",
		},
		Retention: nats.WorkQueuePolicy,
	},
}