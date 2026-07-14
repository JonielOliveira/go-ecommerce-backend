import { api } from "./client";
import type { AuthUser, LoginRequest, LoginResponse, RegisterRequest, UserResponse } from "./types";

export function register(payload: RegisterRequest) {
  return api.post<UserResponse>("/auth/register", payload);
}

export function login(payload: LoginRequest) {
  return api.post<LoginResponse>("/auth/login", payload);
}

export function logout() {
  return api.post<void>("/auth/logout");
}

// skipAuthEvent: essa checagem roda no boot do app; um 401 aqui é o estado
// normal de "visitante não logado", não uma sessão expirada a ser tratada.
export function me() {
  return api.get<AuthUser>("/auth/me", { skipAuthEvent: true });
}
