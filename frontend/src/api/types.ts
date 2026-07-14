// Tipos espelhando os DTOs do backend (internal/dto/*.go).
// users/products usam camelCase; orders usa snake_case — mantido igual ao JSON real da API.

export type UserRole = "customer" | "admin";

export type DeletionState = "not_deleted" | "deleted" | "all";

export type OrderStatus = "PENDING" | "PAID" | "CANCELED";

export interface PageResponse<T> {
  items: T[];
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
}

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthUser {
  id: string;
  name: string;
  email: string;
  role: UserRole;
}

export interface LoginResponse {
  user: AuthUser;
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

export interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
  role?: UserRole;
  avatarUrl?: string | null;
}

export interface UserUpdateRequest {
  name: string;
  email: string;
  password?: string;
  role: UserRole;
  avatarUrl?: string | null;
}

export interface UserResponse {
  id: string;
  name: string;
  email: string;
  avatarUrl: string | null;
  role: UserRole;
  active: boolean;
  emailVerifiedAt: string | null;
  lastLoginAt: string | null;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
}

export interface UserSearchParams {
  name?: string;
  email?: string;
  role?: UserRole | "";
  active?: boolean;
  deletionState?: DeletionState;
  page?: number;
  pageSize?: number;
}

// ---------------------------------------------------------------------------
// Products
// ---------------------------------------------------------------------------

export interface ProductRequest {
  name: string;
  description: string;
  price: number;
  stock: number;
  categoryId?: string | null;
  imageUrl?: string | null;
}

export type ProductUpdateRequest = ProductRequest;

export interface ProductResponse {
  id: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  categoryId: string | null;
  imageUrl: string | null;
  active: boolean;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
}

export interface ProductSearchParams {
  name?: string;
  categoryId?: string;
  active?: boolean;
  deletionState?: DeletionState;
  minPrice?: number;
  maxPrice?: number;
  page?: number;
  pageSize?: number;
}

// ---------------------------------------------------------------------------
// Orders
// ---------------------------------------------------------------------------

export interface CreateOrderItemRequest {
  product_id: string;
  quantity: number;
}

export interface CreateOrderRequest {
  items: CreateOrderItemRequest[];
}

export interface OrderItemResponse {
  id: string;
  product_id: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
}

export interface OrderResponse {
  id: string;
  customer_id: string;
  status: OrderStatus;
  total_amount: number;
  items: OrderItemResponse[];
  paid_at: string | null;
  canceled_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface OrderSearchParams {
  page?: number;
  pageSize?: number;
}
