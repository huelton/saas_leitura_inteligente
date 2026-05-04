const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const tokenKey = "leitura_token";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(tokenKey);
}

export function setToken(t: string) {
  localStorage.setItem(tokenKey, t);
}

export function clearToken() {
  localStorage.removeItem(tokenKey);
}

export async function api(
  path: string,
  opts: RequestInit & { json?: unknown } = {}
): Promise<Response> {
  const headers: HeadersInit = {
    ...(opts.headers || {}),
  };
  const t = getToken();
  if (t) {
    (headers as Record<string, string>)["Authorization"] = `Bearer ${t}`;
  }
  let body = opts.body;
  if (opts.json !== undefined) {
    (headers as Record<string, string>)["Content-Type"] = "application/json";
    body = JSON.stringify(opts.json);
  }
  return fetch(`${API}${path}`, { ...opts, headers, body });
}

export { API };
