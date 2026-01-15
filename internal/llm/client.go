package llm

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	apiKey string
	model  *genai.GenerativeModel
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// Generate gera texto usando Gemini 2.0 Flash
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return "", fmt.Errorf("erro ao criar cliente: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash-exp")
	model.SetTemperature(0.7)
	model.SetTopP(0.95)
	model.SetTopK(40)
	model.SetMaxOutputTokens(2048)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("erro ao gerar conteúdo: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("resposta vazia do modelo")
	}

	// Extrair texto da resposta
	text := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return text, nil
}

// GenerateStructured gera JSON estruturado
func (c *Client) GenerateStructured(ctx context.Context, prompt string, schema interface{}) (string, error) {
	// Por enquanto, usa Generate normal
	// TODO: Implementar schema validation
	return c.Generate(ctx, prompt)
}
