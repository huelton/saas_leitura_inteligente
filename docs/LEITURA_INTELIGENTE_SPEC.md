# Leitura Inteligente — Especificação do produto (consolidado)

Documento derivado do descritivo de produto e da arquitetura discutida: leitura acompanhada por **perguntas progressivas**, avaliação por IA e métricas de compreensão.

## 1. Visão e proposta de valor

- O usuário lê textos (PDF, livro, artigo, Bíblia, etc.).
- O sistema **acompanha** sinais de leitura (tempo, página, scroll, destaques; em evolução: voz).
- A IA **gera perguntas** sobre o trecho lido, com **níveis crescentes de profundidade** conforme o desempenho.
- O app **mede compreensão** e oferece revisão, resumos e flashcards.

Analogia de mercado: combinação de **Kindle + ChatGPT + Duolingo** (leitura + IA + gamificação educacional).

## 2. Níveis de perguntas (taxonomia)

| Nível | Nome | Foco |
|------|------|------|
| 1 | Literal | Resposta explícita no texto |
| 2 | Interpretação | Inferência e causa/efeito |
| 3 | Explicação | Parafrasear e explicar intenção do autor |
| 4 | Aplicação | Transferência para situações reais |
| 5 | Crítica | Julgamento fundamentado |

## 3. Como “observar” a leitura (roadmap de sinais)

1. **Tempo por página / seção** — simples; indica dificuldade ou engajamento.
2. **Scroll e foco** (web/mobile) — parágrafos com dwell time, destaques, cliques.
3. **Leitura em voz alta** — STT, fluência, omissões de palavras (forte para idiomas e crianças).

## 4. Fluxo técnico principal (core)

1. Usuário abre ou faz upload do texto.
2. Sistema segmenta (páginas lógicas ou capítulos).
3. Durante/ após leitura, **geração de perguntas** sobre o trecho.
4. Usuário responde → **avaliação pela IA** (nota + feedback).
5. Ajuste de **dificuldade** e agregação de **score de compreensão**.
6. Extensões: resumo, flashcards, **revisão espaçada** (ex.: 3 dias).

## 5. Arquitetura alvo (SaaS)

- **Cliente:** Web (Next.js/React) → depois mobile (React Native).
- **API:** Monólito modular ou microsserviços: Auth, Reading, AI Questions, Progress/Analytics.
- **Dados:** PostgreSQL (estado, conteúdo indexado, respostas), Redis (sessão/cache), armazenamento de objetos (PDFs).
- **IA:** API de LLM (ex.: OpenAI) para perguntas, correção, resumos e flashcards.
- **Infra:** JWT, filas opcionais para jobs pesados, cloud (ex.: AWS + S3).

## 6. Modelo de dados (resumo)

Entidades principais: `users`, `books`, `book_pages` (texto por página lógica), `reading_sessions`, `highlights`, `questions`, `answers`, `comprehension_scores` / `reading_scores`, `flashcards`, `chapter_summaries` (futuro).

## 7. APIs (MVP e evolução)

**MVP:** upload de PDF, leitura por página, gerar perguntas, responder, avaliar, score, histórico básico.

**Evolução:** resumo por capítulo, flashcards, `/flashcards/review`, gamificação, ranking, dashboard rico, modos (Bíblia, inglês, concursos).

## 8. Prompts (princípios)

- **Gerar perguntas:** instruir níveis (1 literal … 5 crítica), saída estruturada (JSON).
- **Avaliar:** pergunta + resposta do aluno → nota 0–10 + feedback curto (JSON).
- **Resumo / flashcards:** formato fixo para parsing e persistência.

## 9. Roadmap por fases

| Fase | Conteúdo |
|------|-----------|
| **MVP (~30 dias)** | Auth básica, upload PDF, chunking em páginas, perguntas IA, respostas, nota, score, histórico |
| **2** | Resumos, flashcards, spaced repetition, highlights, metas, dashboard |
| **3** | Voz, ranking, gamificação, mobile, exportação, nichos (Bíblia, inglês) |

## 10. Monetização (ideias)

- Freemium: limite de livros/perguntas/dia; planos pagos com ilimitado, flashcards, revisão, exportação, métricas avançadas.
- Nichos: estudantes, concursos, idiomas, espiritual, técnicos.

## 11. Diferencial de fechamento de sessão

Exibir: **% de entendimento**, pontos fortes/fracos, resumo do capítulo, erros, **revisão agendada** (spaced repetition).

---

Este arquivo é a **fonte única de verdade** para alinhar backend, frontend e `commits.json` do repositório.
