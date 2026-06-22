package generator

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

// The provider structs remain deliberately small so each can be replaced by an
// API-backed implementation without changing the scheduler contract. They
// currently share a deterministic editorial fallback; it produces
// category-aware queue copy rather than a generic placeholder.
type OpenAIGenerator struct{ apiKey string }
type ClaudeGenerator struct{ apiKey string }
type GeminiGenerator struct{ apiKey string }

func (o *OpenAIGenerator) RewriteArticle(ctx context.Context, category, title, url, summary string) (*GeneratedContent, error) {
	return buildEditorialContent(category, title, summary), nil
}

func (c *ClaudeGenerator) RewriteArticle(ctx context.Context, category, title, url, summary string) (*GeneratedContent, error) {
	return buildEditorialContent(category, title, summary), nil
}

func (g *GeminiGenerator) RewriteArticle(ctx context.Context, category, title, url, summary string) (*GeneratedContent, error) {
	return buildEditorialContent(category, title, summary), nil
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
		return "The signal isn’t the hype—it’s how teams turn the capability into something useful.", "#AI #MachineLearning #BuildInPublic", "Visualize an intelligent model as an elegant network of luminous nodes connecting to a builder’s workstation, with indigo and electric-cyan accents"
	case "Backend":
		return "Strong systems are built from small engineering decisions made consistently.", "#Backend #SoftwareEngineering #BuildInPublic", "Visualize clean service architecture as flowing data paths around a focused engineer, with charcoal, cobalt, and warm amber accents"
	case "DevOps":
		return "Reliable delivery is a product feature—this is the kind of work users feel without seeing.", "#DevOps #CloudNative #PlatformEngineering", "Visualize a calm operations cockpit with deployment pipelines flowing into resilient cloud infrastructure, with deep navy and emerald accents"
	case "Minimalist":
		return "The clearest systems make room for the work that actually matters.", "#Minimalism #IntentionalWork #MindfulTech", "Visualize an uncluttered desk and a single purposeful digital tool, with soft natural light, warm neutrals, and a calm sense of spaciousness"
	case "Productivity":
		return "Sustainable progress is less about doing more and more about protecting attention.", "#Productivity #DeepWork #BetterHabits", "Visualize a focused creator in a distraction-free workspace, with a simple planning ritual, gentle morning light, and a subtle flow-state atmosphere"
	default:
		return "The best tech stories don’t just announce a change—they reveal where the industry is heading.", "#TechNews #Technology #Developers", "Visualize a forward-looking technology newsroom with a single emerging idea illuminated at the center, with midnight blue and violet accents"
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
