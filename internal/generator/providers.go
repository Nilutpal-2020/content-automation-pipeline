package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

// The provider structs remain deliberately small so each can be replaced by an
// API-backed implementation without changing the scheduler contract. They
// currently share a deterministic editorial fallback; it produces
// category-aware queue copy rather than a generic placeholder.
type OpenAIGenerator struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}
type ClaudeGenerator struct{ apiKey string }
type GeminiGenerator struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func (o *OpenAIGenerator) RewriteArticle(ctx context.Context, category, title, url, summary string) (*GeneratedContent, error) {
	prompt := fmt.Sprintf(BasePrompt, category, title, url, cleanText(summary, 2_000))
	body, err := json.Marshal(openAIResponseRequest{
		Model: o.model,
		Input: prompt,
		Text: openAITextConfig{Format: openAIJSONSchema{
			Type:   "json_schema",
			Name:   "threads_draft",
			Strict: true,
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"postText":    map[string]string{"type": "string"},
					"hashtags":    map[string]string{"type": "string"},
					"imagePrompt": map[string]string{"type": "string"},
				},
				"required":             []string{"postText", "hashtags", "imagePrompt"},
				"additionalProperties": false,
			},
		}},
		MaxOutputTokens: 500,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal OpenAI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(o.baseURL, "/")+"/responses", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create OpenAI request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call OpenAI: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read OpenAI response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("OpenAI returned %s: %s", resp.Status, truncateForError(string(responseBody), 500))
	}

	var response openAIResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("decode OpenAI response: %w", err)
	}
	content := response.OutputText
	if content == "" {
		content = response.firstOutputText()
	}
	if content == "" {
		return nil, errors.New("OpenAI response did not contain text output")
	}

	var generated GeneratedContent
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		return nil, fmt.Errorf("decode OpenAI generated JSON: %w", err)
	}
	if err := validateGeneratedContent(&generated); err != nil {
		return nil, fmt.Errorf("validate OpenAI generated content: %w", err)
	}
	return &generated, nil
}

type openAIResponseRequest struct {
	Model           string           `json:"model"`
	Input           string           `json:"input"`
	Text            openAITextConfig `json:"text"`
	MaxOutputTokens int              `json:"max_output_tokens"`
}

type openAITextConfig struct {
	Format openAIJSONSchema `json:"format"`
}

type openAIJSONSchema struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type openAIResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (r openAIResponse) firstOutputText() string {
	for _, output := range r.Output {
		for _, content := range output.Content {
			if content.Type == "output_text" && content.Text != "" {
				return content.Text
			}
		}
	}
	return ""
}

func validateGeneratedContent(content *GeneratedContent) error {
	content.PostText = strings.TrimSpace(content.PostText)
	content.Hashtags = strings.TrimSpace(content.Hashtags)
	content.ImagePrompt = strings.TrimSpace(content.ImagePrompt)
	if content.PostText == "" || content.Hashtags == "" || content.ImagePrompt == "" {
		return errors.New("postText, hashtags, and imagePrompt are all required")
	}
	if utf8.RuneCountInString(content.PostText) > 500 {
		return errors.New("postText exceeds the 500 character Threads limit")
	}
	if !strings.Contains(content.Hashtags, "#") {
		return errors.New("hashtags must include at least one hashtag")
	}
	return nil
}

func truncateForError(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}

func (c *ClaudeGenerator) RewriteArticle(ctx context.Context, category, title, url, summary string) (*GeneratedContent, error) {
	return buildEditorialContent(category, title, summary), nil
}

