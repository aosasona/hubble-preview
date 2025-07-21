package llm

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/repository"
)

type LLM struct {
	config *config.LLM
	client *openai.Client
}

const DefaultDimensions = 768

func NewService(conf *config.Config, repo repository.Repository) (*LLM, error) {
	clientConfig := openai.DefaultConfig(conf.LLM.ApiKey)
	clientConfig.BaseURL = conf.LLM.BaseURL

	return &LLM{
		config: &conf.LLM,
		client: openai.NewClientWithConfig(clientConfig),
	}, nil
}

func (l *LLM) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	response, err := l.client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input:          []string{text},
		Model:          openai.EmbeddingModel(l.config.EmbeddingsModel),
		EncodingFormat: openai.EmbeddingEncodingFormatFloat,
		Dimensions:     DefaultDimensions,
	})
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return []float32{}, nil
	}

	return response.Data[0].Embedding, nil
}
