import { api, buildQuery } from "./client";
import type { CreateOrderRequest, OrderResponse, OrderSearchParams, PageResponse } from "./types";

export function createOrder(payload: CreateOrderRequest) {
  return api.post<OrderResponse>("/orders", payload);
}

export function searchOrders(params: OrderSearchParams) {
  return api.get<PageResponse<OrderResponse>>(`/orders${buildQuery(params)}`);
}

export function getOrder(id: string) {
  return api.get<OrderResponse>(`/orders/${id}`);
}

export function payOrder(id: string) {
  return api.post<OrderResponse>(`/orders/${id}/pay`);
}

export function cancelOrder(id: string) {
  return api.post<OrderResponse>(`/orders/${id}/cancel`);
}
