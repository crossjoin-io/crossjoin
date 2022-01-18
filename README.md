# Crossjoin [![Docker](https://github.com/crossjoin-io/crossjoin/actions/workflows/docker.yml/badge.svg)](https://github.com/crossjoin-io/crossjoin/actions/workflows/docker.yml) [![CLI](https://github.com/crossjoin-io/crossjoin/actions/workflows/go.yml/badge.svg)](https://github.com/crossjoin-io/crossjoin/actions/workflows/go.yml) [![Security scan](https://github.com/crossjoin-io/crossjoin/actions/workflows/shiftleft.yml/badge.svg)](https://github.com/crossjoin-io/crossjoin/blob/main/SECURITY.md)

Crossjoin is a service to run data-driven workflows.
It joins together data from various data sources and triggers Docker-based workflows.
Workflows are defined as YAML (like GitHub Actions) and are executed by runners.
You can run everything in a single Crossjoin instance, or have 1 server and multiple
runners.

## Status

Crossjoin is under active development. Let @Preetam know if you're interested in using it!

## License

Apache 2.0

## Building

**Requirements**

- Go
- Node.js, NPM

```
cd ui && npm install && npm run build && \
cd .. && go build -o crossjoin
```

Everything will be embedded in the `crossjoin` binary.

## Development

**Requirements**

- Go
- Node.js, NPM
- [entr(1)](https://eradman.com/entrproject/) is useful, but not required

Running:

- In the `ui` directory, run `npm start`
- In the parent directory, run `find . -path './ui/node_modules' -prune -o -name '*.js' -o -name '*.go' | entr -r go run main.go server --runner --config config/example.yml`
  - If you don't have entr, run `go run main.go server --runner --config config/example.yml`
