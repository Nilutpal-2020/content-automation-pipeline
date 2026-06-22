package generator

import (
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
