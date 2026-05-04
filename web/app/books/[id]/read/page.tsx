"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api, getToken } from "@/lib/api";
import Link from "next/link";

export default function ReadPage() {
  const params = useParams();
  const bookId = params.id as string;
  const [page, setPage] = useState(1);
  const [content, setContent] = useState("");
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    const start = Date.now();
    return () => {
      if (!getToken()) return;
      const sec = Math.max(1, Math.round((Date.now() - start) / 1000));
      void api("/reading/progress", {
        method: "POST",
        json: { book_id: bookId, page, seconds: sec },
      });
    };
  }, [bookId, page]);

  useEffect(() => {
    if (!getToken()) {
      setErr("Login necessário");
      return;
    }
    setErr(null);
    (async () => {
      const res = await api(`/books/${bookId}/pages/${page}`);
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setErr((data as { error?: string }).error || res.statusText);
        return;
      }
      setContent((data as { content?: string }).content || "");
    })();
  }, [bookId, page]);

  return (
    <section className="page">
      <div className="card">
        <p><Link href="/dashboard">← Dashboard</Link></p>
        <h1 className="title">Leitura • Página {page}</h1>
        <p className="muted">Após ler, avance ou abra perguntas para treinar compreensão.</p>
        <div className="nav">
          <Link className="btn" href={`/books/${bookId}/questions?page=${page}`}>
            Ir para perguntas da página
          </Link>
        </div>
      </div>
      {err && <p className="error">{err}</p>}
      <div className="card">
        <div className="reader">{content}</div>
      </div>
      <div className="nav">
        <button className="btn btn-secondary" type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
          Anterior
        </button>
        <button className="btn" type="button" onClick={() => setPage((p) => p + 1)}>
          Próxima
        </button>
      </div>
    </section>
  );
}
