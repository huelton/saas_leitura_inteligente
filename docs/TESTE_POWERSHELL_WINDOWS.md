# Teste rápido da API no PowerShell (Windows)

Guia de validação ponta a ponta da aplicação via PowerShell.

## Pré-requisitos

- API rodando em `http://localhost:8080`
- PostgreSQL ativo
- `.env` configurado com `DATABASE_URL`, `JWT_SECRET`, `GEMINI_API_KEY`, `GEMINI_MODEL`

## Passo a passo

```powershell
$api = "http://localhost:8080"
$email = "teste1@local.com"
$pass  = "123456"
$pdfPath = "C:\Users\Huelton\Downloads\seu-arquivo.pdf" # ajuste

# health
Invoke-RestMethod -Uri "$api/health" -Method GET

# register (opcional)
$registerBody = @{
  name     = "Teste User"
  email    = $email
  password = $pass
} | ConvertTo-Json -Compress
try {
  Invoke-RestMethod -Uri "$api/auth/register" -Method POST -ContentType "application/json" -Body $registerBody
} catch {
  Write-Host "Usuário já existe, seguindo..."
}

# login
$loginBody = @{ email = $email; password = $pass } | ConvertTo-Json -Compress
$login = Invoke-RestMethod -Uri "$api/auth/login" -Method POST -ContentType "application/json" -Body $loginBody
$token = $login.access_token.Trim()
$headers = @{ Authorization = "Bearer $token" }

# upload (multipart): usar curl.exe no Windows PowerShell 5.1
$uploadRaw = curl.exe -s -X POST "$api/books/upload" -H "Authorization: Bearer $token" -F "file=@$pdfPath" -F "title=PDF Teste"
$upload = $uploadRaw | ConvertFrom-Json
$jobId = $upload.job_id

# poll do job
$bookId = $null
for ($i=1; $i -le 30; $i++) {
  $job = Invoke-RestMethod -Uri "$api/jobs/$jobId" -Method GET -Headers $headers
  if ($job.status -eq "done") { $bookId = $job.book_id; break }
  if ($job.status -eq "failed") { throw "Job falhou: $($job.error)" }
  Start-Sleep -Seconds 2
}
if (-not $bookId) { throw "Timeout aguardando job." }

# leitura + progresso
$page1 = Invoke-RestMethod -Uri "$api/books/$bookId/pages/1" -Method GET -Headers $headers
$progressBody = @{ book_id = $bookId; page = 1; seconds = 45 } | ConvertTo-Json -Compress
Invoke-RestMethod -Uri "$api/reading/progress" -Method POST -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } -Body $progressBody

# perguntas + resposta
$qBody = @{ book_id = $bookId; page = 1 } | ConvertTo-Json -Compress
$qResp = Invoke-RestMethod -Uri "$api/ai/questions/page" -Method POST -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } -Body $qBody
$qid = $qResp.questions[0].id
$aBody = @{ question_id = $qid; answer = "Minha resposta de teste" } | ConvertTo-Json -Compress
Invoke-RestMethod -Uri "$api/answers" -Method POST -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } -Body $aBody

# dashboard
Invoke-RestMethod -Uri "$api/dashboard/$bookId" -Method GET -Headers $headers

# resumo + flashcards
$summaryBody = @{ book_id = $bookId; chapter_number = 1; page_from = 1; page_to = 2 } | ConvertTo-Json -Compress
Invoke-RestMethod -Uri "$api/ai/summary/chapter" -Method POST -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } -Body $summaryBody

$fBody = @{ book_id = $bookId; page = 1; max_cards = 3 } | ConvertTo-Json -Compress
$fResp = Invoke-RestMethod -Uri "$api/ai/flashcards/page" -Method POST -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } -Body $fBody
$flashId = $fResp.flashcards[0].id

$reviewBody = @{ flashcard_id = $flashId; correct = $true } | ConvertTo-Json -Compress
Invoke-RestMethod -Uri "$api/flashcards/review" -Method POST -Headers @{ Authorization = "Bearer $token"; "Content-Type" = "application/json" } -Body $reviewBody
Invoke-RestMethod -Uri "$api/flashcards/due?limit=10" -Method GET -Headers $headers

# métricas
Invoke-WebRequest -Uri "$api/metrics" -Method GET | Select-Object -ExpandProperty Content
```

## Observação

Se ocorrer erro de autenticação (`token inválido ou expirado`), gere novo token com login novamente e confirme se o `JWT_SECRET` da API é o mesmo do `.env`.
