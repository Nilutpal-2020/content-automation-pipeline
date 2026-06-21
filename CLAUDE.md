# Claude Guidelines

When assisting in this repository, follow these specific instructions:

## Code Generation
- Produce fully functional, production-ready Go code.
- Always include `package` declarations and necessary imports.
- Do not output placeholders like `// ... implement later` unless explicitly instructed.
- Stick to standard Go idioms and naming conventions.
- Prefer table-driven testing in your generated test files.

## Workflows
- **Modifying Code**: Always use the provided tools (like file edit tools) to make inline changes rather than dumping large blocks of code in the chat.
- **Dependency Management**: When adding a new library, run `go mod tidy` afterwards to clean up the `go.mod` and `go.sum` files.
- **Formatting**: Make sure to run or assume `gofmt` style for all written code.

## Commands
Common commands used in this project:
- Build: `go build ./cmd/worker`
- Test: `go test ./...`
- Tidy: `go mod tidy`
- Run locally: `go run ./cmd/worker`
- Docker: `docker-compose up -d` (for Redis & MongoDB)

## Style Preferences
- **Logging**: Use structured logging. Ensure log fields contain useful context (e.g. `logger.WithField("source", "hackernews")`).
- **Error Handling**: Use `fmt.Errorf("failed to fetch from %s: %w", source, err)` to wrap errors.
