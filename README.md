serve
=====

Simple Go HTTP server for the current directory.

## Install

```sh
go install github.com/mikemadden42/serve@latest
```

## Usage

Run `serve` from the directory you want to share:

```sh
serve                       # http://127.0.0.1:8080
serve -port 9000            # different port
serve -bind 0.0.0.0         # reachable from other machines on the network
```

| Flag    | Default     | Description                                          |
|---------|-------------|------------------------------------------------------|
| `-bind` | `127.0.0.1` | Address to listen on. Use `0.0.0.0` for all interfaces. |
| `-port` | `8080`      | Port to listen on.                                   |

Behavior:

- Dotfiles and dot directories (`.git/`, `.env`, ...) return 404 and are hidden
  from directory listings.
- Each request is logged to stderr with client address, method, path, protocol,
  status, and user agent.

## Development

```sh
make check    # golangci-lint
make build    # goreleaser snapshot build
```
