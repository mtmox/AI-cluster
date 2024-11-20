
package main

import (
	"fmt"
	"log"

	"github.com/mtmox/AI-cluster/constants"
)

func main() {
	err := constants.QueryAndWriteModels()
	if err != nil {
		log.Fatalf("Error querying and writing models: %v", err)
	}
	fmt.Println("Models have been successfully written to the JSON file.")
}
