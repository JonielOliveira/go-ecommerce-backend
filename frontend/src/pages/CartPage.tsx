import {
  ActionIcon,
  Alert,
  Button,
  Card,
  Group,
  NumberInput,
  Stack,
  Table,
  Text,
  Title,
} from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconInfoCircle, IconTrash } from "@tabler/icons-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { createOrder } from "../api/orders";
import { useCart } from "../context/CartContext";

export function CartPage() {
  const { items, totalAmount, updateQuantity, removeItem, clear } = useCart();
  const [submitting, setSubmitting] = useState(false);
  const navigate = useNavigate();

  const handleCheckout = async () => {
    setSubmitting(true);
    try {
      await createOrder({
        items: items.map((item) => ({ product_id: item.productId, quantity: item.quantity })),
      });
      clear();
      notifications.show({
        title: "Pedido criado",
        message: "Seu pedido foi criado como pendente. Finalize o pagamento em 'Pedidos'.",
        color: "green",
      });
      navigate("/pedidos");
    } catch (err) {
      notifications.show({
        title: "Não foi possível finalizar o pedido",
        message: err instanceof ApiError ? err.message : "Tente novamente.",
        color: "red",
      });
    } finally {
      setSubmitting(false);
    }
  };

  if (items.length === 0) {
    return (
      <Stack>
        <Title order={2}>Carrinho</Title>
        <Alert icon={<IconInfoCircle size={16} />} color="blue" variant="light">
          Seu carrinho está vazio. Adicione produtos na aba "Produtos".
        </Alert>
      </Stack>
    );
  }

  return (
    <Stack>
      <Title order={2}>Carrinho</Title>

      <Card withBorder p={0}>
        <Table.ScrollContainer minWidth={600}>
          <Table striped verticalSpacing="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Produto</Table.Th>
                <Table.Th>Preço</Table.Th>
                <Table.Th>Quantidade</Table.Th>
                <Table.Th>Subtotal</Table.Th>
                <Table.Th />
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {items.map((item) => (
                <Table.Tr key={item.productId}>
                  <Table.Td>{item.name}</Table.Td>
                  <Table.Td>
                    {item.price.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })}
                  </Table.Td>
                  <Table.Td>
                    <NumberInput
                      value={item.quantity}
                      onChange={(value) => updateQuantity(item.productId, Number(value) || 1)}
                      min={1}
                      max={item.stock}
                      w={90}
                      size="xs"
                    />
                  </Table.Td>
                  <Table.Td>
                    {(item.price * item.quantity).toLocaleString("pt-BR", {
                      style: "currency",
                      currency: "BRL",
                    })}
                  </Table.Td>
                  <Table.Td>
                    <ActionIcon variant="light" color="red" onClick={() => removeItem(item.productId)}>
                      <IconTrash size={16} />
                    </ActionIcon>
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </Table.ScrollContainer>
      </Card>

      <Group justify="space-between" align="center">
        <Button variant="default" onClick={clear}>
          Esvaziar carrinho
        </Button>

        <Group gap="lg">
          <Text size="lg" fw={600}>
            Total: {totalAmount.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })}
          </Text>
          <Button size="md" loading={submitting} onClick={handleCheckout}>
            Finalizar pedido
          </Button>
        </Group>
      </Group>
    </Stack>
  );
}
