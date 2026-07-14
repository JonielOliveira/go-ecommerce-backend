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
import { IconAlertCircle } from "@tabler/icons-react";
import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { useAuth } from "../context/AuthContext";

interface LoginFormValues {
  email: string;
  password: string;
}

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const form = useForm<LoginFormValues>({
    initialValues: { email: "", password: "" },
    validate: {
      email: (value) => (/^\S+@\S+\.\S+$/.test(value) ? null : "Informe um e-mail válido"),
      password: (value) => (value.length >= 8 ? null : "A senha deve ter ao menos 8 caracteres"),
    },
  });

  const handleSubmit = async (values: LoginFormValues) => {
    setError(null);
    setSubmitting(true);
    try {
      await login(values.email, values.password);
      const redirectTo = (location.state as { from?: Location })?.from?.pathname ?? "/produtos";
      navigate(redirectTo, { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Não foi possível entrar. Tente novamente.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Center h="100vh" bg="gray.0">
      <Paper withBorder shadow="md" p="xl" radius="md" w={380}>
        <Title order={2} ta="center" mb="xs">
          Entrar
        </Title>
        <Text c="dimmed" size="sm" ta="center" mb="lg">
          Acesse sua conta para continuar
        </Text>

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack>
            {error && (
              <Alert color="red" icon={<IconAlertCircle size={16} />}>
                {error}
              </Alert>
            )}

            <TextInput
              label="E-mail"
              placeholder="voce@email.com"
              required
              {...form.getInputProps("email")}
            />
            <PasswordInput
              label="Senha"
              placeholder="Sua senha"
              required
              {...form.getInputProps("password")}
            />

            <Button type="submit" fullWidth loading={submitting} mt="sm">
              Entrar
            </Button>
          </Stack>
        </form>

        <Text ta="center" size="sm" mt="lg">
          Não tem uma conta?{" "}
          <Anchor component={Link} to="/registro">
            Cadastre-se
          </Anchor>
        </Text>
      </Paper>
    </Center>
  );
}
