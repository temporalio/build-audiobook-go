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

// May be useful to have a variable to configure the server address if someone is running on a different port locally
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

    app.BearerToken = bearerToken

    c, err := client.Dial(client.Options{})
    if err != nil {
        log.Fatalln("unable to create Temporal client", err)
    }
    defer c.Close()

    w := worker.New(c, TaskQueue, worker.Options{})

    w.RegisterWorkflow(app.TTSWorkflow)
    w.RegisterActivity(app.ReadFile)
    w.RegisterActivity(app.CreateTemporaryFile)
    w.RegisterActivity(app.Process)
    w.RegisterActivity(app.MoveOutputFileToPlace)

    err = w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("unable to start Worker", err)
    }
}

// @@@SNIPEND
