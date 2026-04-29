# Resumo executivo — Leitura Inteligente (SaaS)

## Objetivo

Produto **SaaS educacional** que combina leitura digital com **perguntas geradas por IA** em níveis progressivos (literal → crítica), **avaliação automática das respostas** e **métricas de compreensão**, evoluindo para resumos, flashcards e revisão espaçada.

## Problema central

Medir não só “quanto” a pessoa leu, mas **o quanto entendeu**, usando sinais de leitura (tempo, releitura, scroll, destaques; futuro: voz) e respostas discursivas avaliadas por modelo de linguagem.

## Stack sugerida (referência)

- Backend: Go (Gin) ou Java Spring Boot.
- Frontend: Next.js/React; mobile depois (React Native).
- Dados: PostgreSQL, Redis, storage para PDFs.
- IA: API compatível com chat completions (ex.: OpenAI).

## Estado deste repositório

Backend **MVP em Go** na raiz (`cmd/`, `internal/`, `migrations/`), com Docker apenas para PostgreSQL. Detalhes de domínio e roadmap: `docs/LEITURA_INTELIGENTE_SPEC.md`.

## Próximos passos de produto

1. Consolidar MVP (upload, páginas, perguntas, avaliação, score).
2. Resumos e flashcards com revisão espaçada.
3. Gamificação, modos por nicho e app mobile.
