package services

import (
	"context"
	"log"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
)

func GetResponseOpenAi() string {
	// API key read from environment
	client := openai.NewClient()

	ctx := context.Background()

	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		// Simple text prompt
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String("Explain Kubernetes in one sentence")},
		Model: openai.ChatModelGPT4_1Mini, // or another model you have access to
	})
	if err != nil {
		log.Fatalln(err)
	}

	return resp.OutputText()
}
