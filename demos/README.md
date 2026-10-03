# GoCat Demos

This directory contains [VHS](https://github.com/charmbracelet/vhs) tape files
for recording terminal demo GIFs.

## Prerequisites

```bash
brew install vhs   # macOS
# or
go install github.com/charmbracelet/vhs@latest
```

## Recording All Demos

```bash
make demos
# or
./demos/record-all.sh
```

## Individual Recording

```bash
vhs demos/version.tape
vhs demos/console.tape
```

## Tape Files

| Tape | Description | Output |
|------|-------------|--------|
| `version.tape` | Version, doctor, verify commands | `demos/gifs/version.gif` |
| `scan.tape` | Port scanning | `demos/gifs/scan.gif` |
| `console.tape` | Interactive console REPL | `demos/gifs/console.gif` |
| `payload.tape` | Payload generation | `demos/gifs/payload.gif` |
| `listen-connect.tape` | Listen & connect flow | `demos/gifs/listen-connect.gif` |
| `transfer.tape` | File transfer with progress | `demos/gifs/transfer.gif` |
| `stabilize.tape` | Shell stabilization tips | `demos/gifs/stabilize.gif` |
| `session.tape` | Session management | `demos/gifs/session.gif` |
