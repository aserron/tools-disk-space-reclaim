# `space-reclaim` Documentation Hub

Welcome to the central documentation directory for **`space-reclaim`** (`E:\aserron\tools\disk\space-reclaim`).

---

## Documentation Directory Structure

```text
docs/
├── README.md                                 # Central documentation index (this file)
├── adr/                                      # Architectural Decision Records (in chronological order)
│   ├── README.md                             # Chronological ADR index & Mermaid decision timeline
│   ├── 0001-domain-scoped-repository-layout.md
│   ├── 0002-go-stdlib-first-orchestrator-architecture.md
│   ├── 0003-issue-driven-ground-up-delivery.md
│   ├── 0004-three-tier-risk-model-and-protected-reserved-storage.md
│   ├── 0005-injected-command-runner-and-wsl-utf16le-decoding.md
│   ├── 0006-concurrent-reparse-safe-filesystem-scanner.md
│   ├── 0007-vhdx-compaction-pipeline-with-diskpart-fallback.md
│   ├── 0008-gotmpdir-workspace-isolation-for-app-control.md
│   ├── 0009-interactive-ansi-tui-selector-and-cli-automation.md
│   └── 0010-separation-of-runtime-logs-and-generated-artifacts.md
├── wiki/                                     # User help docs, CLI guides & migration guides
│   ├── Home.md                               # Wiki home & navigation
│   ├── 01-quickstart-and-installation.md     # Build, $PROFILE aliases, and first scan
│   ├── 02-interactive-tui-guide.md           # TUI selector commands (ranges, stages, presets)
│   ├── 03-cli-flags-and-automation.md        # Complete CLI flag reference & automation recipes
│   ├── 04-stages-reference.md                # Deep-dive into all 6 stages (wsl..system)
│   ├── 05-safety-model-and-risk-levels.md    # Risk tiers, Admin guards & Reserved Storage policy
│   └── 06-chocolatey-to-scoop-migration.md   # Audit & step-by-step migration from Choco to Scoop
├── architecture/                             # System internals & dataflow diagrams
│   └── README.md                             # Component diagrams, sequence diagrams & package map
├── runbooks/                                 # Step-by-step operational playbooks
│   └── README.md                             # Emergency C:\ recovery, VHDX compaction & Choco->Scoop
└── development/                              # Contributor & extension guide
    └── README.md                             # Adding new caches/stages, DryRun testing & build.ps1
```

---

## Quick Links

* **Architectural Decisions (Chronological):** [`docs/adr/README.md`](adr/README.md)
* **Help Wiki Home:** [`docs/wiki/Home.md`](wiki/Home.md)
* **Chocolatey to Scoop Migration Guide:** [`docs/wiki/06-chocolatey-to-scoop-migration.md`](wiki/06-chocolatey-to-scoop-migration.md)
* **System Architecture & Diagrams:** [`docs/architecture/README.md`](architecture/README.md)
* **Operational Runbooks:** [`docs/runbooks/README.md`](runbooks/README.md)
* **Development & Extending Guide:** [`docs/development/README.md`](development/README.md)
