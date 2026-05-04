"use client";

import React from "react";
import { useEffect, useState } from "react";
import { api, getToken } from "@/lib/api";
import Link from "next/link";

type Book = {
  id: string;
  title: string;
  total_pages: number;
};

export default function DashboardPage() {
  const [books, setBooks] = useState<Book[]>([]);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      setErr("Faça login primeiro.");
      return;
    }
    (async () => {
      const res = await api("/books");
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setErr((data as { error?: string }).error || res.statusText);
        return;
      }
      setBooks((data as { books?: Book[] }).books || []);
    })();
  }, []);

  return (
    <section className="page">
      <div className="card">
        <h1 className="title">Meus livros</h1>
        <p className="muted">Escolha um livro para continuar a leitura e responder perguntas por página.</p>
        <div className="nav">
          <Link className="btn" href="/upload">Enviar PDF</Link>
        </div>
      </div>
      {!getToken() && (
        <div className="card">
          <p><Link href="/login">Faça login para acessar seus livros.</Link></p>
        </div>
      )}
      {err && <p className="error">{err}</p>}
      <div className="grid grid-2">
        {books.map((b) => (
          <article key={b.id} className="card">
            <h3 className="title">{b.title}</h3>
            <p className="muted">{b.total_pages} páginas</p>
            <div className="nav">
              <Link className="btn btn-secondary" href={`/books/${b.id}/read`}>Ler</Link>
              <Link className="btn" href={`/books/${b.id}/questions`}>Perguntas</Link>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
