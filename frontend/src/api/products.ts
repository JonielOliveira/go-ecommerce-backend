import { api, buildQuery } from "./client";
import type { PageResponse, ProductRequest, ProductResponse, ProductSearchParams, ProductUpdateRequest } from "./types";

export function searchProducts(params: ProductSearchParams) {
  return api.get<PageResponse<ProductResponse>>(`/products${buildQuery(params)}`);
}

export function getProduct(id: string) {
  return api.get<ProductResponse>(`/products/${id}`);
}

export function createProduct(payload: ProductRequest) {
  return api.post<ProductResponse>("/products", payload);
}

export function updateProduct(id: string, payload: ProductUpdateRequest) {
  return api.put<ProductResponse>(`/products/${id}`, payload);
}

export function deleteProduct(id: string) {
  return api.delete<void>(`/products/${id}`);
}

export function restoreProduct(id: string) {
  return api.patch<void>(`/products/${id}/restore`);
}

export function activateProduct(id: string) {
  return api.patch<void>(`/products/${id}/activate`);
}

export function deactivateProduct(id: string) {
  return api.patch<void>(`/products/${id}/deactivate`);
}
