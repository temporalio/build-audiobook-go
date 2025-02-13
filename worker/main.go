// @@@SNIPSTART audiobook-project-go-worker
package main

import (
	"audiobook/app"
	"log"
	"os"
	"strings"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const TaskQueue = "tts-task-queue"

func main() {
	bearerToken := os.Getenv("OPEN_AI_BEARER_TOKEN")
	if bearerToken == "" {
		log.Fatalln("Environment variable OPEN_AI_BEARER_TOKEN not found")
	}

	bearerToken = strings.TrimSpace(bearerToken)
	bearerToken = strings.Map(func(r rune) rune {
		if r >= 32 && r <= 126 { // Printable characters range
			return r
		}
		return -1
	}, bearerToken)

	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	w := worker.New(c, TaskQueue, worker.Options{})

	w.RegisterWorkflow(app.TTSWorkflow)
	w.RegisterActivity(&app.Activities{BearerToken: bearerToken})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}

// @@@SNIPEND
