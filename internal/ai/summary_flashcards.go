package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

// ChapterSummaryPayload is returned by GenerateChapterSummary (structured JSON).
type ChapterSummaryPayload struct {
	PontosPrincipais []string `json:"pontos_principais"`
	ResumoGeral      string   `json:"resumo_geral"`
	Conceitos        []string `json:"conceitos_importantes"`
}

// FlashcardPair is one Q/A for persistence.
type FlashcardPair struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// GenerateChapterSummary produces structured summary from concatenated chapter text.
func (c *Client) GenerateChapterSummary(ctx context.Context, text string) (*ChapterSummaryPayload, string, error) {
	prompt := `Você é um professor. Resuma o texto abaixo para estudo.
Responda APENAS com JSON válido (sem markdown), neste formato exato:
{"pontos_principais":["..."],"resumo_geral":"...","conceitos_importantes":["..."]}
Use português. Máximo 5 pontos principais e 5 conceitos.

Texto:
` + text

	raw, err := c.chat(ctx, prompt)
	if err != nil {
		return nil, "", err
	}
	payload := extractJSON(raw)
	var out ChapterSummaryPayload
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return nil, "", fmt.Errorf("parse summary JSON: %w", err)
	}
	flat, err := json.Marshal(out)
	if err != nil {
		return nil, "", err
	}
	return &out, string(flat), nil
}

// GenerateFlashcards returns up to maxCards flashcards from page text.
func (c *Client) GenerateFlashcards(ctx context.Context, text string, maxCards int) ([]FlashcardPair, error) {
	if maxCards <= 0 {
		maxCards = 10
	}
	prompt := fmt.Sprintf(`Crie flashcards de pergunta e resposta com base no texto abaixo.
No máximo %d itens. Responda APENAS com array JSON válido, sem markdown:
[{"question":"...","answer":"..."},...]
Use português. Perguntas curtas.

Texto:
%s`, maxCards, text)

	raw, err := c.chat(ctx, prompt)
	if err != nil {
		return nil, err
	}
	payload := extractJSON(raw)
	var items []FlashcardPair
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return nil, fmt.Errorf("parse flashcards JSON: %w", err)
	}
	return items, nil
}
