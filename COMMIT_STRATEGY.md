# Estratégia de commits — Leitura Inteligente

Commits em lotes lógicos (~2–3 h de trabalho cada), para uso com `commit_scheduler.py` e o arquivo **`commits.json`** na raiz.

## Objetivo

- Ordem cronológica de desenvolvimento refletida na lista.
- Intervalo entre pushes (modo agendado): base **1,5 h** + jitter **10–30 min** (ver script).
- Horário comercial local **08:00–19:00**, seg–sáb (configurável no script).
- Sem co-autor nos commits.

## Uso

- **Manual:** `python commit_scheduler.py` — faz commit e push do **próximo** item da lista.
- **Agendado:** `python commit_scheduler.py --schedule` — repete até esgotar a lista.
- **Estado:** `.strategy_state.json` guarda o último índice enviado (não versionar segredos).

## Regras

1. Um commit da lista por execução (no modo manual).
2. Cada entrada tem `message` e `files` (caminhos relativos ao repositório).
3. Novas fases: **append** ao final de `commits.json`.
4. Todo arquivo **novo** ou **deletado** deve aparecer em ao menos uma entrada do `commits.json`.
5. O agendador deve comitar **somente** os caminhos da entrada atual (pathspec), sem puxar outros arquivos staged.

## Fases resumidas (Leitura Inteligente)

1. Base: `.gitignore`, `go.mod`, README.
2. Documentação: spec do produto.
3. Infra: `docker-compose` (PostgreSQL).
4. Config e conexão com banco + migrações SQL.
5. Servidor HTTP (Gin) e health check.
6. Cliente de IA (Gemini, com fallback legado OpenAI).
7. Upload de PDF, extração, páginas lógicas e persistência.
8. Perguntas por página e respostas com score.
9. Dashboard de leitura e documentação de API.
10. Cache inteligente de perguntas (dedupe + prewarm assíncrono).
11. Storage de PDFs em S3 (MinIO local / AWS remoto) com upload multipart otimizado.
12. Hardening: CORS por origem, rate limit em auth/upload, proteção de `/metrics`, limite/validação de upload PDF.
13. Testes: Go (`testify`, testes de integração com tag `integration`), frontend Vitest + Playwright — ver `docs/TESTING.md`.
14. CI: GitHub Actions com backend, integração condicional por `TEST_DATABASE_URL` e frontend (Vitest + Playwright opcional em main/master).
15. Testes unitários adicionais para `cmd/server`, `internal/repository` e `internal/database`.

Sincronizar com histórico existente: ajustar `last_pushed_index` em `.strategy_state.json` para o índice do último commit já enviado (base 0).
