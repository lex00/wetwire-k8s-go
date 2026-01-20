# Contributing to wetwire-k8s-go

Thank you for your interest in contributing to wetwire-k8s-go!

## Getting Started

See the [Developer Guide](docs/DEVELOPERS.md) for:
- Development environment setup
- Project structure
- Running tests

## Code Style

- **Formatting**: Use `gofmt` (automatic in most editors)
- **Linting**: Use `go vet` and `golangci-lint`
- **Imports**: Use `goimports` for automatic import management

```bash
# Format code
gofmt -w .

# Lint
go vet ./...
golangci-lint run ./...

# Check for common issues
go build ./...
```

## Commit Messages

Follow conventional commits:

```
feat: Add support for StatefulSet resources
fix: Correct namespace handling in selector
docs: Update installation instructions
test: Add tests for lint rules
chore: Update dependencies
```

## Pull Request Process

1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes with tests
3. Run tests: `go test ./...`
4. Run linter: `golangci-lint run ./...`
5. Commit with clear messages
6. Push and open Pull Request
7. Address review comments

## Adding a New Lint Rule

1. Add rule to `internal/lint/rules.go`
2. Implement the check function
3. Add test case in `internal/lint/rules_test.go`
4. Update docs/LINT_RULES.md with the new rule
5. Update CLAUDE.md if it affects syntax guidance

Lint rules use the `WK8xxx` prefix. See [docs/LINT_RULES.md](docs/LINT_RULES.md) for the complete rule reference with category ranges.

## Adding a New CLI Command

1. Create command file in `cmd/wetwire-k8s/`
2. Implement using Cobra command pattern
3. Register in `main.go` with `rootCmd.AddCommand()`
4. Add tests
5. Update docs/CLI.md documentation

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include reproduction steps for bugs
- Check existing issues before creating new ones

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
