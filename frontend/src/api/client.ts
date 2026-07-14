const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api/v1";

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

// Disparado sempre que a API responde 401. O AuthContext escuta esse evento
// para limpar o usuário autenticado sem que cada chamada precise saber disso.
const UNAUTHORIZED_EVENT = "auth:unauthorized";

export function onUnauthorized(handler: () => void): () => void {
  const listener = () => handler();
  window.addEventListener(UNAUTHORIZED_EVENT, listener);
  return () => window.removeEventListener(UNAUTHORIZED_EVENT, listener);
}

async function request<T>(path: string, options: RequestInit = {}, skipAuthEvent = false): Promise<T> {
  const response = await fetch(`${BASE_URL}${path}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (response.status === 401 && !skipAuthEvent) {
    window.dispatchEvent(new Event(UNAUTHORIZED_EVENT));
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const contentType = response.headers.get("content-type") ?? "";
  const data = contentType.includes("application/json") ? await response.json() : undefined;

  if (!response.ok) {
    const message = (data && typeof data === "object" && "error" in data ? String(data.error) : undefined) ?? response.statusText ?? "Erro inesperado";
    throw new ApiError(response.status, message);
  }

  return data as T;
}

export function buildQuery(params: object): string {
  const searchParams = new URLSearchParams();

  for (const [key, value] of Object.entries(params as Record<string, unknown>)) {
    if (value === undefined || value === null || value === "") continue;
    searchParams.set(key, String(value));
  }

  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export const api = {
  get: <T,>(path: string, options?: { skipAuthEvent?: boolean }) =>
    request<T>(path, { method: "GET" }, options?.skipAuthEvent),
  post: <T,>(path: string, body?: unknown) =>
    request<T>(path, { method: "POST", body: body !== undefined ? JSON.stringify(body) : undefined }),
  put: <T,>(path: string, body?: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body) }),
  patch: <T,>(path: string) => request<T>(path, { method: "PATCH" }),
  delete: <T,>(path: string) => request<T>(path, { method: "DELETE" }),
};
