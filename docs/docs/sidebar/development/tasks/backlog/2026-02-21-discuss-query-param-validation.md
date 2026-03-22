---
title: Evaluate query param validation strategy
status: backlog
created: 2026-02-21
updated: 2026-02-21
---

## Objective

Decide whether to keep manual query parameter validation or adopt a different
approach.

## Context

oapi-codegen generates `validate:` struct tags from `x-oapi-codegen-extra-tags`
on **request body properties** but NOT on **query parameters**. This means:

- Request bodies: `validation.Struct(request.Body)` works automatically with
  tags like `validate:"required,ip"`.
- Query params: require manual validation in the handler using temporary structs
  with validate tags (current approach).

This creates an inconsistency — body input gets auto-validated, query params
need handwritten validation code that can be missed.

## Options

1. **Keep current approach** (manual validation for query params). Pros:
   idiomatic REST (GET with query params for reads). Well-documented in
   CLAUDE.md. Cons: easy to forget, extra boilerplate.

2. **Use POST with request bodies for complex queries**. Pros: auto-generated
   validate tags, single `validation.Struct()` call. Cons: non-standard REST
   (POST for reads), breaks client expectations.

3. **Write a helper that validates query params from the OpenAPI spec
   constraints**. Pros: avoids manual structs, stays RESTful. Cons: custom code
   to maintain.

4. **Contribute upstream** to oapi-codegen to support validate tags on query
   params.

## Notes

- Current domains affected: job (list), audit (list), system (status).
- The manual approach is documented in CLAUDE.md under "Validation in OpenAPI
  Specs" with explicit examples and integration test requirements.
- Recommendation: keep option 1 for now. The CLAUDE.md documentation and
  integration test requirements mitigate the risk of missing validation.
