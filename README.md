# pencap

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-00ADD8.svg)

A pentest evidence and reporting helper. Captures screenshots and files into
a per-engagement evidence vault, tracks findings, and exports a markdown
report — all from the command line.

## Contents

- [Features](#features)
- [Requirements](#requirements)
- [Install](#install)
- [Usage](#usage)
- [scope.yaml](#scopeyaml)
- [Development](#development)
- [Security](#security)
- [License](#license)

## Features

- Screenshot and file capture into a per-engagement, per-category evidence tree
- sha256 sidecar on every captured file for chain-of-custody
- Timestamped notes, keyed by category
- Lightweight findings tracker (plain markdown, no database)
- One-command markdown report export
- Cross-platform, obfuscated release builds (`garble`) via `make release`

## Requirements

- Go 1.22+ to build from source
- `pencap ss <type>` (screenshot capture) shells out to a native tool:
  - macOS: `screencapture` (built in, no setup needed)
  - Linux: one of `scrot`, `gnome-screenshot`, or ImageMagick's `import`
  - Windows: not supported yet — use `pencap ss file <path>` to import a
    screenshot taken by another tool instead

## Install

```
make install
```

Builds and installs to `$(go env GOPATH)/bin`. Make sure that directory is
on your `PATH`.

## Usage

```
pencap init <engagement-name>          scaffold a new engagement folder
pencap ss <type> [note...]             capture a screenshot into evidence/<type>/
pencap ss file <src-path> [note...]    copy a file into evidence/files/
pencap note <type> <text...>           append a timestamped note
pencap finding add <title> [--severity crit|high|med|low|info]
pencap finding link <id> <evidence-path>
pencap finding list
pencap export                          write findings + evidence to report/draft/report.md
```

Every `init` scaffolds a `scope.yaml` marker, so `ss`/`note`/`finding`/`export`
work from any subdirectory of that engagement — no need to `cd` back to the root.

## scope.yaml

`pencap init <name>` generates this template at the engagement root:

```yaml
# scope.yaml - engagement scope and rules of engagement
engagement: acme-2026
client: ""
start_date: ""
end_date: ""
in_scope:
  - ""
out_of_scope:
  - ""
contacts:
  - ""
```

Fill it in by hand, e.g.:

```yaml
# scope.yaml - engagement scope and rules of engagement
engagement: acme-2026
client: "Acme Corp"
start_date: "2026-07-10"
end_date: "2026-07-24"
in_scope:
  - "*.acme.com"
  - "10.20.30.0/24"
out_of_scope:
  - "billing.acme.com"
  - "corporate VPN infrastructure"
contacts:
  - "Jane Doe <jane@acme.com> (primary technical contact)"
```

It currently doubles as pencap's own engagement-root marker (how `ss`/`note`/
`finding`/`export` find the right folder from a subdirectory) — the
`in_scope`/`out_of_scope` fields are documentation only, not yet enforced
against captured evidence.

## Development

```
make test    # unit tests
make lint    # vet + staticcheck + golangci-lint
make audit   # gosec + govulncheck
make release # cross-compiled, garble-obfuscated binaries in dist/
```

See `make help` for the full target list. Before opening a PR, make sure
`make lint` and `make audit` are both clean and `make test` passes.

## Security

pencap is a local CLI that writes to disk with your own user permissions —
it has no network listener and no remote attack surface. See
[SECURITY.md](SECURITY.md) for how to report a vulnerability.

## License

MIT — see [LICENSE](LICENSE).
