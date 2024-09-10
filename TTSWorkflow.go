// @@@SNIPSTART audiobook-project-go-tts-workflow

package app

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

func TTSWorkflow(ctx workflow.Context, fileInputPath string) (string, error) {
	logger := workflow.GetLogger(ctx)
	var a *Activities
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
	// This sample only works if all the activities are run on the same worker, if you have multiple workers they will need to use a sessions or per worker task queue.
	err = workflow.ExecuteActivity(ctx, a.ReadFile, fileInputPath).Get(ctx, &chunks)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	chunkCount := len(chunks)
	logger.Info("File content has %d chunk(s) to process.", chunkCount)

	var tempOutputPath string

	err = workflow.ExecuteActivity(ctx, a.CreateTemporaryFile).Get(ctx, &tempOutputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	logger.Info("Created temporary file for processing: %s", tempOutputPath)

	for index := 0; index < chunkCount; index++ {
		logger.Info("Processing part %d of %d", index+1, chunkCount)
		message = fmt.Sprintf("Processing part %d of %d", index+1, chunkCount)
		err = workflow.ExecuteActivity(ctx, a.Process, chunks[index], tempOutputPath).Get(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to process chunk %d: %w", index+1, err)
		}
	}

	var outputPath string
	err = workflow.ExecuteActivity(ctx, a.MoveOutputFileToPlace, tempOutputPath, fileInputPath).Get(ctx, &outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to move output file: %w", err)
	}

	logger.Info("Output file: %s", outputPath)
	return outputPath, nil
}

// @@@SNIPEND
