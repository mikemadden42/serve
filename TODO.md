# TODO

Findings from a project review (2026-11-14).

## Bugs

- [x] **Directory listings log status `0`** (`serve.go:63`). `statusRespWr` only records
      a status when `WriteHeader` is called; directory listings never call it. Initialize
      `status: http.StatusOK` or also record it in a `Write` override.
- [x] **`-port` default is unreachable** (`serve.go:18-21`). The `flag.NFlag() == 0` check
      makes plain `serve` print usage and exit 1, so the 8080 default never applies. Drop the
      check or drop the default.
- [x] **`WriteTimeout: 10s` cuts off large downloads** (`serve.go:39`). Any transfer taking
      longer than 10s is killed. Remove it, or rely on `ReadHeaderTimeout` + `IdleTimeout`.

## Security

- [x] **Listens on all interfaces and serves dotfiles.** `":port"` binds every interface and
      `http.FileServer` serves `.git/`, `.env`, etc. Add a `-bind` flag defaulting to
      `127.0.0.1`, and refuse paths with a `.`-prefixed component.
- [x] **Log injection** (gosec G706, `serve.go:63`). `r.URL.Path` is decoded, so `%0A` can
      forge log lines. Log path and user agent with `%q`.

## Code

- [ ] Send errors to stderr (`log.Fatal`) instead of `fmt.Println` to stdout (`serve.go:28`).
- [ ] Response writer wrapper hides `io.ReaderFrom` (loses sendfile). Add `Unwrap()` and a
      `ReadFrom` pass-through.
- [ ] Graceful shutdown on Ctrl-C via `signal.NotifyContext` + `server.Shutdown`.
- [x] Replace the Stack Overflow link header with a real `// Command serve ...` package
      comment (revive `package-comments`).

## Tooling

- [x] **Fix `.golangci.yml` for golangci-lint v2.** `golangci-lint config verify` rejects
      `disable-all`, `linters-settings`, `issues.exclude-dirs`, and numeric `version: 2`
      (must be `"2"`); these are silently ignored. Use `linters.default: none`,
      `linters.settings`, `formatters.settings.goimports.local-prefixes: [...]`,
      `linters.exclusions.paths` — or run `golangci-lint migrate`.
- [x] Bump `go.mod` from `go 1.18` to the minimum version actually supported (CI tests
      oldstable and stable).
- [x] CI: add a `golangci-lint-action` step.
- [ ] Add tests (`go test` currently runs nothing).
- [x] `.goreleaser.yaml`: add `version: 2`; remove the no-op `go generate` hook.
- [x] `.gitignore`: remove stray `sumdir`; ignore the `serve` binary.
- [x] README: document usage and flags.
