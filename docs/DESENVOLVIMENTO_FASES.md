# Plano de desenvolvimento em fases (alinhado ao descritivo ChatGPT)

Este arquivo liga o **descritivo de produto** ao **código** e ao **`commits.json`** (fluxo de commits para push remoto agendado).

## O que foi consolidado do descritivo

1. **Proposta:** leitura acompanhada + perguntas progressivas (literal → crítica) + medição de compreensão.
2. **Sinais de leitura (evolução):** tempo por página, scroll, destaques; depois voz (STT).
3. **Stack sugerida:** Go + PostgreSQL + LLM; frontend Next.js depois.
4. **MVP:** upload PDF → páginas lógicas → perguntas IA → respostas avaliadas → score → histórico/dashboard básico.
5. **Próximas ondas:** resumos, flashcards, spaced repetition, gamificação, ranking, mobile.

## Fases implementadas neste repositório (backend)

| Fase | Conteúdo | Commit (mensagem resumida) |
|------|-----------|----------------------------|
| 1 | Base Go, `.gitignore`, dependências | `chore: gitignore e módulo Go` |
| 2 | Documentação de produto | `docs: especificação do produto` |
| 3 | Postgres via Docker | `infra: docker-compose PostgreSQL` |
| 4 | Schema + migrate na subida | `feat(db): schema SQL e conexão` |
| 5 | Config (`DATABASE_URL`, modelo OpenAI, chunk) | `feat(config): carregar env` |
| 6 | Cliente IA (perguntas + avaliação) | `feat(ai): cliente OpenAI` |
| 7 | PDF + split em páginas | `feat(reading): extração de PDF` |
| 8 | Repositórios | `feat(repo): usuários, livros` |
| 9 | Handlers HTTP | `feat(http): auth, upload` |
| 10 | `main` Gin | `feat(server): main Gin` |
| 11 | README, `.env.example`, scheduler | `docs: README, env exemplo` |
| 12 | Schema fase 2 + migrate idempotente | `feat(db): phase2` |
| 13 | IA resumo + flashcards JSON | `feat(ai): summary_flashcards` |
| 14 | Repositórios summary/flashcard + concat páginas | `feat(repo): phase2` |
| 15 | Rotas `/ai/summary/chapter`, flashcards, review | `feat(http): phase2` |
| 16 | README e commits.json | `docs: fase 2` |
| 17 | JWT, jobs PDF, limites free, métricas, logs | `feat: fase 3 auth async` |
| 18 | Frontend Next.js em `web/` | `feat: web nextjs` |
| 19 | Migração da IA para Gemini (config/env/cliente) | `feat(ai): migrate to gemini` |
| 20 | Guia de teste PowerShell no `docs/` | `docs: adicionar guia de teste powershell` |
| 21 | Checklist de go-live (backend/frontend/infra) | `docs: adicionar go-live checklist` |
| 22 | Cache MCQ por página (30), deduplicação semântica e prewarm assíncrono | `feat(ai): cache mcq com dedupe e prewarm` |
| 23 | PDFs em S3 (MinIO local/AWS), upload multipart e worker lendo do objeto | `feat(storage): s3 para pdf com upload otimizado` |
| 24 | CORS explícito, rate limit auth/upload, métricas opcionais/basic auth, teto e validação de PDF | `feat(security): hardening api` |
| 25 | Testes críticos: Go (testify, integração opt-in), Vitest e Playwright no `web/` | `test: cobertura critica backend e frontend` |
| 26 | CI completo no GitHub Actions (backend + integração opcional + frontend + e2e opcional) | `ci: pipeline de testes automatizada` |
| 27 | Testes unitários focados em `cmd/server`, `internal/repository` e `internal/database` | `test: unitarios server repository database` |

## Uso do agendador de commits

1. Inicialize o repositório Git e faça o **primeiro push** manual do remoto vazio (se aplicável).
2. Ajuste `.strategy_state.json`: `last_pushed_index`: **-1** para começar do primeiro item de `commits.json` (ou o índice do último commit já enviado).
3. Execute `python commit_scheduler.py` na raiz do projeto para um commit+push, ou `--schedule` para o loop com intervalo.

Ordem dos arquivos em cada entrada de `commits.json` segue dependências de implementação (documentação → infra → DB → serviços → HTTP).

**Histórico já completo no Git local:** o fluxo acima serve para **reproduzir pushes incrementais** em outro remoto ou para um clone onde você ainda vai criar os arquivos em lotes. Se o repositório já contém todos os commits locais, basta `git push origin main` uma vez (não é necessário rodar o agendador).
