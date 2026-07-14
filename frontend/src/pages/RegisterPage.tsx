import {
  Alert,
  Anchor,
  Button,
  Center,
  Paper,
  PasswordInput,
  Stack,
  Text,
  TextInput,
  Title,
} from "@mantine/core";
import { useForm } from "@mantine/form";
import { notifications } from "@mantine/notifications";
import { IconAlertCircle } from "@tabler/icons-react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { useAuth } from "../context/AuthContext";

interface RegisterFormValues {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const form = useForm<RegisterFormValues>({
    initialValues: { name: "", email: "", password: "", confirmPassword: "" },
    validate: {
      name: (value) => (value.trim().length > 0 ? null : "Informe seu nome"),
      email: (value) => (/^\S+@\S+\.\S+$/.test(value) ? null : "Informe um e-mail válido"),
      password: (value) => (value.length >= 8 ? null : "A senha deve ter ao menos 8 caracteres"),
      confirmPassword: (value, values) => (value === values.password ? null : "As senhas não coincidem"),
    },
  });

  const handleSubmit = async (values: RegisterFormValues) => {
    setError(null);
    setSubmitting(true);
    try {
      await register(values.name, values.email, values.password);
      notifications.show({
        title: "Cadastro realizado",
        message: "Agora faça login com suas credenciais.",
        color: "green",
      });
      navigate("/login", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Não foi possível cadastrar. Tente novamente.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Center h="100vh" bg="gray.0">
      <Paper withBorder shadow="md" p="xl" radius="md" w={380}>
        <Title order={2} ta="center" mb="xs">
          Criar conta
        </Title>
        <Text c="dimmed" size="sm" ta="center" mb="lg">
          Cadastre-se para comprar na loja
        </Text>

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            {error && (
              <Alert color="red" icon={<IconAlertCircle size={16} />}>
                {error}
              </Alert>
            )}

            <TextInput label="Nome" placeholder="Seu nome" required {...form.getInputProps("name")} />
            <TextInput
              label="E-mail"
              placeholder="voce@email.com"
              required
              {...form.getInputProps("email")}
            />
            <PasswordInput
              label="Senha"
              placeholder="Mínimo 8 caracteres"
              required
              {...form.getInputProps("password")}
            />
            <PasswordInput
              label="Confirmar senha"
              placeholder="Repita a senha"
              required
              {...form.getInputProps("confirmPassword")}
            />

            <Button type="submit" fullWidth loading={submitting} mt="sm">
              Cadastrar
            </Button>
          </Stack>
        </form>

        <Text ta="center" size="sm" mt="lg">
          Já tem uma conta?{" "}
          <Anchor component={Link} to="/login">
            Entrar
          </Anchor>
        </Text>
      </Paper>
    </Center>
  );
}
