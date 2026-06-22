# Content Automation Pipeline - Walkthrough (Notion Pivot)

The Go backend for the Social Media Content Automation Pipeline has been successfully pivoted to use **Notion** as the publishing queue, skipping direct platform publication and making standard databases optional.

## 🏗 Updated System Architecture

- **`/internal/collector`**: Fetches articles using **HackerNews** and **Dev.to**.
- **`/internal/filter`**: The `Scorer` logic weights recency, popularity, and relevance.
- **`/internal/scheduler`**: Fetches all items, deduplicates them by URL, ranks them by score within **AI, Backend, DevOps, Minimalist, Productivity, and Tech News**, and queues the top **3 items per category** daily by default.
- **`/internal/generator`**: Instructs the LLM (OpenAI, Claude, or Gemini) to generate the post and returns structured metadata including the Post Text, Hashtags, and an Image Prompt.
- **`/internal/publisher`**: Uses the official Notion API to create pages in your database with properly mapped properties.

## 🛠 Configuration

Configure your environment via `.env`:

```env
NOTION_TOKEN=your_integration_secret
NOTION_DATABASE_ID=your_database_id

OPENAI_API_KEY=your_key_here # Or ANTHROPIC_API_KEY / GEMINI_API_KEY

MONGO_URI= # Optional
REDIS_ADDR= # Optional
```

## 🚀 How to Run

Because this pipeline now uses `robfig/cron/v3`, simply start the worker process:

```bash
go run ./cmd/worker
```

The pipeline will stay alive in the background and trigger automatically at **8:00 AM daily** based on your server's local time. If `ENV` is not set to `production`, it runs every minute for local testing!

## 📋 Notion Database Structure

Ensure your target Notion database has the following properties created:
- `Title` (Title)
- `Category` (Select)
- `Source` (URL)
- `Generated Post` (Text)
- `Hashtags` (Text)
- `Image Prompt` (Text)
- `Status` (Select)
- `Date` (Date)
- `Content Key` (Text) — required for durable idempotency across scheduler runs.
