
package backend

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func StartBackend(js nats.JetStreamContext, logger *log.Logger) {
	fmt.Printf("Started Backend")
	ProcessMessage(js, logger)
}
