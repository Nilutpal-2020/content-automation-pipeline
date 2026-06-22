package generator

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestBuildEditorialContentIsCategorySpecific(t *testing.T) {
	content := buildEditorialContent("DevOps", "Kubernetes makes deployments safer", "A practical guide to safer rollouts.")

	if !strings.Contains(content.PostText, "Kubernetes makes deployments safer") {
		t.Errorf("post did not preserve article title: %q", content.PostText)
	}
	if !strings.Contains(content.Hashtags, "#DevOps") {
		t.Errorf("hashtags were not category-specific: %q", content.Hashtags)
	}
	if !strings.Contains(content.ImagePrompt, "DevOps") || !strings.Contains(content.ImagePrompt, "Kubernetes makes deployments safer") {
		t.Errorf("image prompt was not article-specific: %q", content.ImagePrompt)
	}
}

func TestOpenAIGeneratorSendsPromptAndParsesStructuredOutput(t *testing.T) {
	client := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/responses" {
			t.Errorf("path = %q, want /v1/responses", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("authorization = %q", got)
		}
		var request openAIResponseRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(request.Input, "Category:\nProductivity") || !strings.Contains(request.Input, "Title:\nProtect your focus") {
			t.Errorf("prompt did not contain article context: %q", request.Input)
		}
		if request.Text.Format.Type != "json_schema" || !request.Text.Format.Strict {
			t.Errorf("structured output was not requested: %#v", request.Text.Format)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"output_text":"{\"postText\":\"Protect your focus. Small boundaries create big momentum. What will you protect today? 👇\",\"hashtags\":\"#Productivity #DeepWork\",\"imagePrompt\":\"A focused creator in a calm workspace, natural morning light, no text.\"}"}`)),
		}, nil
	})}

	generator := &OpenAIGenerator{apiKey: "test-key", model: "test-model", baseURL: "https://openai.test/v1", client: client}
	content, err := generator.RewriteArticle(context.Background(), "Productivity", "Protect your focus", "https://example.com/article", "A practical guide to reducing distractions.")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content.PostText, "Protect your focus") || content.Hashtags != "#Productivity #DeepWork" {
		t.Errorf("unexpected generated content: %#v", content)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
