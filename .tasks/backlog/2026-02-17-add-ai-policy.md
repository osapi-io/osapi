---
title: Add AI policy to the project
status: backlog
created: 2026-02-17
updated: 2026-02-17
---

## Objective

Add an AI policy to the project covering how AI tools are used in
development, contribution guidelines for AI-assisted work, and
transparency about AI-generated code.

## Tasks

- [ ] Decide on policy scope (development only, or also covering the
      product itself)
- [ ] Draft AI policy document
- [ ] Add to osapi repo (e.g., `AI_POLICY.md` or section in
      CONTRIBUTING)
- [ ] Add to Docusaurus docs site
- [ ] Add to all other osapi-io repos (nats-client, nats-server,
      osapi-io-justfiles, osapi-io-taskfiles)
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
