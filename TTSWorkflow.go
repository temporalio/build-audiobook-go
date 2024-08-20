// @@@SNIPSTART audiobook-project-go-tts-workflow

package app

import (
    "fmt"
    "log"
    "time"

    "go.temporal.io/sdk/workflow"
)

func TTSWorkflow(ctx workflow.Context, fileInputPath string) (string, error) {
    var message = "Conversion request received"
	// Generally we recommend users define QueryType as a global constant so other code can use it, for this example it is probably OK here.
    queryType := "fetchMessage"

    err := workflow.SetQueryHandler(ctx, queryType, func() (string, error) {
        return message, nil
    })

    if err != nil {
        message = "Failed to register query handler"
        return "", err // Return an empty string and the error
    }

    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 120 * time.Second,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    var chunks []string
	// This sample only works if all the activities are run on the same worker, if you have multiple workers they will need to use a sessions or per worker task queue.
	// I understand if the goal of the sample isn't to show sessions or multiple workers, but it is something we should call out clearly.
    err = workflow.ExecuteActivity(ctx, ReadFile, fileInputPath).Get(ctx, &chunks)
    if err != nil {
        return "", fmt.Errorf("failed to read file: %w", err)
    }

    chunkCount := len(chunks)
    log.Printf("File content has %d chunk(s) to process.", chunkCount)

    var tempOutputPath string

    err = workflow.ExecuteActivity(ctx, CreateTemporaryFile).Get(ctx, &tempOutputPath)
    if err != nil {
        return "", fmt.Errorf("failed to create temporary file: %w", err)
    }
	// We should not use `log` in a workflow. users should call log := workflow.GetLogger(ctx) and use that logger.
    log.Printf("Created temporary file for processing: %s", tempOutputPath)

    for index := 0; index < chunkCount; index++ {
        log.Printf("Processing part %d of %d", index + 1, chunkCount)
        message = fmt.Sprintf("Processing part %d of %d", index + 1, chunkCount)
        err = workflow.ExecuteActivity(ctx, Process, chunks[index], tempOutputPath).Get(ctx, nil)
        if err != nil {
            return "", fmt.Errorf("failed to process chunk %d: %w", index + 1, err)
        }
    }

    var outputPath string
    err = workflow.ExecuteActivity(ctx, MoveOutputFileToPlace, tempOutputPath, fileInputPath).Get(ctx, &outputPath)
    if err != nil {
        return "", fmt.Errorf("failed to move output file: %w", err)
    }

    log.Printf("Output file: %s", outputPath)
    return outputPath, nil
}

// @@@SNIPEND
