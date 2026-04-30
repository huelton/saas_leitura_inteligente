package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 120 * time.Second},
	}
}

type geminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
	GenerationConfig struct {
		Temperature float64 `json:"temperature,omitempty"`
	} `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) chat(ctx context.Context, userPrompt string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("GEMINI_API_KEY is not set")
	}
	body := geminiRequest{}
	body.Contents = []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	}{
		{
			Parts: []struct {
				Text string `json:"text"`
			}{{Text: userPrompt}},
		},
	}
	body.GenerationConfig.Temperature = 0.4
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(c.model), url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("gemini HTTP %d: %s", resp.StatusCode, string(b))
	}
	var out geminiResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", errors.New(out.Error.Message)
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini: empty candidates")
	}
	var sb strings.Builder
	for _, p := range out.Candidates[0].Content.Parts {
		sb.WriteString(p.Text)
	}
	return strings.TrimSpace(sb.String()), nil
}

// GeneratedQuestion is one QA item from the model (levels literal → critical).
type GeneratedQuestion struct {
	Text       string `json:"pergunta"`
	Difficulty string `json:"nivel"`
}

// MCQQuestion represents one multiple-choice question.
type MCQQuestion struct {
	Question   string   `json:"question"`
	Difficulty string   `json:"difficulty"`
	Options    []string `json:"options"`
	CorrectIdx int      `json:"correct_idx"`
}

var fenceRE = regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)```")

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if m := fenceRE.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return s
}

// GenerateQuestions returns up to 5 questions with Bloom-style levels.
func (c *Client) GenerateQuestions(ctx context.Context, text string) ([]GeneratedQuestion, error) {
	prompt := `Você é um professor especialista em interpretação de texto.
Com base APENAS no texto abaixo, gere exatamente 5 perguntas, uma por nível:
1) literal 2) interpretacao 3) explicacao 4) aplicacao 5) critica

Responda somente com um array JSON válido, sem markdown, neste formato:
[{"pergunta":"...","nivel":"literal"},...]

Texto:
` + text

	raw, err := c.chat(ctx, prompt)
	if err != nil {
		return nil, err
	}
	payload := extractJSON(raw)
	var items []GeneratedQuestion
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return nil, fmt.Errorf("parse questions JSON: %w (raw=%s)", err, truncate(raw, 400))
	}
	return items, nil
}

// GenerateQuestionsMCQ returns N multiple-choice questions based on the page text.
func (c *Client) GenerateQuestionsMCQ(ctx context.Context, text string, count int) ([]MCQQuestion, error) {
	if count <= 0 {
		count = 5
	}
	if count > 30 {
		count = 30
	}
	prompt := `Você é um professor especialista em interpretação de texto.
Com base APENAS no texto abaixo, gere exatamente ` + fmt.Sprintf("%d", count) + ` perguntas de múltipla escolha
com níveis progressivos: literal, interpretacao, explicacao, aplicacao, critica.

Retorne SOMENTE um array JSON válido, sem markdown, no formato:
[
  {
    "question":"...",
    "difficulty":"literal",
    "options":["A","B","C","D"],
    "correct_idx":0
  }
]

Regras:
- 4 alternativas por pergunta.
- Apenas 1 correta.
- correct_idx entre 0 e 3.
- Texto em português.
- Não inventar fatos fora do texto.

Texto:
` + text

	raw, err := c.chat(ctx, prompt)
	if err != nil {
		return nil, err
	}
	payload := extractJSON(raw)
	var items []MCQQuestion
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return nil, fmt.Errorf("parse mcq JSON: %w (raw=%s)", err, truncate(raw, 400))
	}
	for i := range items {
		if len(items[i].Options) > 4 {
			items[i].Options = items[i].Options[:4]
		}
		if len(items[i].Options) < 2 {
			items[i].Options = []string{"Opção A", "Opção B", "Opção C", "Opção D"}
		}
		if items[i].CorrectIdx < 0 || items[i].CorrectIdx >= len(items[i].Options) {
			items[i].CorrectIdx = 0
		}
	}
	if len(items) > count {
		items = items[:count]
	}
	return items, nil
}

type evaluationResult struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// EvaluateAnswer returns score 0–10 and short feedback in Portuguese.
func (c *Client) EvaluateAnswer(ctx context.Context, question, answer string) (int, string, error) {
	prompt := fmt.Sprintf(`Você é um professor avaliando a resposta de um aluno.

Pergunta:
%s

Resposta do aluno:
%s

Avalie de 0 a 10 (inteiro) a adequação da resposta ao texto implícito na pergunta.
Responda somente com JSON: {"score":0,"feedback":"..."}
feedback deve ser curto (1-2 frases), em português, construtivo.`,
		question, answer)

	raw, err := c.chat(ctx, prompt)
	if err != nil {
		return 0, "", err
	}
	payload := extractJSON(raw)
	var ev evaluationResult
	if err := json.Unmarshal([]byte(payload), &ev); err != nil {
		return 0, "", fmt.Errorf("parse evaluation JSON: %w", err)
	}
	if ev.Score < 0 {
		ev.Score = 0
	}
	if ev.Score > 10 {
		ev.Score = 10
	}
	return ev.Score, ev.Feedback, nil
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// EnvAPIKey reads GEMINI_API_KEY from the environment.
func EnvAPIKey() string {
	return os.Getenv("GEMINI_API_KEY")
}
