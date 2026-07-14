import { Group, Pagination, Select, Text } from "@mantine/core";

interface PageControlsProps {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

const PAGE_SIZE_OPTIONS = ["10", "20", "50", "100"];

export function PageControls({
  page,
  pageSize,
  totalItems,
  totalPages,
  onPageChange,
  onPageSizeChange,
}: PageControlsProps) {
  return (
    <Group justify="space-between" mt="md" wrap="wrap">
      <Text size="sm" c="dimmed">
        {totalItems} {totalItems === 1 ? "resultado" : "resultados"}
      </Text>

      <Group gap="sm">
        <Pagination
          value={page}
          onChange={onPageChange}
          total={Math.max(totalPages, 1)}
          size="sm"
        />
        <Select
          data={PAGE_SIZE_OPTIONS}
          value={String(pageSize)}
          onChange={(value) => value && onPageSizeChange(Number(value))}
          w={90}
          size="sm"
          allowDeselect={false}
        />
      </Group>
    </Group>
  );
}
