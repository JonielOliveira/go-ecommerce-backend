import {
  ActionIcon,
  Badge,
  Button,
  Card,
  Group,
  Loader,
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
import { IconArrowBackUp, IconEdit, IconPlus, IconTrash } from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";
import { ApiError } from "../api/client";
import { activateUser, deactivateUser, deleteUser, restoreUser, searchUsers } from "../api/users";
import type { DeletionState, UserResponse, UserRole, UserSearchParams } from "../api/types";
import { PageControls } from "../components/PageControls";
import { ActiveBadge, RoleBadge } from "../components/StatusBadge";
import { useAuth } from "../context/AuthContext";
import { UserFormModal } from "./users/UserFormModal";

const DEFAULT_PAGE_SIZE = 20;

export function UsersPage() {
  const { user: currentUser } = useAuth();

  const [nameFilter, setNameFilter] = useState("");
  const [debouncedName] = useDebouncedValue(nameFilter, 400);
  const [emailFilter, setEmailFilter] = useState("");
  const [debouncedEmail] = useDebouncedValue(emailFilter, 400);
  const [roleFilter, setRoleFilter] = useState<UserRole | "">("");
  const [activeFilter, setActiveFilter] = useState<string>("all");
  const [deletionState, setDeletionState] = useState<DeletionState>("not_deleted");

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);

  const [items, setItems] = useState<UserResponse[]>([]);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);

  const [formOpened, { open: openForm, close: closeForm }] = useDisclosure(false);
  const [editingUser, setEditingUser] = useState<UserResponse | null>(null);

  const params: UserSearchParams = useMemo(
    () => ({
      name: debouncedName || undefined,
      email: debouncedEmail || undefined,
      role: roleFilter || undefined,
      active: activeFilter === "all" ? undefined : activeFilter === "true",
      deletionState,
      page,
      pageSize,
    }),
    [debouncedName, debouncedEmail, roleFilter, activeFilter, deletionState, page, pageSize],
  );

  const load = () => {
    setLoading(true);
    searchUsers(params)
      .then((response) => {
        setItems(response.items);
        setTotalItems(response.totalItems);
        setTotalPages(response.totalPages);
      })
      .catch((err) => {
        notifications.show({
          title: "Erro ao carregar usuários",
          message: err instanceof ApiError ? err.message : "Tente novamente.",
          color: "red",
        });
      })
      .finally(() => setLoading(false));
  };

  useEffect(load, [params]);

  useEffect(() => {
    setPage(1);
  }, [debouncedName, debouncedEmail, roleFilter, activeFilter, deletionState]);

  const handleCreate = () => {
    setEditingUser(null);
    openForm();
  };

  const handleEdit = (user: UserResponse) => {
    setEditingUser(user);
    openForm();
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

  const confirmDelete = (user: UserResponse) => {
    modals.openConfirmModal({
      title: "Excluir usuário",
      children: <Text size="sm">Tem certeza que deseja excluir "{user.name}"?</Text>,
      labels: { confirm: "Excluir", cancel: "Cancelar" },
      confirmProps: { color: "red" },
      onConfirm: () => runAction(() => deleteUser(user.id), "Usuário excluído."),
    });
  };

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>Usuários</Title>
        <Button leftSection={<IconPlus size={16} />} onClick={handleCreate}>
          Novo usuário
        </Button>
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
            label="E-mail"
            placeholder="Buscar por e-mail"
            value={emailFilter}
            onChange={(e) => setEmailFilter(e.currentTarget.value)}
            w={200}
          />
          <Select
            label="Papel"
            data={[
              { value: "", label: "Todos" },
              { value: "customer", label: "Customer" },
              { value: "admin", label: "Admin" },
            ]}
            value={roleFilter}
            onChange={(value) => setRoleFilter((value as UserRole) || "")}
            w={140}
            allowDeselect={false}
          />
          <Select
            label="Status"
            data={[
              { value: "all", label: "Todos" },
              { value: "true", label: "Ativos" },
              { value: "false", label: "Inativos" },
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
        </Group>
      </Card>

      <Card withBorder p={0}>
        <Table.ScrollContainer minWidth={800}>
          <Table striped highlightOnHover verticalSpacing="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Nome</Table.Th>
                <Table.Th>E-mail</Table.Th>
                <Table.Th>Papel</Table.Th>
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
                      Nenhum usuário encontrado.
                    </Text>
                  </Table.Td>
                </Table.Tr>
              ) : (
                items.map((user) => (
                  <Table.Tr key={user.id}>
                    <Table.Td>{user.name}</Table.Td>
                    <Table.Td>{user.email}</Table.Td>
                    <Table.Td>
                      <RoleBadge role={user.role} />
                    </Table.Td>
                    <Table.Td>
                      <Group gap={4}>
                        <ActiveBadge active={user.active} />
                        {user.deletedAt && <Badge color="red">Excluído</Badge>}
                      </Group>
                    </Table.Td>
                    <Table.Td>
                      <Group justify="flex-end" gap="xs" wrap="nowrap">
                        <Tooltip label="Editar">
                          <ActionIcon variant="light" onClick={() => handleEdit(user)}>
                            <IconEdit size={16} />
                          </ActionIcon>
                        </Tooltip>
                        {user.deletedAt ? (
                          <Tooltip label="Restaurar">
                            <ActionIcon
                              variant="light"
                              color="teal"
                              onClick={() => runAction(() => restoreUser(user.id), "Usuário restaurado.")}
                            >
                              <IconArrowBackUp size={16} />
                            </ActionIcon>
                          </Tooltip>
                        ) : (
                          <>
                            <Tooltip label={user.active ? "Desativar" : "Ativar"}>
                              <ActionIcon
                                variant="light"
                                color={user.active ? "orange" : "teal"}
                                disabled={user.id === currentUser?.id}
                                onClick={() =>
                                  runAction(
                                    () => (user.active ? deactivateUser(user.id) : activateUser(user.id)),
                                    user.active ? "Usuário desativado." : "Usuário ativado.",
                                  )
                                }
                              >
                                {user.active ? "⏸" : "▶"}
                              </ActionIcon>
                            </Tooltip>
                            <Tooltip label="Excluir">
                              <ActionIcon
                                variant="light"
                                color="red"
                                disabled={user.id === currentUser?.id}
                                onClick={() => confirmDelete(user)}
                              >
                                <IconTrash size={16} />
                              </ActionIcon>
                            </Tooltip>
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

      <UserFormModal opened={formOpened} onClose={closeForm} onSaved={load} user={editingUser} />
    </Stack>
  );
}
