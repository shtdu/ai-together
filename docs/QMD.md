
# QMD - Query Markup Documents


The intention of this document to setup `qmd` [repo](https://github.com/tobi/qmd) locally to support query product or design document for requirement building.

## DeepWiki

It's great, Devin built a [DeepWiki](https://deepwiki.com/tobi/qmd) for this great tool.

## Claude Plugin

Install Claude Plugin, so we might build SKILLS for it.

```
claude plugin marketplace add tobi/qmd
claude plugin install qmd@qmd
```

## Manul Download Model

Due to some network limitation

```
hf download ggml-org/embeddinggemma-300M-GGUF embeddinggemma-300M-Q8_0.gguf --local-dir $HOME/.cache/qmd/models
hf download ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF qwen3-reranker-0.6b-q8_0.gguf --local-dir $HOME/.cache/qmd/models
hf download tobil/qmd-query-expansion-1.7B-gguf qmd-query-expansion-1.7B-q4_k_m.gguf --local-dir $HOME/.cache/qmd/models

cd $HOME/.cache/qmd/models
ln -s qmd-query-expansion-1.7B-q4_k_m.gguf hf_ggml-org_qmd-query-expansion-1.7B-q4_k_m.gguf
ln -s embeddinggemma-300M-Q8_0.gguf hf_ggml-org_embeddinggemma-300M-Q8_0.gguf

```

## Learning Purpose

I was inspired to make more study with the AI and LLM, so I keep the links of this tool used.

https://huggingface.co/ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF
https://huggingface.co/ggml-org/embeddinggemma-300M-GGUF
https://huggingface.co/tobil/qmd-query-expansion-1.7B-gguf

