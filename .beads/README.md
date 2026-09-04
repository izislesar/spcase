# Beads

This repository uses Beads (`bd`) as its only durable work tracker.

The canonical store is `.beads/issues.jsonl`. There is no separate planning
database or roadmap file.

Start a session with:

```bash
bd prime
bd ready
```

Inspect and claim work with:

```bash
bd show <id>
bd update <id> --claim
```

Close only completed work and include validation evidence:

```bash
bd close <id> --reason="Completed; validation: <commands>"
```

Put blockers, dependencies, follow-up work, and durable project memory in
Beads. Do not create Markdown TODOs, roadmap files, planning documents, or
ad-hoc memory files.
