# QMD

QMD (Query Markup Documents) is the local document index for this repository. Use it to query requirements and design documents before implementing features or making architectural changes.

In this repo, QMD should be part of the normal document query flow:

1. Query QMD first to find the relevant requirements or design documents.
2. Open the returned files and read the exact sections you need.
3. Treat `docs/ears/` as the product source of truth and `docs/design/` as the architecture source of truth.

## What To Query

- `docs/ears/`: product requirements, user stories, acceptance criteria
- `docs/design/`: constitution and architecture guidance
- Other Markdown files under `docs/`: supporting implementation notes

## Preferred Agent Workflow

When an agent needs documentation context, use the QMD MCP tools instead of broad file greps:

1. `mcp__qmd__status`
   Confirm the index is healthy and the `docs` collection is available.
2. `mcp__qmd__query`
   Search by keyword, semantic intent, or a hybrid query to identify the right documents.
3. `mcp__qmd__get`
   Open a specific result by file path or `docid`.
4. `mcp__qmd__multi_get`
   Pull a small set of related Markdown files when one result is not enough.

This keeps document discovery fast and reduces accidental reliance on stale assumptions.

## Query Patterns

Use the query type that matches the question:

- `lex`: exact names, terms, or phrases
- `vec`: natural-language questions
- `hyde`: hypothetical answer text for nuanced retrieval
- Hybrid: combine `lex` and `vec` for the best default recall

Recommended default for feature work:

```json
[
  { "type": "lex", "query": "\"provider configuration\" routing" },
  { "type": "vec", "query": "where does the repo define provider management requirements and constraints" }
]
```

Recommended default for architecture work:

```json
[
  { "type": "lex", "query": "\"multi-tenant\" isolation \"privacy-first\"" },
  { "type": "vec", "query": "which design documents define the architectural constraints for this change" }
]
```

## MCP Examples

Check index status:

```text
mcp__qmd__status()
```

Find the requirements for role permissions:

```json
mcp__qmd__query({
  "searches": [
    { "type": "lex", "query": "\"roles\" permissions RBAC" },
    { "type": "vec", "query": "where are user role and permission requirements documented" }
  ],
  "limit": 5
})
```

Open a result after query:

```json
mcp__qmd__get({
  "file": "docs/ears/01_identity_and_access/02_roles_permissions/README.md",
  "lineNumbers": true,
  "maxLines": 120
})
```

Open by `docid` when query already returned one:

```json
mcp__qmd__get({
  "file": "#abc123",
  "lineNumbers": true
})
```

Pull a small document set for a broader topic:

```json
mcp__qmd__multi_get({
  "pattern": "docs/ears/02_provider_management/**/*.md,docs/design/architecture.md",
  "lineNumbers": true,
  "maxLines": 160
})
```

## Local Setup

If the local QMD index needs to be rebuilt, use the setup script in this directory:

```bash
cd docs
./qmd-setup.sh
```

The script exists to prepare the local `docs` collection and generate embeddings for semantic search.

If you need to install QMD manually:

```bash
npm install -g @tobilu/qmd
```

The repository also includes a workaround in [`docs/qmd-setup.sh`](/Users/jian/workspaces/github/code-tegether.opensource/docs/qmd-setup.sh) for the current `BUN_INSTALL` launcher issue.

## Model Downloads

If your environment cannot fetch models automatically, download them manually:

```bash
hf download ggml-org/embeddinggemma-300M-GGUF embeddinggemma-300M-Q8_0.gguf --local-dir $HOME/.cache/qmd/models
hf download ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF qwen3-reranker-0.6b-q8_0.gguf --local-dir $HOME/.cache/qmd/models
hf download tobil/qmd-query-expansion-1.7B-gguf qmd-query-expansion-1.7B-q4_k_m.gguf --local-dir $HOME/.cache/qmd/models

cd $HOME/.cache/qmd/models
ln -s qmd-query-expansion-1.7B-q4_k_m.gguf hf_ggml-org_qmd-query-expansion-1.7B-q4_k_m.gguf
ln -s embeddinggemma-300M-Q8_0.gguf hf_ggml-org_embeddinggemma-300M-Q8_0.gguf
```

## References

- QMD repository: <https://github.com/tobi/qmd>
- DeepWiki overview: <https://deepwiki.com/tobi/qmd>
- Embedding model: <https://huggingface.co/ggml-org/embeddinggemma-300M-GGUF>
- Reranker model: <https://huggingface.co/ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF>
- Query expansion model: <https://huggingface.co/tobil/qmd-query-expansion-1.7B-gguf>
