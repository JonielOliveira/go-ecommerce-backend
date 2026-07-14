import { Badge } from "@mantine/core";
import type { OrderStatus } from "../api/types";

const ORDER_STATUS_COLOR: Record<OrderStatus, string> = {
  PENDING: "yellow",
  PAID: "green",
  CANCELED: "red",
};

const ORDER_STATUS_LABEL: Record<OrderStatus, string> = {
  PENDING: "Pendente",
  PAID: "Pago",
  CANCELED: "Cancelado",
};

export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  return <Badge color={ORDER_STATUS_COLOR[status]}>{ORDER_STATUS_LABEL[status]}</Badge>;
}

export function ActiveBadge({ active }: { active: boolean }) {
  return <Badge color={active ? "green" : "gray"}>{active ? "Ativo" : "Inativo"}</Badge>;
}

export function DeletedBadge() {
  return <Badge color="red">Excluído</Badge>;
}

export function RoleBadge({ role }: { role: string }) {
  return (
    <Badge color={role === "admin" ? "grape" : "blue"} tt="capitalize">
      {role}
    </Badge>
  );
}
