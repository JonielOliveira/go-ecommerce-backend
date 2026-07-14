import {
  ActionIcon,
  Alert,
  Button,
  Card,
  Collapse,
  Group,
  Loader,
  SegmentedControl,
  Stack,
  Table,
  Text,
  Title,
  Tooltip,
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { modals } from "@mantine/modals";
import { notifications } from "@mantine/notifications";
import { IconChevronDown, IconChevronRight, IconInfoCircle } from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";
import { ApiError } from "../api/client";
import { cancelOrder, payOrder, searchOrders } from "../api/orders";
import type { OrderResponse, OrderStatus } from "../api/types";
import { OrderStatusBadge } from "../components/StatusBadge";
import { PageControls } from "../components/PageControls";
import { useAuth } from "../context/AuthContext";
import { OrderItemsTable } from "./orders/OrderItemsTable";

const DEFAULT_PAGE_SIZE = 20;

function OrderRow({ order, onChanged }: { order: OrderResponse; onChanged: () => void }) {
  const { user } = useAuth();
  const [expanded, { toggle }] = useDisclosure(false);
  const isOwner = order.customer_id === user?.id;
  const isAdmin = user?.role === "admin";

  const runAction = async (action: () => Promise<OrderResponse>, successMessage: string) => {
    try {
      await action();
      notifications.show({ message: successMessage, color: "green" });
      onChanged();
    } catch (err) {
      notifications.show({
        title: "Não foi possível concluir a ação",
        message: err instanceof ApiError ? err.message : "Tente novamente.",
        color: "red",
      });
    }
  };

  const confirmCancel = () => {
    modals.openConfirmModal({
      title: "Cancelar pedido",
      children: <Text size="sm">Tem certeza que deseja cancelar este pedido? O estoque será devolvido.</Text>,
      labels: { confirm: "Cancelar pedido", cancel: "Voltar" },
      confirmProps: { color: "red" },
      onConfirm: () => runAction(() => cancelOrder(order.id), "Pedido cancelado."),
    });
  };

  return (
    <>
      <Table.Tr>
        <Table.Td>
          <ActionIcon variant="subtle" onClick={toggle} size="sm">
            {expanded ? <IconChevronDown size={16} /> : <IconChevronRight size={16} />}
          </ActionIcon>
        </Table.Td>
        <Table.Td>
          <Tooltip label={order.id}>
            <Text ff="monospace" size="sm">
              {order.id.slice(-8)}
            </Text>
          </Tooltip>
        </Table.Td>
        {isAdmin && (
          <Table.Td>
            <Tooltip label={order.customer_id}>
              <Text ff="monospace" size="sm">
                {order.customer_id.slice(-8)}
              </Text>
            </Tooltip>
          </Table.Td>
        )}
        <Table.Td>
          <OrderStatusBadge status={order.status} />
        </Table.Td>
        <Table.Td>{order.items.length}</Table.Td>
        <Table.Td>
          {order.total_amount.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })}
        </Table.Td>
        <Table.Td>{new Date(order.created_at).toLocaleString("pt-BR")}</Table.Td>
        <Table.Td>
          <Group justify="flex-end" gap="xs" wrap="nowrap">
            {order.status === "PENDING" && isOwner && (
              <Tooltip label="Pagar pedido">
                <Button
                  size="xs"
                  color="green"
                  onClick={() => runAction(() => payOrder(order.id), "Pedido pago com sucesso.")}
                >
                  Pagar
                </Button>
              </Tooltip>
            )}
            {order.status === "PENDING" && (isOwner || isAdmin) && (
              <Tooltip label="Cancelar pedido">
                <Button size="xs" color="red" variant="light" onClick={confirmCancel}>
                  Cancelar
                </Button>
              </Tooltip>
            )}
          </Group>
        </Table.Td>
      </Table.Tr>
      <Table.Tr>
        <Table.Td colSpan={isAdmin ? 8 : 7} p={0}>
          <Collapse expanded={expanded}>
            <div style={{ padding: 16 }}>
              <OrderItemsTable items={order.items} />
            </div>
          </Collapse>
        </Table.Td>
      </Table.Tr>
    </>
  );
}

export function OrdersPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [statusFilter, setStatusFilter] = useState<OrderStatus | "ALL">("ALL");

  const [items, setItems] = useState<OrderResponse[]>([]);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);

  const load = () => {
    setLoading(true);
    searchOrders({ page, pageSize })
      .then((response) => {
        setItems(response.items);
        setTotalItems(response.totalItems);
        setTotalPages(response.totalPages);
      })
      .catch((err) => {
        notifications.show({
          title: "Erro ao carregar pedidos",
          message: err instanceof ApiError ? err.message : "Tente novamente.",
          color: "red",
        });
      })
      .finally(() => setLoading(false));
  };

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(load, [page, pageSize]);

  const filteredItems = useMemo(
    () => (statusFilter === "ALL" ? items : items.filter((order) => order.status === statusFilter)),
    [items, statusFilter],
  );

  return (
    <Stack>
      <Title order={2}>{isAdmin ? "Todos os pedidos" : "Meus pedidos"}</Title>

      <Card withBorder>
        <Stack gap="xs">
          <Text size="sm" fw={500}>
            Filtrar por status
          </Text>
          <SegmentedControl
            value={statusFilter}
            onChange={(value) => setStatusFilter(value as OrderStatus | "ALL")}
            data={[
              { value: "ALL", label: "Todos" },
              { value: "PENDING", label: "Pendente" },
              { value: "PAID", label: "Pago" },
              { value: "CANCELED", label: "Cancelado" },
            ]}
          />
          <Alert icon={<IconInfoCircle size={16} />} color="blue" variant="light" p="xs">
            <Text size="xs">
              O filtro de status é aplicado apenas sobre os pedidos já carregados nesta página — a API de
              pedidos pagina por página/tamanho, sem filtro de status no servidor.
            </Text>
          </Alert>
        </Stack>
      </Card>

      <Card withBorder p={0}>
        <Table.ScrollContainer minWidth={isAdmin ? 900 : 800}>
          <Table striped highlightOnHover verticalSpacing="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th w={40} />
                <Table.Th>Pedido</Table.Th>
                {isAdmin && <Table.Th>Cliente</Table.Th>}
                <Table.Th>Status</Table.Th>
                <Table.Th>Itens</Table.Th>
                <Table.Th>Total</Table.Th>
                <Table.Th>Criado em</Table.Th>
                <Table.Th ta="right">Ações</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {loading ? (
                <Table.Tr>
                  <Table.Td colSpan={isAdmin ? 8 : 7}>
                    <Group justify="center" py="lg">
                      <Loader size="sm" />
                    </Group>
                  </Table.Td>
                </Table.Tr>
              ) : filteredItems.length === 0 ? (
                <Table.Tr>
                  <Table.Td colSpan={isAdmin ? 8 : 7}>
                    <Text ta="center" c="dimmed" py="lg">
                      Nenhum pedido encontrado.
                    </Text>
                  </Table.Td>
                </Table.Tr>
              ) : (
                filteredItems.map((order) => <OrderRow key={order.id} order={order} onChanged={load} />)
              )}
            </Table.Tbody>
          </Table>
        </Table.ScrollContainer>
      </Card>

      <PageControls
        page={page}
        pageSize={pageSize}
        totalItems={totalItems}
        totalPages={totalPages}
        onPageChange={setPage}
        onPageSizeChange={(size) => {
          setPageSize(size);
          setPage(1);
        }}
      />
    </Stack>
  );
}
