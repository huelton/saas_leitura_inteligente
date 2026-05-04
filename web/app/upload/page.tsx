"use client";

import React from "react";
import { useEffect, useState } from "react";
import { api, getToken } from "@/lib/api";
import Link from "next/link";

export default function UploadPage() {
  const [mounted, setMounted] = useState(false);
  const [status, setStatus] = useState<string>("");
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return <section className="page"><div className="card"><p>Carregando...</p></div></section>;
  }

  if (!getToken()) {
    return (
      <section className="page">
        <div className="card">
        <p><Link href="/login">Faça login</Link> para enviar PDF.</p>
        </div>
      </section>
    );
  }

  async function onChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setErr(null);
    setStatus("Enviando…");
    const fd = new FormData();
    fd.append("file", file);
    fd.append("title", file.name);
    const res = await api("/books/upload", { method: "POST", body: fd });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      setErr((data as { error?: string }).error || res.statusText);
      setStatus("");
      return;
    }
    const jobId = (data as { job_id?: string }).job_id;
    setStatus(`Job ${jobId} aceito. Acompanhe em GET /jobs/${jobId}.`);
  }

  return (
    <section className="page">
      <div className="card" style={{ maxWidth: 640 }}>
        <h1 className="title">Enviar PDF</h1>
        <p className="muted">Seu arquivo será processado em segundo plano e aparecerá no dashboard.</p>
        <input className="input" type="file" accept="application/pdf" onChange={onChange} />
        {status && <p className="ok">{status}</p>}
        {err && <p className="error">{err}</p>}
        <p><Link href="/dashboard">Voltar ao dashboard</Link></p>
      </div>
    </section>
  );
}
