---
title: Diagrams & Assets
description: Source, generation, and commit guidelines for architecture diagrams.
last_reviewed: 2025-09-08
---

# Diagrams & Assets

Place diagram sources in `docs/diagrams-src/.mmd.puml`) and commit generated PNG/SVG files into `docs/assets/` with the same basename. Include the generation command in the source header or this file.

```bash
(Mermaid
```

```bash
or PlantUML
```

Example generation commands:

```bash
# Mermaid (node-based):
npx @mermaid-js/mermaid-cli -i docs/diagrams-src/architecture.mmd -o docs/assets/architecture_overview.png

# PlantUML (jar):
java -jar plantuml.jar -tpng docs/diagrams-src/container.puml -o docs/assets/
```

Keep sources small and focused. Prefer vector (`.svg`) for diagrams where possible. When updating a diagram, update both source and generated artifact in the same PR.

Sample source file: `docs/diagrams-src/architecture.mmddocs/assets/architecture_overview.png`).

```bash
(renders to
```
