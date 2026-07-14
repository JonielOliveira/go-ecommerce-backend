import {
  ActionIcon,
  Badge,
  Button,
  Card,
  Group,
  Loader,
  NumberInput,
  Select,
  Stack,
  Table,
  Text,
  TextInput,
  Title,
  Tooltip,
} from "@mantine/core";
import { useDebouncedValue, useDisclosure } from "@mantine/hooks";
import { modals } from "@mantine/modals";
import { notifications } from "@mantine/notifications";
import {
  IconArrowBackUp,
  IconEdit,
  IconPlus,
  IconShoppingCartPlus,
  IconTrash,
} from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";
import { ApiError } from "../api/client";
import {
  activateProduct,
  deactivateProduct,
  deleteProduct,
  restoreProduct,
  searchProducts,
} from "../api/products";
import type { DeletionState, ProductResponse, ProductSearchParams } from "../api/types";
import { ActiveBadge } from "../components/StatusBadge";
import { PageControls } from "../components/PageControls";
import { useAuth } from "../context/AuthContext";
import { useCart } from "../context/CartContext";
import { ProductFormModal } from "./products/ProductFormModal";

const DEFAULT_PAGE_SIZE = 20;

export function ProductsPage() {
  const { user } = useAuth();
  const { addItem } = useCart();
  const isAdmin = user?.role === "admin";

  const [nameFilter, setNameFilter] = useState("");
  const [debouncedName] = useDebouncedValue(nameFilter, 400);
  const [categoryFilter, setCategoryFilter] = useState("");
  const [minPrice, setMinPrice] = useState<number | "">("");
  const [maxPrice, setMaxPrice] = useState<number | "">("");
  const [activeFilter, setActiveFilter] = useState<string>("true");
  const [deletionState, setDeletionState] = useState<DeletionState>("not_deleted");

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);

  const [items, setItems] = useState<ProductResponse[]>([]);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);

  const [formOpened, { open: openForm, close: closeForm }] = useDisclosure(false);
  const [editingProduct, setEditingProduct] = useState<ProductResponse | null>(null);
  const [quantities, setQuantities] = useState<Record<string, number>>({});

  const params: ProductSearchParams = useMemo(
    () => ({
      name: debouncedName || undefined,
      categoryId: categoryFilter || undefined,
      minPrice: minPrice === "" ? undefined : minPrice,
      maxPrice: maxPrice === "" ? undefined : maxPrice,
      active: isAdmin ? (activeFilter === "all" ? undefined : activeFilter === "true") : true,
      deletionState: isAdmin ? deletionState : "not_deleted",
      page,
      pageSize,
    }),
    [debouncedName, categoryFilter, minPrice, maxPrice, activeFilter, deletionState, isAdmin, page, pageSize],
  );

  const load = () => {
    setLoading(true);
    searchProducts(params)
      .then((response) => {
        setItems(response.items);
        setTotalItems(response.totalItems);
        setTotalPages(response.totalPages);
      })
      .catch((err) => {
        notifications.show({
          title: "Erro ao carregar produtos",
          message: err instanceof ApiError ? err.message : "Tente novamente.",
          color: "red",
        });
      })
      .finally(() => setLoading(false));
  };

  useEffect(load, [params]);

  useEffect(() => {
    setPage(1);
  }, [debouncedName, categoryFilter, minPrice, maxPrice, activeFilter, deletionState]);

  const handleCreate = () => {
    setEditingProduct(null);
    openForm();
  };

  const handleEdit = (product: ProductResponse) => {
    setEditingProduct(product);
    openForm();
  };

  const handleAddToCart = (product: ProductResponse) => {
    const quantity = quantities[product.id] ?? 1;
    addItem(product, quantity);
    notifications.show({ message: `${product.name} adicionado ao carrinho.`, color: "green" });
  };

  const runAction = async (action: () => Promise<void>, successMessage: string) => {
    try {
      await action();
      notifications.show({ message: successMessage, color: "green" });
      load();
    } catch (err) {
      notifications.show({
        title: "Não foi possível concluir a ação",
        message: err instanceof ApiError ? err.message : "Tente novamente.",
        color: "red",
      });
    }
  };

  const confirmDelete = (product: ProductResponse) => {
    modals.openConfirmModal({
      title: "Excluir produto",
      children: <Text size="sm">Tem certeza que deseja excluir "{product.name}"?</Text>,
      labels: { confirm: "Excluir", cancel: "Cancelar" },
      confirmProps: { color: "red" },
      onConfirm: () => runAction(() => deleteProduct(product.id), "Produto excluído."),
    });
  };

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>Produtos</Title>
        {isAdmin && (
          <Button leftSection={<IconPlus size={16} />} onClick={handleCreate}>
            Novo produto
          </Button>
        )}
      </Group>

      <Card withBorder>
        <Group align="flex-end" wrap="wrap">
          <TextInput
            label="Nome"
            placeholder="Buscar por nome"
            value={nameFilter}
            onChange={(e) => setNameFilter(e.currentTarget.value)}
            w={200}
          />
          <TextInput
            label="Categoria (ID)"
            placeholder="opcional"
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.currentTarget.value)}
            w={180}
          />
          <NumberInput
            label="Preço mínimo"
            value={minPrice}
            onChange={(value) => setMinPrice(value === "" ? "" : Number(value))}
            min={0}
            prefix="R$ "
            w={140}
          />
          <NumberInput
            label="Preço máximo"
            value={maxPrice}
            onChange={(value) => setMaxPrice(value === "" ? "" : Number(value))}
            min={0}
            prefix="R$ "
            w={140}
          />
          {isAdmin && (
            <>
              <Select
                label="Status"
                data={[
                  { value: "true", label: "Ativos" },
                  { value: "false", label: "Inativos" },
                  { value: "all", label: "Todos" },
                ]}
                value={activeFilter}
                onChange={(value) => value && setActiveFilter(value)}
                w={130}
                allowDeselect={false}
              />
              <Select
                label="Exclusão"
                data={[
                  { value: "not_deleted", label: "Não excluídos" },
                  { value: "deleted", label: "Excluídos" },
                  { value: "all", label: "Todos" },
                ]}
                value={deletionState}
                onChange={(value) => value && setDeletionState(value as DeletionState)}
                w={160}
                allowDeselect={false}
              />
            </>
          )}
        </Group>
      </Card>

      <Card withBorder p={0}>
        <Table.ScrollContainer minWidth={800}>
          <Table striped highlightOnHover verticalSpacing="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Nome</Table.Th>
                <Table.Th>Preço</Table.Th>
                <Table.Th>Estoque</Table.Th>
                <Table.Th>Status</Table.Th>
                <Table.Th ta="right">Ações</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {loading ? (
                <Table.Tr>
                  <Table.Td colSpan={5}>
                    <Group justify="center" py="lg">
                      <Loader size="sm" />
                    </Group>
                  </Table.Td>
                </Table.Tr>
              ) : items.length === 0 ? (
                <Table.Tr>
                  <Table.Td colSpan={5}>
                    <Text ta="center" c="dimmed" py="lg">
                      Nenhum produto encontrado.
                    </Text>
                  </Table.Td>
                </Table.Tr>
              ) : (
                items.map((product) => (
                  <Table.Tr key={product.id}>
                    <Table.Td>
                      <Text fw={500}>{product.name}</Text>
                      <Text size="xs" c="dimmed" lineClamp={1}>
                        {product.description}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      {product.price.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })}
                    </Table.Td>
                    <Table.Td>
                      {product.stock === 0 ? (
                        <Badge color="red">Sem estoque</Badge>
                      ) : (
                        product.stock
                      )}
                    </Table.Td>
                    <Table.Td>
                      <Group gap={4}>
                        <ActiveBadge active={product.active} />
                        {product.deletedAt && <Badge color="red">Excluído</Badge>}
                      </Group>
                    </Table.Td>
                    <Table.Td>
                      <Group justify="flex-end" gap="xs" wrap="nowrap">
                        {!isAdmin && (
                          <>
                            <NumberInput
                              value={quantities[product.id] ?? 1}
                              onChange={(value) =>
                                setQuantities((q) => ({ ...q, [product.id]: Number(value) || 1 }))
                              }
                              min={1}
                              max={product.stock || 1}
                              w={70}
                              size="xs"
                              disabled={product.stock === 0 || !product.active}
                            />
                            <Tooltip label="Adicionar ao carrinho">
                              <ActionIcon
                                variant="light"
                                onClick={() => handleAddToCart(product)}
                                disabled={product.stock === 0 || !product.active}
                              >
                                <IconShoppingCartPlus size={16} />
                              </ActionIcon>
                            </Tooltip>
                          </>
                        )}
                        {isAdmin && (
                          <>
                            <Tooltip label="Editar">
                              <ActionIcon variant="light" onClick={() => handleEdit(product)}>
                                <IconEdit size={16} />
                              </ActionIcon>
                            </Tooltip>
                            {product.deletedAt ? (
                              <Tooltip label="Restaurar">
                                <ActionIcon
                                  variant="light"
                                  color="teal"
                                  onClick={() =>
                                    runAction(() => restoreProduct(product.id), "Produto restaurado.")
                                  }
                                >
                                  <IconArrowBackUp size={16} />
                                </ActionIcon>
                              </Tooltip>
                            ) : (
                              <>
                                <Tooltip label={product.active ? "Desativar" : "Ativar"}>
                                  <ActionIcon
                                    variant="light"
                                    color={product.active ? "orange" : "teal"}
                                    onClick={() =>
                                      runAction(
                                        () =>
                                          product.active
                                            ? deactivateProduct(product.id)
                                            : activateProduct(product.id),
                                        product.active ? "Produto desativado." : "Produto ativado.",
                                      )
                                    }
                                  >
                                    {product.active ? "⏸" : "▶"}
                                  </ActionIcon>
                                </Tooltip>
                                <Tooltip label="Excluir">
                                  <ActionIcon variant="light" color="red" onClick={() => confirmDelete(product)}>
                                    <IconTrash size={16} />
                                  </ActionIcon>
                                </Tooltip>
                              </>
                            )}
                          </>
                        )}
                      </Group>
                    </Table.Td>
                  </Table.Tr>
                ))
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

      <ProductFormModal
        opened={formOpened}
        onClose={closeForm}
        onSaved={load}
        product={editingProduct}
      />
    </Stack>
  );
}
