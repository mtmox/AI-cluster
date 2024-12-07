
package backend

import (
	"log"

	"github.com/nats-io/nats.go"
)

func StartBackend(js nats.JetStreamContext, logger *log.Logger) {
	ProcessMessage(js, logger)

	select{}
}
