import { useEffect, useState } from "react";
import { getProduct } from "../api/products";
import type { ProductResponse } from "../api/types";

// Pedidos só guardam product_id; a API pública GET /products/:id resolve o
// nome/imagem para exibição. Cache simples em memória (módulo), compartilhado
// entre todos os componentes que exibem itens de pedido/carrinho na sessão.
const cache = new Map<string, ProductResponse>();
const inflight = new Map<string, Promise<ProductResponse>>();

async function fetchProduct(id: string): Promise<ProductResponse> {
  const cached = cache.get(id);
  if (cached) return cached;

  const pending = inflight.get(id);
  if (pending) return pending;

  const promise = getProduct(id)
    .then((product) => {
      cache.set(id, product);
      inflight.delete(id);
      return product;
    })
    .catch((error) => {
      inflight.delete(id);
      throw error;
    });

  inflight.set(id, promise);
  return promise;
}

export function useProduct(id: string | undefined) {
  const [product, setProduct] = useState<ProductResponse | null>(id ? cache.get(id) ?? null : null);
  const [loading, setLoading] = useState(!!id && !cache.has(id));

  useEffect(() => {
    if (!id) return;

    const cached = cache.get(id);
    if (cached) {
      setProduct(cached);
      setLoading(false);
      return;
    }

    let active = true;
    setLoading(true);

    fetchProduct(id)
      .then((result) => {
        if (active) setProduct(result);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [id]);

  return { product, loading };
}
