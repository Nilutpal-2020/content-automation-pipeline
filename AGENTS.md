# AI Agent Guidelines

This repository contains the backend for a Social Media Content Automation Pipeline. When working in this repository, please adhere to the following guidelines:

## Tech Stack
- **Language**: Go
- **Scheduler**: Standard cron/worker loop
- **LLM**: Integration with OpenAI / Claude / Gemini
- **Publishing**: Notion API

## Project Architecture
- `/cmd/worker`: Main entry point for the background worker.
- `/internal/collector`: RSS, Reddit, HackerNews, GitHub fetching logic.
- `/internal/filter`: Logic to score content based on popularity, recency, and relevance.
- `/internal/generator`: AI integrations to rewrite articles into bite-sized content (Threads).
- `/internal/publisher`: API wrappers for social platforms (Notion).
- `/internal/scheduler`: Cron setups that trigger collections and publishing.
- `/pkg/logger`: Structured logging.
- `/pkg/config`: Environment variable parsing.

## Code Standards
- Follow standard Go conventions (`gofmt`, idiomatic naming).
- Use structured logging instead of `fmt.Println` or `log.Printf`.
- Error handling must be explicit. Wrap errors with context where applicable.
- Keep components modular. Handlers/collectors should define interfaces that can be easily mocked for testing.
- Write table-driven tests for pure logic components (like filters/scorers).

## Execution
- Ensure environment variables are loaded properly using `.env` for local development.
- Before suggesting new external dependencies, verify if standard library options are sufficient.
