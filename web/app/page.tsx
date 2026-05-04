import Link from "next/link";

export default function Home() {
  return (
    <section className="page">
      <div className="card">
        <h1 className="title">Leitura inteligente com avaliação progressiva</h1>
        <p className="muted">
          Experiência no estilo mercado EdTech: leitura por páginas, perguntas em níveis de
          profundidade, score de compreensão e revisão por flashcards.
        </p>
        <div className="nav">
          <Link className="btn" href="/dashboard">
            Acessar dashboard
          </Link>
          <Link className="btn btn-secondary" href="/login">
            Fazer login
          </Link>
        </div>
      </div>
      <div className="grid grid-2">
        <div className="card">
          <h3 className="title">Leitura com contexto</h3>
          <p className="muted">Upload de PDF, paginação lógica e acompanhamento de progresso em tempo real.</p>
        </div>
        <div className="card">
          <h3 className="title">Perguntas por página</h3>
          <p className="muted">Geração automática e versão múltipla escolha para prática e validação rápida.</p>
        </div>
      </div>
    </section>
  );
}
