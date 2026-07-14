import { Table, Text } from "@mantine/core";
import { useProduct } from "../../hooks/useProduct";
import type { OrderItemResponse } from "../../api/types";

function OrderItemRow({ item }: { item: OrderItemResponse }) {
  const { product, loading } = useProduct(item.product_id);

  return (
    <Table.Tr>
      <Table.Td>{loading ? "Carregando..." : product?.name ?? item.product_id}</Table.Td>
      <Table.Td>{item.quantity}</Table.Td>
      <Table.Td>{item.unit_price.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })}</Table.Td>
      <Table.Td>{item.subtotal.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })}</Table.Td>
    </Table.Tr>
  );
}

export function OrderItemsTable({ items }: { items: OrderItemResponse[] }) {
  if (items.length === 0) {
    return (
      <Text size="sm" c="dimmed">
        Nenhum item.
      </Text>
    );
  }

  return (
    <Table withTableBorder withColumnBorders>
      <Table.Thead>
        <Table.Tr>
          <Table.Th>Produto</Table.Th>
          <Table.Th>Quantidade</Table.Th>
          <Table.Th>Preço unitário</Table.Th>
          <Table.Th>Subtotal</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {items.map((item) => (
          <OrderItemRow key={item.id} item={item} />
        ))}
      </Table.Tbody>
    </Table>
  );
}
