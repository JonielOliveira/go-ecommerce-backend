import { Button, Group, Modal, PasswordInput, Select, Stack, TextInput } from "@mantine/core";
import { useForm } from "@mantine/form";
import { notifications } from "@mantine/notifications";
import { useEffect, useState } from "react";
import { ApiError } from "../../api/client";
import { createUser, updateUser } from "../../api/users";
import type { UserResponse, UserRole } from "../../api/types";

interface UserFormValues {
  name: string;
  email: string;
  password: string;
  role: UserRole;
  avatarUrl: string;
}

interface UserFormModalProps {
  opened: boolean;
  onClose: () => void;
  onSaved: () => void;
  user: UserResponse | null;
}

const EMPTY_VALUES: UserFormValues = {
  name: "",
  email: "",
  password: "",
  role: "customer",
  avatarUrl: "",
};

export function UserFormModal({ opened, onClose, onSaved, user }: UserFormModalProps) {
  const [submitting, setSubmitting] = useState(false);
  const isEditing = !!user;

  const form = useForm<UserFormValues>({
    initialValues: EMPTY_VALUES,
    validate: {
      name: (value) => (value.trim().length > 0 ? null : "Informe o nome"),
      email: (value) => (/^\S+@\S+\.\S+$/.test(value) ? null : "Informe um e-mail válido"),
      password: (value) =>
        !isEditing && value.length < 8 ? "A senha deve ter ao menos 8 caracteres" : null,
    },
  });

  useEffect(() => {
    if (!opened) return;

    if (user) {
      form.setValues({
        name: user.name,
        email: user.email,
        password: "",
        role: user.role,
        avatarUrl: user.avatarUrl ?? "",
      });
    } else {
      form.setValues(EMPTY_VALUES);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [opened, user]);

  const handleSubmit = async (values: UserFormValues) => {
    setSubmitting(true);
    try {
      if (user) {
        await updateUser(user.id, {
          name: values.name,
          email: values.email,
          password: values.password || undefined,
          role: values.role,
          avatarUrl: values.avatarUrl.trim() || null,
        });
        notifications.show({ message: "Usuário atualizado com sucesso.", color: "green" });
      } else {
        await createUser({
          name: values.name,
          email: values.email,
          password: values.password,
          role: values.role,
          avatarUrl: values.avatarUrl.trim() || null,
        });
        notifications.show({ message: "Usuário criado com sucesso.", color: "green" });
      }

      onSaved();
      onClose();
    } catch (err) {
      notifications.show({
        title: "Erro ao salvar usuário",
        message: err instanceof ApiError ? err.message : "Tente novamente.",
        color: "red",
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal opened={opened} onClose={onClose} title={user ? "Editar usuário" : "Novo usuário"} centered>
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack>
          <TextInput label="Nome" required {...form.getInputProps("name")} />
          <TextInput label="E-mail" required {...form.getInputProps("email")} />
          <PasswordInput
            label={isEditing ? "Nova senha" : "Senha"}
            placeholder={isEditing ? "Deixe em branco para não alterar" : "Mínimo 8 caracteres"}
            required={!isEditing}
            {...form.getInputProps("password")}
          />
          <Select
            label="Papel"
            data={[
              { value: "customer", label: "Customer" },
              { value: "admin", label: "Admin" },
            ]}
            allowDeselect={false}
            {...form.getInputProps("role")}
          />
          <TextInput label="URL do avatar" placeholder="opcional" {...form.getInputProps("avatarUrl")} />

          <Group justify="flex-end" mt="sm">
            <Button variant="default" onClick={onClose}>
              Cancelar
            </Button>
            <Button type="submit" loading={submitting}>
              Salvar
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
}