func (g *GeminiGenerator) RewriteArticle(ctx context.Context, category, title, url, summary string) (*GeneratedContent, error) {
	prompt := fmt.Sprintf(BasePrompt, category, title, url, cleanText(summary, 2_000))
	body, err := json.Marshal(geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema: geminiSchemaObject{
				Type: "OBJECT",
				Properties: map[string]geminiSchemaObject{
					"postText":    {Type: "STRING"},
					"hashtags":    {Type: "STRING"},
					"imagePrompt": {Type: "STRING"},
				},
				Required: []string{"postText", "hashtags", "imagePrompt"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal Gemini request: %w", err)
	}

	reqURL := fmt.Sprintf("%s/%s:generateContent?key=%s", strings.TrimRight(g.baseURL, "/"), g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create Gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call Gemini: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read Gemini response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Gemini returned %s: %s", resp.Status, truncateForError(string(responseBody), 500))
	}

	var response geminiResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("decode Gemini response: %w", err)
	}
	
	content := response.firstOutputText()
	if content == "" {
		return nil, errors.New("Gemini response did not contain text output")
	}

	var generated GeneratedContent
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		return nil, fmt.Errorf("decode Gemini generated JSON: %w", err)
	}
	if err := validateGeneratedContent(&generated); err != nil {
		return nil, fmt.Errorf("validate Gemini generated content: %w", err)
	}
	return &generated, nil
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string             `json:"responseMimeType"`
	ResponseSchema   geminiSchemaObject `json:"responseSchema"`
}

type geminiSchemaObject struct {
	Type       string                        `json:"type"`
	Properties map[string]geminiSchemaObject `json:"properties,omitempty"`
	Required   []string                      `json:"required,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (r geminiResponse) firstOutputText() string {
	if len(r.Candidates) > 0 && len(r.Candidates[0].Content.Parts) > 0 {
		return r.Candidates[0].Content.Parts[0].Text
	}
	return ""
}

func buildEditorialContent(category, title, summary string) *GeneratedContent {
	angle, hashtags, visualStyle := categoryStyle(category)
	contextLine := cleanText(summary, 150)
	if contextLine == "" {
		contextLine = "The practical takeaway is worth a closer look."
	}

	post := fmt.Sprintf("%s\n\n%s\n\n%s\n\nWhat would this change in your workflow? 👇", title, contextLine, angle)
	return &GeneratedContent{
		PostText:    post,
		Hashtags:    hashtags,
		ImagePrompt: fmt.Sprintf("Editorial social-media illustration for a %s article titled %q. %s. Show one clear focal subject, a modern technical workspace, rich depth, crisp lighting, and generous negative space for a Threads carousel headline. No logos, no readable text, no watermark.", category, title, visualStyle),
	}
}

func categoryStyle(category string) (angle, hashtags, visualStyle string) {
	switch category {
	case "AI":
		return "The signal isn't the hype—it's how teams turn the capability into something useful.", "#AI #MachineLearning #BuildInPublic", "Visualize an intelligent model as an elegant network of luminous nodes connecting to a builder's workstation, with indigo and electric-cyan accents"
	case "Backend":
		return "Strong systems are built from small engineering decisions made consistently.", "#Backend #SoftwareEngineering #BuildInPublic", "Visualize clean service architecture as flowing data paths around a focused engineer, with charcoal, cobalt, and warm amber accents"
	case "DevOps":
		return "Reliable delivery is a product feature—this is the kind of work users feel without seeing.", "#DevOps #CloudNative #PlatformEngineering", "Visualize a calm operations cockpit with deployment pipelines flowing into resilient cloud infrastructure, with deep navy and emerald accents"
	case "Minimalist":
		return "The clearest systems make room for the work that actually matters.", "#Minimalism #IntentionalWork #MindfulTech", "Visualize an uncluttered desk and a single purposeful digital tool, with soft natural light, warm neutrals, and a calm sense of spaciousness"
	case "Productivity":
		return "Sustainable progress is less about doing more and more about protecting attention.", "#Productivity #DeepWork #BetterHabits", "Visualize a focused creator in a distraction-free workspace, with a simple planning ritual, gentle morning light, and a subtle flow-state atmosphere"
	default:
		return "The best tech stories don't just announce a change—they reveal where the industry is heading.", "#TechNews #Technology #Developers", "Visualize a forward-looking technology newsroom with a single emerging idea illuminated at the center, with midnight blue and violet accents"
	}
}

func cleanText(text string, maxRunes int) string {
	text = strings.Join(strings.Fields(strings.NewReplacer("<p>", " ", "</p>", " ", "<br>", " ", "&quot;", "\"").Replace(text)), " ")
	if utf8.RuneCountInString(text) <= maxRunes {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:maxRunes])) + "…"
}
