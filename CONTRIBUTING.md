# Contributing to cxas-scrapi-go

Thank you for your interest in contributing! This document outlines the contribution process.

## Prerequisites

- Go 1.22 or higher
- A GCP project with CX Agent Studio enabled (for integration tests)
- `golangci-lint` for local linting

## Development Setup

```bash
git clone https://github.com/Yash-Kavaiya/cxas-scrapi-go.git
cd cxas-scrapi-go
go mod download
```

Build the CLI:

```bash
make build
```

Run all unit tests:

```bash
go test ./...
```

Run the linter:

```bash
make lint
```

## Project Structure

| Directory | Purpose |
|-----------|---------|
| `cmd/cxas` | CLI entry point (`main.go`) |
| `cli/` | Cobra command implementations |
| `pkg/` | Public SDK packages |
| `internal/` | Internal helpers (auth, httpclient, resource, textproto) |

## Making Changes

1. Fork the repository and create a feature branch from `main`.
2. Make your changes, adding tests where appropriate.
3. Ensure `go test ./...` passes.
4. Ensure `go vet ./...` reports no issues.
5. Open a Pull Request against `main`.

## Coding Conventions

- All public types and functions must have GoDoc comments.
- Every new package must include a `doc.go` file with a package-level comment.
- Follow standard Go formatting — run `gofmt -w .` before committing.
- Keep packages small and focused; avoid circular dependencies.
- New packages in `pkg/` must include at least one `_test.go` file.

## Adding a New Package

1. Create the directory under `pkg/<name>`.
2. Add a `doc.go` with the package doc comment.
3. Add a `types.go` for all request/response types.
4. Add the implementation in `<name>.go`.
5. Add tests in `<name>_test.go`.
6. Update the **Package Reference** table in `README.md`.

## Running Integration Tests

Integration tests are tagged with `//go:build integration` and require real GCP credentials:

```bash
export CXAS_OAUTH_TOKEN=$(gcloud auth print-access-token)
go test -tags=integration -v ./...
```

## Reporting Issues

Please open a [GitHub Issue](https://github.com/Yash-Kavaiya/cxas-scrapi-go/issues) with:

- A clear description of the problem
- Steps to reproduce
- Expected vs actual behaviour
- Go version and OS

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
