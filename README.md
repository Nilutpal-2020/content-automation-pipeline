# Content Automation Pipeline

A background Go worker that automatically fetches trending tech news, scores them based on popularity and recency, uses AI (OpenAI, Claude, or Gemini) to rewrite them into engaging social media posts, and pushes them directly to a Notion Database as a publishing queue.

## 🌟 Features

- **Automated Collection**: Fetches daily content from HackerNews and Dev.to.
- **Intelligent Scoring**: Ranks articles based on recency, popularity, and keyword relevance.
- **AI Rewrites**: Automatically generates the post body, relevant hashtags, and a suggested image prompt using your choice of LLM (OpenAI, Claude, or Gemini).
- **Notion Integration**: Directly queues the top 10 generated posts into a Notion database.
- **Cron Scheduling**: Built-in `cron` scheduling to run the pipeline automatically every day at 8:00 AM.
- **Optional Databases**: Support for MongoDB and Redis if you want to expand the architecture for complex queuing and historical data storage (completely optional).

## 🏗 Architecture

- **`internal/collector`**: API fetchers for data sources.
- **`internal/filter`**: The scoring and ranking engine.
- **`internal/generator`**: The AI factory that routes to your provided LLM.
- **`internal/publisher`**: The Notion API wrapper to create database pages.
- **`internal/scheduler`**: The orchestrator that ties the entire pipeline together.

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- A Notion Integration Token & Database ID
- An API Key for OpenAI, Anthropic, or Google Gemini.

### 1. Configure the Environment
Create a `.env` file in the root directory (you can copy `.env.example`):
```bash
cp .env.example .env
```
Fill in your API keys and Notion details.

### 2. Set up Notion Database
Ensure your Notion database has the following properties:
- `Title` (Title)
- `Category` (Select)
- `Source` (URL)
- `Generated Post` (Text)
- `Hashtags` (Text)
- `Image Prompt` (Text)
- `Status` (Select)
- `Date` (Date)

**Critical:** Make sure to click the `...` menu on your database page, go to "Connections", and explicitly share the database with your Notion Integration!

### 3. Run the Pipeline
Start the background worker:
```bash
go run ./cmd/worker
```

When `ENV` is not set to `production`, the worker will run every minute for testing. In `production`, it runs daily at 8:00 AM.

## 🛠 Advanced Usage
If you want to use MongoDB and Redis, simply provide `MONGO_URI`, `MONGO_DB_NAME`, and `REDIS_ADDR` in your `.env`. The worker will automatically connect to them. A `docker-compose.yml` is included to easily spin these up locally:
```bash
docker-compose up -d
```

## 📜 Agent Guidelines
If you are an AI assistant working on this repository, please read the `AGENTS.md` and `CLAUDE.md` files for structural rules and coding guidelines.
