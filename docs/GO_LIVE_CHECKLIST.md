# Go-Live Checklist — Base de conhecimento (Próxima Fase)

Checklist oficial para ajuste fino antes de liberar o projeto em produção.
Use este documento como referência de validação final para **backend**, **frontend** e **infra/operação**.

## Como usar

- Marque cada item com: `[ ]` pendente, `[x]` concluído, `[-]` não aplicável.
- Sempre anexar evidência curta (link, print, log, comando executado).
- Não iniciar publicação automática (`commit_scheduler`) sem checklist mínimo concluído.

---

## 1) Backend — Pronto para produção

### 1.1 API e regras de negócio
- [ ] Endpoints críticos testados: auth, upload, jobs, leitura, perguntas, respostas, resumo, flashcards.
- [ ] Fluxo de erros validado (400/401/403/404/429/500) com mensagens consistentes.
- [ ] Regras freemium validadas (`FREE_MAX_BOOKS`, `FREE_MAX_QUESTIONS_PER_DAY`).
- [ ] Verificação de ownership de livro/recursos funcionando em todos os endpoints protegidos.

### 1.2 Segurança
- [ ] `JWT_SECRET` forte (>= 32 chars), diferente por ambiente.
- [ ] Secrets fora do repositório (nada sensível versionado).
- [ ] CORS: `CORS_ALLOWED_ORIGINS` com domínio(s) reais do frontend em produção (ver `docs/SECURITY.md`).
- [ ] Rate limit: `AUTH_RATE_LIMIT_PER_MIN` e `UPLOAD_RATE_LIMIT_PER_MIN` adequados ao tráfego.
- [ ] Upload: `MAX_UPLOAD_BYTES` alinhado ao proxy/PaaS; validação de PDF ativa (assinatura `%PDF-`).
- [ ] Métricas: `METRICS_ENABLED` / Basic auth (`METRICS_BASIC_*`) ou métricas só via rede interna.
- [ ] Conferir implementação detalhada em `docs/SECURITY.md`.

### 1.3 Qualidade técnica
- [ ] `go test ./...` sem falhas; opcional em CI: `go test -tags=integration ./internal/handler/` com `TEST_DATABASE_URL`.
- [ ] Frontend: `npm test` e `npm run test:e2e` (pasta `web/`) conforme `docs/TESTING.md`.
- [ ] Build da API sem warnings críticos.
- [ ] Health check (`/health`) público; métricas (`/metrics`) conforme política (desabilitado no edge ou protegido).
- [ ] Logs estruturados com campos mínimos: método, rota, status, latência.

### 1.4 Banco e dados
- [ ] Migrações aplicando limpas em ambiente novo.
- [ ] Migrações idempotentes (ambiente já existente).
- [ ] Índices essenciais revisados para consultas quentes.
- [ ] Política de backup/restauração definida e testada.

---

## 2) Frontend — Pronto para uso real

### 2.1 Fluxos essenciais
- [ ] Login/logout.
- [ ] Upload de PDF e acompanhamento de job.
- [ ] Leitura por página e envio de progresso.
- [ ] Geração de perguntas e resposta.
- [ ] Visualização de score/dashboard.
- [ ] Resumo e flashcards/revisão.

### 2.2 UX e robustez
- [ ] Estados de loading/sucesso/erro em telas principais.
- [ ] Mensagens de erro amigáveis para falhas de API.
- [ ] Token expirado tratado (redireciona para login).
- [ ] Navegação mínima sem páginas quebradas.

### 2.3 Build e ambiente
- [ ] `npm run build` sem erros.
- [ ] `NEXT_PUBLIC_API_URL` correto para ambiente de produção.
- [ ] Sem variáveis sensíveis expostas no bundle cliente.

---

## 3) Infra e Operação

### 3.1 Deploy
- [ ] `Dockerfile` funcionando em build limpo.
- [ ] Variáveis obrigatórias configuradas no ambiente de deploy.
- [ ] Banco gerenciado/externo acessível com SSL quando aplicável.
- [ ] Processo de rollback documentado (versão anterior + restore).

### 3.2 Observabilidade e incidentes
- [ ] Dashboard básico com métricas de erro, latência e throughput.
- [ ] Alertas mínimos (API fora, erro 5xx alto, banco indisponível).
- [ ] Logs centralizados com retenção definida.
- [ ] Runbook curto para incidentes comuns.

### 3.3 Performance inicial
- [ ] Teste de carga leve (smoke) para endpoints principais.
- [ ] Timeout/retry revisados para chamadas externas de IA.
- [ ] Worker assíncrono de upload sem fila bloqueada em uso básico.

---

## 4) Publicação e Governança

- [ ] `README`, `DEPLOY`, `DESENVOLVIMENTO_FASES` e este checklist atualizados.
- [ ] `commits.json` alinhado com os artefatos novos/deletados.
- [ ] Árvore git limpa (sem mudanças locais pendentes) antes do release.
- [ ] Tag/versionamento definido para o release.

---

## Critério de Go / No-Go

### Go (pode publicar)
- Backend: itens críticos concluídos.
- Frontend: fluxos essenciais funcionando.
- Infra: deploy e observabilidade mínimos ativos.

### No-Go (não publicar)
- Falha em auth, upload/job, perguntas/respostas, ou migração.
- Ausência de monitoramento básico.
- Secrets/versionamento inconsistentes.

---

## Próxima fase recomendada (execução)

1. Fechar pendências técnicas do backend.
2. Estabilizar frontend nos fluxos essenciais.
3. Validar deploy em ambiente staging.
4. Rodar checklist completo com evidências.
5. Liberar publish/scheduler somente após Go.
