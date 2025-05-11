package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/notblessy/anggar-service/model"
	"github.com/oklog/ulid/v2"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type recognizerRepository struct {
	openAi *openai.Client
}

func NewHandler(openAi *openai.Client) model.RecognizerRepository {
	return &recognizerRepository{
		openAi: openAi,
	}
}

func (r *recognizerRepository) RecognizeTransaction(ctx context.Context, withPrompt, text string) (model.Transaction, error) {
	logger := logrus.WithContext(ctx).WithField("text", text)

	resp, err := r.openAi.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: withPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
		},
	)

	if err != nil {
		logger.Error(fmt.Errorf("failed to create chat completion: %w", err))
		return model.Transaction{}, err
	}

	var recognized model.Transaction

	recognized.CreatedAt = time.Now()

	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &recognized)
	if err != nil {
		logger.Error(fmt.Errorf("failed to unmarshal recognized: %w", err))

	}

	recognized.ID = ulid.Make().String()

	return recognized, nil
}
