import { api, buildQuery } from "./client";
import type { CreateUserRequest, PageResponse, UserResponse, UserSearchParams, UserUpdateRequest } from "./types";

export function searchUsers(params: UserSearchParams) {
  return api.get<PageResponse<UserResponse>>(`/users${buildQuery(params)}`);
}

export function getUser(id: string) {
  return api.get<UserResponse>(`/users/${id}`);
}

export function createUser(payload: CreateUserRequest) {
  return api.post<UserResponse>("/users", payload);
}

export function updateUser(id: string, payload: UserUpdateRequest) {
  return api.put<UserResponse>(`/users/${id}`, payload);
}

export function deleteUser(id: string) {
  return api.delete<void>(`/users/${id}`);
}

export function restoreUser(id: string) {
  return api.patch<void>(`/users/${id}/restore`);
}

export function activateUser(id: string) {
  return api.patch<void>(`/users/${id}/activate`);
}

export function deactivateUser(id: string) {
  return api.patch<void>(`/users/${id}/deactivate`);
}
