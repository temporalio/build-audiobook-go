// @@@SNIPSTART

package app

import (
    "fmt"
    "log"
    "time"

    "go.temporal.io/sdk/workflow"
)

func TTSWorkflow(ctx workflow.Context, fileInputPath string) (string, error) {
    var message = "Conversion request received"

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
