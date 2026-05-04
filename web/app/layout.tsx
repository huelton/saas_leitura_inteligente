import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "Leitura Inteligente",
  description: "Leitura com perguntas e compreensão",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="pt-BR">
      <body>
        <div className="app-shell">
          <header className="app-header">
            <div className="app-header-inner">
              <Link href="/" className="brand">
                Leitura Inteligente
              </Link>
              <nav className="nav">
                <Link href="/dashboard">Dashboard</Link>
                <Link href="/upload">Upload</Link>
                <Link href="/login">Login</Link>
              </nav>
            </div>
          </header>
          <main className="container">{children}</main>
        </div>
      </body>
    </html>
  );
}
