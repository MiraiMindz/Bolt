package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// TestEvent corresponds to a single line of the `go test -json` output.
// We only need a few fields for this conversion.
type TestEvent struct {
	Action string `json:"Action"`
	Output string `json:"Output,omitempty"`
}

func main() {
	// Create a new JSON decoder that reads from standard input.
	dec := json.NewDecoder(os.Stdin)

	// Loop through the stream of JSON objects.
	for {
		var event TestEvent
		// Decode the next JSON object from the input stream.
		if err := dec.Decode(&event); err != nil {
			// If we've reached the end of the input, break the loop.
			if err == io.EOF {
				break
			}
			// If there's another error, log it and exit.
			log.Fatalf("Error decoding JSON: %v", err)
		}

		// The standard benchmark text is contained in the "Output" field
		// of events where the "Action" is "output".
		if event.Action == "output" {
			// The Output string already contains a newline, so we use fmt.Print.
			fmt.Print(event.Output)
		}
	}
}