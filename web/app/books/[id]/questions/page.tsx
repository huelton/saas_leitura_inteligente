"use client";

import Link from "next/link";
import { useParams, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";
import { api, getToken } from "@/lib/api";

type MCQ = {
  id: string;
  question: string;
  difficulty: string;
  options: string[];
  correct_idx: number;
};

export default function QuestionsPage() {
  const params = useParams();
  const search = useSearchParams();
  const bookId = params.id as string;
  const initialPage = Number(search.get("page") || "1");

  const [page, setPage] = useState(Number.isFinite(initialPage) && initialPage > 0 ? initialPage : 1);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [items, setItems] = useState<MCQ[]>([]);
  const [picked, setPicked] = useState<Record<string, number>>({});
  const [feedback, setFeedback] = useState<Record<string, string>>({});

  const allAnswered = useMemo(
    () => items.length > 0 && items.every((q) => picked[q.id] !== undefined),
    [items, picked]
  );

  async function generate() {
    if (!getToken()) {
      setErr("Faça login para gerar perguntas.");
      return;
    }
    setLoading(true);
    setErr(null);
    setFeedback({});
    const res = await api("/ai/questions/page-mcq", {
      method: "POST",
      json: { book_id: bookId, page },
    });
    const data = await res.json().catch(() => ({}));
    setLoading(false);
    if (!res.ok) {
      setErr((data as { error?: string }).error || res.statusText);
      return;
    }
    setItems((data as { questions?: MCQ[] }).questions || []);
    setPicked({});
  }

  async function submitOne(q: MCQ) {
    const idx = picked[q.id];
    if (idx === undefined) return;
    const answer = q.options[idx] || "";
    const res = await api("/answers", {
      method: "POST",
      json: { question_id: q.id, answer },
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      setFeedback((prev) => ({ ...prev, [q.id]: (data as { error?: string }).error || res.statusText }));
      return;
    }
    const d = data as { score?: number; feedback?: string };
    setFeedback((prev) => ({
      ...prev,
      [q.id]: `Nota: ${d.score ?? "-"} • ${d.feedback ?? "Resposta enviada"}`,
    }));
  }

  async function submitAll() {
    for (const q of items) {
      if (picked[q.id] !== undefined) {
        // eslint-disable-next-line no-await-in-loop
        await submitOne(q);
      }
    }
  }

  return (
    <section className="page">
      <div className="card">
        <p><Link href={`/books/${bookId}/read`}>← Voltar para leitura</Link></p>
        <h1 className="title">Perguntas por página</h1>
        <p className="muted">Gere questões de múltipla escolha com base no conteúdo da página atual.</p>
        <div className="nav">
          <button className="btn btn-secondary" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1 || loading}>
            Página anterior
          </button>
          <button className="btn btn-secondary" onClick={() => setPage((p) => p + 1)} disabled={loading}>
            Próxima página
          </button>
          <button className="btn" onClick={generate} disabled={loading}>
            {loading ? "Gerando..." : `Gerar perguntas da página ${page}`}
          </button>
        </div>
      </div>

      {err && <p className="error">{err}</p>}

      {items.map((q) => (
        <article className="card question-card" key={q.id}>
          <h3 className="title">{q.question}</h3>
          <p className="muted">Nível: {q.difficulty}</p>
          {q.options.map((opt, idx) => (
            <label className="choice" key={`${q.id}-${idx}`}>
              <input
                type="radio"
                name={`q-${q.id}`}
                checked={picked[q.id] === idx}
                onChange={() => setPicked((prev) => ({ ...prev, [q.id]: idx }))}
              />
              <span>{opt}</span>
            </label>
          ))}
          <div className="nav" style={{ marginTop: "0.8rem" }}>
            <button
              className="btn"
              disabled={picked[q.id] === undefined}
              onClick={() => submitOne(q)}
            >
              Enviar resposta
            </button>
            {feedback[q.id] && <span className="muted">{feedback[q.id]}</span>}
          </div>
        </article>
      ))}

      {items.length > 0 && (
        <div className="card">
          <button className="btn" disabled={!allAnswered} onClick={submitAll}>
            Enviar todas
          </button>
        </div>
      )}
    </section>
  );
}
