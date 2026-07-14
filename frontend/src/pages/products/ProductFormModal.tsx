import { Button, Group, Modal, NumberInput, Stack, Textarea, TextInput } from "@mantine/core";
import { useForm } from "@mantine/form";
import { notifications } from "@mantine/notifications";
import { useEffect, useState } from "react";
import { createProduct, updateProduct } from "../../api/products";
import { ApiError } from "../../api/client";
import type { ProductResponse } from "../../api/types";

interface ProductFormValues {
  name: string;
  description: string;
  price: number | "";
  stock: number | "";
  categoryId: string;
  imageUrl: string;
}

interface ProductFormModalProps {
  opened: boolean;
  onClose: () => void;
  onSaved: () => void;
  product: ProductResponse | null;
}

const EMPTY_VALUES: ProductFormValues = {
  name: "",
  description: "",
  price: "",
  stock: "",
  categoryId: "",
  imageUrl: "",
};

export function ProductFormModal({ opened, onClose, onSaved, product }: ProductFormModalProps) {
  const [submitting, setSubmitting] = useState(false);

  const form = useForm<ProductFormValues>({
    initialValues: EMPTY_VALUES,
    validate: {
      name: (value) => (value.trim().length > 0 ? null : "Informe o nome"),
      description: (value) => (value.trim().length > 0 ? null : "Informe a descrição"),
      price: (value) => (typeof value === "number" && value > 0 ? null : "Preço deve ser maior que zero"),
      stock: (value) => (typeof value === "number" && value >= 0 ? null : "Estoque não pode ser negativo"),
    },
  });

  useEffect(() => {
    if (!opened) return;

    if (product) {
      form.setValues({
        name: product.name,
        description: product.description,
        price: product.price,
        stock: product.stock,
        categoryId: product.categoryId ?? "",
        imageUrl: product.imageUrl ?? "",
      });
    } else {
      form.setValues(EMPTY_VALUES);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [opened, product]);

  const handleSubmit = async (values: ProductFormValues) => {
    setSubmitting(true);
    try {
      const payload = {
        name: values.name,
        description: values.description,
        price: Number(values.price),
        stock: Number(values.stock),
        categoryId: values.categoryId.trim() || null,
        imageUrl: values.imageUrl.trim() || null,
      };

      if (product) {
        await updateProduct(product.id, payload);
        notifications.show({ message: "Produto atualizado com sucesso.", color: "green" });
      } else {
        await createProduct(payload);
        notifications.show({ message: "Produto criado com sucesso.", color: "green" });
      }

      onSaved();
      onClose();
    } catch (err) {
      notifications.show({
        title: "Erro ao salvar produto",
        message: err instanceof ApiError ? err.message : "Tente novamente.",
        color: "red",
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal opened={opened} onClose={onClose} title={product ? "Editar produto" : "Novo produto"} centered>
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack>
          <TextInput label="Nome" required {...form.getInputProps("name")} />
          <Textarea label="Descrição" required minRows={2} {...form.getInputProps("description")} />
          <Group grow>
            <NumberInput
              label="Preço"
              required
              min={0.01}
              step={0.01}
              decimalScale={2}
              fixedDecimalScale
              prefix="R$ "
              {...form.getInputProps("price")}
            />
            <NumberInput label="Estoque" required min={0} step={1} {...form.getInputProps("stock")} />
          </Group>
          <TextInput label="ID da categoria" placeholder="opcional" {...form.getInputProps("categoryId")} />
          <TextInput label="URL da imagem" placeholder="opcional" {...form.getInputProps("imageUrl")} />

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
