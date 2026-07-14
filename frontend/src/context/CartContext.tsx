import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import type { ProductResponse } from "../api/types";
import { useAuth } from "./AuthContext";

export interface CartItem {
  productId: string;
  name: string;
  price: number;
  stock: number;
  quantity: number;
}

interface CartContextValue {
  items: CartItem[];
  totalCount: number;
  totalAmount: number;
  addItem: (product: ProductResponse, quantity: number) => void;
  updateQuantity: (productId: string, quantity: number) => void;
  removeItem: (productId: string) => void;
  clear: () => void;
}

const CartContext = createContext<CartContextValue | null>(null);

function storageKey(userId: string | undefined) {
  return `cart:${userId ?? "anon"}`;
}

export function CartProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const key = storageKey(user?.id);

  const [items, setItems] = useState<CartItem[]>(() => {
    try {
      const raw = localStorage.getItem(key);
      return raw ? (JSON.parse(raw) as CartItem[]) : [];
    } catch {
      return [];
    }
  });

  // Cada usuário tem seu próprio carrinho: ao trocar de sessão, recarrega do
  // localStorage da chave correspondente em vez de manter o estado anterior.
  useEffect(() => {
    try {
      const raw = localStorage.getItem(key);
      setItems(raw ? (JSON.parse(raw) as CartItem[]) : []);
    } catch {
      setItems([]);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key]);

  useEffect(() => {
    localStorage.setItem(key, JSON.stringify(items));
  }, [key, items]);

  const addItem = (product: ProductResponse, quantity: number) => {
    setItems((current) => {
      const existing = current.find((item) => item.productId === product.id);
      const maxQuantity = product.stock;

      if (existing) {
        const nextQuantity = Math.min(existing.quantity + quantity, maxQuantity);
        return current.map((item) =>
          item.productId === product.id ? { ...item, quantity: nextQuantity } : item,
        );
      }

      return [
        ...current,
        {
          productId: product.id,
          name: product.name,
          price: product.price,
          stock: product.stock,
          quantity: Math.min(quantity, maxQuantity),
        },
      ];
    });
  };

  const updateQuantity = (productId: string, quantity: number) => {
    setItems((current) =>
      current.map((item) =>
        item.productId === productId
          ? { ...item, quantity: Math.max(1, Math.min(quantity, item.stock)) }
          : item,
      ),
    );
  };

  const removeItem = (productId: string) => {
    setItems((current) => current.filter((item) => item.productId !== productId));
  };

  const clear = () => setItems([]);

  const totalCount = useMemo(() => items.reduce((sum, item) => sum + item.quantity, 0), [items]);
  const totalAmount = useMemo(
    () => items.reduce((sum, item) => sum + item.quantity * item.price, 0),
    [items],
  );

  return (
    <CartContext.Provider
      value={{ items, totalCount, totalAmount, addItem, updateQuantity, removeItem, clear }}
    >
      {children}
    </CartContext.Provider>
  );
}

export function useCart() {
  const context = useContext(CartContext);
  if (!context) throw new Error("useCart deve ser usado dentro de um CartProvider");
  return context;
}
