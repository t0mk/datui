# datui

Vibe-coded terminal UI for [Datum Cloud](https://datum.net) — browse and manage networking resources in your project.

[Demo video](https://github.com/user-attachments/assets/9fbe4f12-fc81-417a-8ff3-ef804e102ceb)

## Features

- Lists HTTPProxy, Gateway, Domain, Connector, ExportPolicy, DNSRecordSet resources
- Switch resource types with `:`
- YAML detail view with syntax highlighting (`y` or Enter)
- Search in YAML view (`/`, then `n` / `N` to navigate matches)
- Delete resources with `d`

## Requirements

- `~/.kube/config` with a `datum-project-*` context (set automatically by `datumctl`)

## Build & Run

```sh
go build -o datui ./cmd/datui
./datui
```

## Keys

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate resources |
| `Enter` / `y` | Open YAML view |
| `d` | Delete selected resource |
| `:` | Switch resource type |
| `/` | Search in YAML view |
| `n` / `N` | Next / previous match |
| `Esc` | Go back |
| `q` / `Ctrl+C` | Quit |
