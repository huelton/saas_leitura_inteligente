# Deploy (API)

## Variáveis obrigatórias em produção

- `DATABASE_URL` — PostgreSQL gerenciado (SSL recomendado).
- `JWT_SECRET` — segredo forte e único (nunca versionar).
- `GEMINI_API_KEY` — chave da API de IA (Gemini).
- `S3_BUCKET` — bucket para persistir PDFs.
- `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` — credenciais S3.
- `CORS_ALLOWED_ORIGINS` — domínio(s) do frontend (produção).
- Opcional: `METRICS_ENABLED=false` no edge público ou `METRICS_BASIC_*` (ver `docs/SECURITY.md`).

## Opções comuns

1. **VM + Docker:** build da imagem com `Dockerfile` na raiz, publicar a porta `8080`, Postgres separado.
2. **PaaS (Fly.io, Render, Railway):** definir as variáveis no painel, expor HTTP, escalar workers se necessário.
3. **Observabilidade:** logs em stdout; métricas em `GET /metrics` (proteger ou desabilitar no público — `docs/SECURITY.md`).
4. **Storage PDF (AWS):** `STORAGE_PROVIDER=s3`, `S3_ENDPOINT` vazio (AWS oficial), `S3_USE_PATH_STYLE=false`.
5. **Performance upload:** ajustar `S3_UPLOAD_PART_SIZE_MB` e `S3_UPLOAD_CONCURRENCY` conforme tráfego e tamanho médio dos PDFs.

## Frontend Next.js

Build em `web/` (`npm run build`), definir `NEXT_PUBLIC_API_URL` com a URL pública da API.
