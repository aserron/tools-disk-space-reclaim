# ADR-0010: Separation of Runtime Execution Logs (`.log/`) and Produced Artifacts (`.out/`, `.artifacts/`)

* **Status:** Accepted
* **Date:** 2026-11-11
* **Order:** 10 of 10
* **Related Files:** [`.gitignore`](../../.gitignore)

---

## Context

Initially, the repository's `.gitignore` included `.log/*` and stored both runtime logs and generated session walkthroughs/reports (such as `iter_01.md`) inside `.log/`. Mixing runtime execution traces with produced reports, exports, and walkthrough artifacts clutters troubleshooting and blurs the distinction between diagnostic logs and generated deliverables.

---

## Decision

We separated local git-ignored directories into three distinct categories in [`.gitignore`](../../.gitignore):

1. **Runtime Execution Logs (`.log/`, `*.log`):**
   Reserved strictly for runtime execution logs, debug traces, and error reports.
2. **Produced / Generated Files & Artifacts (`.out/`, `.artifacts/`):**
   Dedicated git-ignored directories for produced reports, exported audit tables, iteration walkthroughs (`.out/iter_01.md`), and build/run deliverables.
3. **Ephemeral Scratch & Temporary Files (`.tmp/`, `.scratch/`, `/tmp/`, `.gotmp/`):**
   For short-lived intermediate data and temporary test binaries that can be wiped at any time.

```gitignore
# Runtime execution logs
*.log
.log/

# Produced / generated files & artifacts
.out/
.artifacts/

# Ephemeral scratch & temporary files
.tmp/
.scratch/
/tmp/
```

---

## Consequences

* **Positive:** Keeps `.log/` clean for troubleshooting runtime execution while giving generated reports and iteration summaries a predictable home in `.out/`.
* **Positive:** Standardized `.gitignore` rules also cover Go profiling/coverage outputs (`*.prof`, `coverage.txt`, `coverage.html`), environment files (`.env.*`, `!.env.example`), and OS metadata (`Thumbs.db`, `Desktop.ini`, `.DS_Store`).
