import { AppShell, Avatar, Burger, Group, Indicator, Menu, Tabs, Text, Title, UnstyledButton } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { IconChevronDown, IconLogout, IconShoppingCart } from "@tabler/icons-react";
import type { ReactNode } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { useCart } from "../context/CartContext";

const TABS = [
  { value: "/produtos", label: "Produtos" },
  { value: "/pedidos", label: "Pedidos" },
  { value: "/carrinho", label: "Carrinho" },
] as const;

const ADMIN_TABS = [{ value: "/usuarios", label: "Usuários" }] as const;

export function AppLayout({ children }: { children: ReactNode }) {
  const [opened, { toggle }] = useDisclosure();
  const { user, logout } = useAuth();
  const { totalCount } = useCart();
  const navigate = useNavigate();
  const location = useLocation();

  const tabs = user?.role === "admin" ? [...TABS, ...ADMIN_TABS] : TABS;

  const activeTab = tabs.find((tab) => location.pathname.startsWith(tab.value))?.value ?? tabs[0].value;

  const handleLogout = async () => {
    await logout();
    navigate("/login", { replace: true });
  };

  return (
    <AppShell header={{ height: 64 }} padding="md">
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between" wrap="nowrap">
          <Group wrap="nowrap">
            <Burger opened={opened} onClick={toggle} hiddenFrom="sm" size="sm" />
            <Title order={3} c="blue">
              E-commerce
            </Title>
          </Group>

          <Tabs
            value={activeTab}
            onChange={(value) => value && navigate(value)}
            visibleFrom="sm"
            variant="pills"
          >
            <Tabs.List>
              {tabs.map((tab) => (
                <Tabs.Tab
                  key={tab.value}
                  value={tab.value}
                  leftSection={
                    tab.value === "/carrinho" && totalCount > 0 ? (
                      <Indicator label={totalCount} size={16} offset={-2}>
                        <IconShoppingCart size={16} />
                      </Indicator>
                    ) : tab.value === "/carrinho" ? (
                      <IconShoppingCart size={16} />
                    ) : undefined
                  }
                >
                  {tab.label}
                </Tabs.Tab>
              ))}
            </Tabs.List>
          </Tabs>

          <Menu shadow="md" width={200} position="bottom-end">
            <Menu.Target>
              <UnstyledButton>
                <Group gap={8} wrap="nowrap">
                  <Avatar radius="xl" color="blue">
                    {user?.name?.charAt(0).toUpperCase()}
                  </Avatar>
                  <div style={{ minWidth: 0 }}>
                    <Text size="sm" fw={500} truncate>
                      {user?.name}
                    </Text>
                    <Text size="xs" c="dimmed" tt="capitalize">
                      {user?.role}
                    </Text>
                  </div>
                  <IconChevronDown size={14} />
                </Group>
              </UnstyledButton>
            </Menu.Target>

            <Menu.Dropdown>
              <Menu.Item color="red" leftSection={<IconLogout size={16} />} onClick={handleLogout}>
                Sair
              </Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
      </AppShell.Header>

      <AppShell.Main bg="gray.0">{children}</AppShell.Main>
    </AppShell>
  );
}
