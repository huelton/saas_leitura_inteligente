# Leitura Inteligente — backend (MVP)

[![CI](https://github.com/huelton/leitura-inteligente/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/huelton/leitura-inteligente/actions/workflows/ci.yml)
[![CI (master)](https://github.com/huelton/leitura-inteligente/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/huelton/leitura-inteligente/actions/workflows/ci.yml)
[![Coverage](https://github.com/huelton/leitura-inteligente/actions/workflows/coverage.yml/badge.svg?branch=main)](https://github.com/huelton/leitura-inteligente/actions/workflows/coverage.yml)

API em **Go (Gin)** para leitura de PDF com **perguntas geradas por IA** (níveis literal → crítica), avaliação de respostas e **score de compreensão** (0–100%).

Documentação de produto: `docs/LEITURA_INTELIGENTE_SPEC.md`.

## Pré-requisitos

- Go 1.22+
- Docker (PostgreSQL + MinIO local) ou infraestrutura remota equivalente
- Chave `GEMINI_API_KEY` para geração de perguntas e avaliação

## Subir o banco

```bash
docker compose up -d
```

Copie `.env.example` para `.env` e ajuste `DATABASE_URL`/S3 se necessário.

## Rodar a API

```bash
go run ./cmd/server
```

Health: `GET http://localhost:8080/health`

## Autenticação (fase 3)

Rotas abertas: `GET /health`, `POST /auth/register`, `POST /auth/login`.

Demais rotas exigem header **`Authorization: Bearer <access_token>`** (JWT HS256).

- `POST /auth/register` e `POST /auth/login` retornam `access_token`, `expires_in`, dados do usuário.
- Configure **`JWT_SECRET`** no servidor (obrigatório para login/registro).

## Fluxo rápido (API)

1. `POST /auth/register` ou `POST /auth/login` — guardar `access_token`.
2. `POST /books/upload` — `multipart/form-data` (`file`, opcional `title`, `author`) — resposta **202** com `job_id` (processamento assíncrono).
3. `GET /jobs/:id` — status do job; quando `status=done`, usar `book_id`.
4. `GET /books` — lista livros do usuário.
5. `GET /books/:id/pages/:page` — texto da página lógica.
6. `POST /ai/questions/page` — `{ "book_id", "page" }` — perguntas abertas (IA).
7. `POST /ai/questions/page-mcq` — `{ "book_id", "page", "count?" }` — cacheia até 30 questões por página no banco e depois serve aleatório sem nova chamada de IA.
8. `POST /answers` — `{ "question_id", "answer" }` (sem `user_id` no corpo). Para MCQ, corrige localmente sem IA.
9. `GET /dashboard/:bookId` — score e contagens.
10. `POST /reading/progress` — `{ "book_id", "page", "seconds" }` — tracking de tempo por página.

**Freemium:** `FREE_MAX_BOOKS` (padrão 2) e `FREE_MAX_QUESTIONS_PER_DAY` (padrão 10) para plano `free`. Planos diferentes de `free` não aplicam estes limites nesta versão.

**Observabilidade:** logs JSON em stdout; **`GET /metrics`** (Prometheus).

## Fase 2 — resumo por capítulo e flashcards

7. `POST /ai/summary/chapter` — JSON `{ "book_id", "chapter_number", "page_from", "page_to" }` — concatena páginas, gera resumo estruturado (IA) e persiste em `chapter_summaries`.
8. `GET /summaries?book_id=` — lista resumos salvos do livro.
9. `POST /ai/flashcards/page` — JSON `{ "book_id", "page", "max_cards" (opcional) }` — JWT define o usuário.
10. `GET /flashcards/due?limit=` — cartões vencidos do usuário do token.
11. `POST /flashcards/review` — JSON `{ "flashcard_id", "correct" }`.

Migrações: `migrations/002_phase2.sql`, `migrations/003_phase3.sql` (jobs + uso diário). Esquemas embutidos em `internal/database/`.

## Frontend (Next.js)

Pasta `web/`: `npm install` e `npm run dev`. Defina `NEXT_PUBLIC_API_URL=http://localhost:8080` (ou URL da API em produção). Inclui login, lista de livros, upload e leitor por página com envio de progresso ao trocar de página.

## Docker (API)

`docker build -t leitura-api .` — binário em Alpine (ver `Dockerfile`). Ver `docs/DEPLOY.md`.

## Publicar no GitHub (remoto)

Substitua a URL pelo seu repositório:

```bash
git remote add origin https://github.com/SEU_USUARIO/leitura-inteligente.git
git push -u origin main
```

Se o remoto já existir: `git remote set-url origin <url>`.

## Variáveis de ambiente

| Variável | Descrição |
|----------|-----------|
| `DATABASE_URL` | URL PostgreSQL (obrigatório) |
| `GEMINI_API_KEY` | Chave da API Gemini |
| `GEMINI_MODEL` | Padrão: `gemini-1.5-flash` |
| `APP_ENV` | `development` (padrão) ou `production` |
| `HTTP_ADDR` | Padrão: `:8080` |
| `PAGE_CHUNK_SIZE` | Tamanho de cada “página” lógica em caracteres (padrão 1000) |
| `JWT_SECRET` | Segredo HS256 (obrigatório para auth) |
| `JWT_EXPIRY_HOURS` | Validade do token (padrão 72) |
| `FREE_MAX_BOOKS` | Máx. livros no plano free (padrão 2) |
| `FREE_MAX_QUESTIONS_PER_DAY` | Máx. perguntas geradas/dia no free (padrão 10) |
| `STORAGE_PROVIDER` | `s3` (padrão) |
| `S3_BUCKET` | Bucket para PDFs |
| `S3_ENDPOINT` | Endpoint S3 compatível (ex.: `http://localhost:9000` no MinIO) |
| `S3_USE_PATH_STYLE` | `true` para MinIO/local |
| `AWS_ACCESS_KEY_ID` | Access key S3 |
| `AWS_SECRET_ACCESS_KEY` | Secret key S3 |
| `AWS_REGION` | Região (ex.: `us-east-1`) |
| `S3_UPLOAD_PART_SIZE_MB` | Tamanho da parte multipart upload (padrão 8) |
| `S3_UPLOAD_CONCURRENCY` | Paralelismo no upload multipart (padrão 4) |
| `CORS_ALLOWED_ORIGINS` | Origens permitidas (vírgula). Em `development` vazio → `http://localhost:3000` |
| `AUTH_RATE_LIMIT_PER_MIN` | Limite por IP em `/auth/*` (padrão 30) |
| `UPLOAD_RATE_LIMIT_PER_MIN` | Limite por IP em `POST /books/upload` (padrão 20) |
| `MAX_UPLOAD_BYTES` | Tamanho máximo do upload multipart (padrão 50 MiB) |
| `METRICS_ENABLED` | `true`/`false` — se `false`, rota `/metrics` não é exposta |
| `METRICS_BASIC_USER` / `METRICS_BASIC_PASSWORD` | Se ambos definidos, Basic auth em `/metrics` |

Detalhes de segurança: `docs/SECURITY.md`.

> Em `APP_ENV=development`, os limites freemium são desabilitados automaticamente para facilitar testes locais.
> Em `APP_ENV=production`, os limites configurados são aplicados.
> Upload de PDF em produção usa S3 com multipart upload para menor latência e melhor throughput.

## Testes (backend + frontend)

Resumo da estratégia, comandos e integração com Postgres: **`docs/TESTING.md`**.

- **Go:** `go test ./...` — middlewares, limites free, config, PDF; integração com `go test -tags=integration ./internal/handler/` e `TEST_DATABASE_URL`.
- **Web:** na pasta `web/`, `npm test` (Vitest) e `npm run test:e2e` (Playwright).

## Commits agendados (remoto)

Lista em `commits.json`; uso do script em `COMMIT_STRATEGY.md` (intervalo 2h + jitter em modo `--schedule`).

## Teste rápido no PowerShell (Windows)

Este fluxo valida ponta a ponta: auth, upload assíncrono, leitura, perguntas, respostas, resumo e flashcards.

### 1) Pré-requisitos

- API rodando em `http://localhost:8080`
- PostgreSQL ativo
- `.env` configurado (`DATABASE_URL`, `JWT_SECRET`, `GEMINI_API_KEY`, `GEMINI_MODEL`)

### 2) Definir variáveis

```powershell
$api = "http://localhost:8080"
$email = "teste1@local.com"
$pass  = "123456"
$pdfPath = "C:\Users\Huelton\Downloads\seu-arquivo.pdf" # ajuste aqui
```

### 3) Health

```powershell
Invoke-RestMethod -Uri "$api/health" -Method GET
```

### 4) Registro (opcional) + login

```powershell
$registerBody = @{
  name     = "Teste User"
  email    = $email
  password = $pass
} | ConvertTo-Json -Compress

try {
  Invoke-RestMethod `
    -Uri "$api/auth/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $registerBody
} catch {
  Write-Host "Usuário já existe, seguindo para login..."
}

$loginBody = @{
  email    = $email
  password = $pass
} | ConvertTo-Json -Compress

$login = Invoke-RestMethod `
  -Uri "$api/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body $loginBody

$token = $login.access_token.Trim()
$headers = @{ Authorization = "Bearer $token" }
```

### 5) Upload de PDF (multipart com curl.exe)

No Windows PowerShell 5.1, use `curl.exe` para `multipart/form-data`.

```powershell
$uploadRaw = curl.exe -s -X POST "$api/books/upload" `
  -H "Authorization: Bearer $token" `
  -F "file=@$pdfPath" `
  -F "title=PDF Teste"

$upload = $uploadRaw | ConvertFrom-Json
$jobId = $upload.job_id
$jobId
```

### 6) Acompanhar job até concluir

```powershell
$maxTries = 30
$bookId = $null

for ($i=1; $i -le $maxTries; $i++) {
  $job = Invoke-RestMethod -Uri "$api/jobs/$jobId" -Method GET -Headers $headers
  Write-Host "Tentativa $i - status: $($job.status)"

  if ($job.status -eq "done") {
    $bookId = $job.book_id
    break
  }

  if ($job.status -eq "failed") {
    throw "Job falhou: $($job.error)"
  }

  Start-Sleep -Seconds 2
}

if (-not $bookId) { throw "Timeout esperando job finalizar." }
$bookId
```

### 7) Ler página e registrar progresso

```powershell
$page1 = Invoke-RestMethod -Uri "$api/books/$bookId/pages/1" -Method GET -Headers $headers
$page1.content.Substring(0, [Math]::Min(300, $page1.content.Length))

$progressBody = @{
  book_id = $bookId
  page    = 1
  seconds = 45
} | ConvertTo-Json -Compress

Invoke-RestMethod `
  -Uri "$api/reading/progress" `
  -Method POST `
  -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } `
  -Body $progressBody
```

### 8) Gerar perguntas e responder

```powershell
$qBody = @{
  book_id = $bookId
  page    = 1
} | ConvertTo-Json -Compress

$qResp = Invoke-RestMethod `
  -Uri "$api/ai/questions/page" `
  -Method POST `
  -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } `
  -Body $qBody

$qid = $qResp.questions[0].id

$aBody = @{
  question_id = $qid
  answer      = "Minha resposta de teste"
} | ConvertTo-Json -Compress

Invoke-RestMethod `
  -Uri "$api/answers" `
  -Method POST `
  -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } `
  -Body $aBody
```

### 9) Dashboard

```powershell
Invoke-RestMethod -Uri "$api/dashboard/$bookId" -Method GET -Headers $headers
```

### 10) Resumo e flashcards

```powershell
$summaryBody = @{
  book_id        = $bookId
  chapter_number = 1
  page_from      = 1
  page_to        = 2
} | ConvertTo-Json -Compress

Invoke-RestMethod `
  -Uri "$api/ai/summary/chapter" `
  -Method POST `
  -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } `
  -Body $summaryBody

$fBody = @{
  book_id   = $bookId
  page      = 1
  max_cards = 3
} | ConvertTo-Json -Compress

$fResp = Invoke-RestMethod `
  -Uri "$api/ai/flashcards/page" `
  -Method POST `
  -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } `
  -Body $fBody

$flashId = $fResp.flashcards[0].id

$reviewBody = @{
  flashcard_id = $flashId
  correct      = $true
} | ConvertTo-Json -Compress

Invoke-RestMethod `
  -Uri "$api/flashcards/review" `
  -Method POST `
  -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } `
  -Body $reviewBody

Invoke-RestMethod -Uri "$api/flashcards/due?limit=10" -Method GET -Headers $headers
```

### 11) Métricas (observabilidade)

```powershell
Invoke-WebRequest -Uri "$api/metrics" -Method GET | Select-Object -ExpandProperty Content
```
