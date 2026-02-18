---
title: Add AI policy to the project
status: done
created: 2026-02-17
updated: 2026-02-18
---

## Objective

Add an AI policy to the project covering how AI tools are used in
development, contribution guidelines for AI-assisted work, and
transparency about AI-generated code.

## Tasks

- [x] Decide on policy scope (development only, or also covering the
      product itself)
- [x] Draft AI policy document
- [x] Add to osapi repo (e.g., `AI_POLICY.md` or section in
      CONTRIBUTING)
- [ ] Add to Docusaurus docs site
- [x] Add to all other osapi-io repos (nats-client, nats-server,
      osapi-io-justfiles) — excluded osapi-io-taskfiles per user request
- [ ] Update PR template to include AI disclosure if applicable
- [ ] Consider adding to LICENSE or NOTICE file if needed

## Notes

- Project already uses Claude Code for development (commits tagged with
  `Co-Authored-By: Claude`)
- Policy should cover: acceptable use of AI in contributions, disclosure
  requirements, review expectations for AI-generated code, IP/licensing
  considerations
- Use Ghostty's AI policy as the basis:
  https://github.com/ghostty-org/ghostty/blob/main/AI_POLICY.md
- Must be applied consistently across all repos in the osapi-io
  organization, not just the main osapi repo

## Outcome

Created `AI_POLICY.md` adapted from Ghostty's policy for osapi-io.
Deployed identical copies to 4 repos:

- `osapi/AI_POLICY.md`
- `nats-client/AI_POLICY.md`
- `nats-server/AI_POLICY.md`
- `osapi-io-justfiles/AI_POLICY.md`

Key adaptations from Ghostty's original:
- References osapi-io project name instead of Ghostty
- Mentions existing `Co-Authored-By` commit trailer convention
- Replaced "public denouncement list" with simpler "blocked" language
- Kept same structure: disclosure, understanding, human review, no AI
  media, bad drivers blocked, maintainer exemption

Remaining items (Docusaurus docs, PR template, LICENSE) left as future
backlog if desired.
