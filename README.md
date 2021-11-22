# Crossjoin

This is the next version of Crossjoin, which is still a work-in-progress.

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
- In the parent directory, run `find * | entr -r go run main.go server`
  - If you don't have entr, run `go run main.go server`
