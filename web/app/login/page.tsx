"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api, setToken } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErr(null);
    const res = await api("/auth/login", { method: "POST", json: { email, password } });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      setErr((data as { error?: string }).error || res.statusText);
      return;
    }
    const token = (data as { access_token?: string }).access_token;
    if (!token) {
      setErr("Resposta sem access_token");
      return;
    }
    setToken(token);
    router.push("/dashboard");
  }

  return (
    <section className="page">
      <div className="card" style={{ maxWidth: 460 }}>
        <h1 className="title">Entrar</h1>
        <p className="muted">Acesse sua conta para continuar os estudos.</p>
        <form onSubmit={onSubmit} className="grid">
          <div className="field">
            <label>Email</label>
          <input
            className="input"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          </div>
          <div className="field">
            <label>Senha</label>
          <input
            className="input"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          </div>
          <button className="btn" type="submit">Entrar</button>
        </form>
        {err && <p className="error">{err}</p>}
      </div>
    </section>
  );
}
