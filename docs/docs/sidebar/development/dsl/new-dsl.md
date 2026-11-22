---
title: Build a new DSL
sidebar_position: 6
---

# Build a new DSL

The guide to adding a DSL to osapi, or to changing what the three DSLs share, is
the `build-a-dsl` skill in the binary. The binary is its only source, so an
agent and a person read the same text:

```bash
osapi skill get --name build-a-dsl
osapi skill get --name build-a-dsl --reference references/parts.md
```

The skill names each part that the DSLs share: the spec and the checks tables,
validate, typed values and the tag keys, the permissions that the role rule
reads, the generators, the scaffold command, the docs page and the skill of a
DSL. `references/parts.md` tells where each part is and what a new DSL does for
it. [Developing a skill](../skills.md) tells how the skills are kept.
