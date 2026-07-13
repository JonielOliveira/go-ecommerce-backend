package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"ecommerce/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func scanOrder(row pgx.Row) (*domain.Order, error) {
	var (
		order  domain.Order
		status string
	)

	if err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&status,
		&order.TotalAmount,
		&order.PaidAt,
		&order.CanceledAt,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return nil, err
	}

	order.Status = domain.OrderStatus(status)

	return &order, nil
}

func scanOrderItem(row pgx.Row) (domain.OrderItem, error) {
	var item domain.OrderItem

	err := row.Scan(
		&item.ID,
		&item.OrderID,
		&item.ProductID,
		&item.Quantity,
		&item.UnitPrice,
		&item.CreatedAt,
	)

	return item, err
}

func (r *PostgresOrderRepository) findItemsByOrderID(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	const query = `
		SELECT id, order_id, product_id, quantity, unit_price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at, id
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.OrderItem, 0)

	for rows.Next() {
		item, err := scanOrderItem(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// validateOwner confirma, dentro da transação, que o proprietário do
// pedido existe, está ativo e não foi excluído (regras de negócio 3, 5 e 6
// da TASK_3). Na prática o middleware de autenticação já garante isso a
// cada requisição, mas a checagem aqui cobre a janela entre a validação do
// JWT e a efetivação do pedido.
func (r *PostgresOrderRepository) validateOwner(ctx context.Context, tx pgx.Tx, ownerID string) error {
	const query = `SELECT active, deleted_at FROM users WHERE id = $1`

	var (
		active    bool
		deletedAt *time.Time
	)

	err := tx.QueryRow(ctx, query, ownerID).Scan(&active, &deletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrOrderOwnerNotFound
	}
	if err != nil {
		return err
	}

	if !active || deletedAt != nil {
		return domain.ErrOrderOwnerUnavailable
	}

	return nil
}

type lockedProduct struct {
	price     float64
	stock     int
	active    bool
	deletedAt *time.Time
}

func (r *PostgresOrderRepository) Create(
	ctx context.Context,
	ownerID string,
	items []domain.CreateOrderItem,
) (*domain.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := r.validateOwner(ctx, tx, ownerID); err != nil {
		return nil, err
	}

	quantities := make(map[string]int, len(items))
	productIDs := make([]string, 0, len(items))

	for _, item := range items {
		quantities[item.ProductID] = item.Quantity
		productIDs = append(productIDs, item.ProductID)
	}

	sort.Strings(productIDs)

	const lockQuery = `
		SELECT id, price, stock, active, deleted_at
		FROM products
		WHERE id = ANY($1)
		ORDER BY id
		FOR UPDATE
	`

	rows, err := tx.Query(ctx, lockQuery, productIDs)
	if err != nil {
		return nil, err
	}

	locked := make(map[string]lockedProduct, len(productIDs))

	for rows.Next() {
		var (
			id      string
			product lockedProduct
		)

		if err := rows.Scan(&id, &product.price, &product.stock, &product.active, &product.deletedAt); err != nil {
			rows.Close()
			return nil, err
		}

		locked[id] = product
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	type resolvedItem struct {
		productID string
		quantity  int
		unitPrice float64
	}

	resolved := make([]resolvedItem, 0, len(productIDs))
	var total float64

	for _, productID := range productIDs {
		product, ok := locked[productID]
		if !ok {
			return nil, domain.ErrProductNotFound
		}

		if !product.active || product.deletedAt != nil {
			return nil, domain.ErrProductUnavailable
		}

		quantity := quantities[productID]
		if product.stock < quantity {
			return nil, domain.ErrInsufficientStock
		}

		resolved = append(resolved, resolvedItem{
			productID: productID,
			quantity:  quantity,
			unitPrice: product.price,
		})

		total += float64(quantity) * product.price
	}

	const insertOrder = `
		INSERT INTO orders (customer_id, status, total_amount)
		VALUES ($1, 'PENDING', $2)
		RETURNING id, customer_id, status, total_amount, paid_at, canceled_at, created_at, updated_at
	`

	order, err := scanOrder(tx.QueryRow(ctx, insertOrder, ownerID, total))
	if err != nil {
		return nil, err
	}

	const insertItem = `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, order_id, product_id, quantity, unit_price, created_at
	`

	const decrementStock = `
		UPDATE products
		SET stock = stock - $2, updated_at = NOW()
		WHERE id = $1
		  AND stock >= $2
	`

	order.Items = make([]domain.OrderItem, 0, len(resolved))

	for _, item := range resolved {
		orderItem, err := scanOrderItem(tx.QueryRow(ctx, insertItem, order.ID, item.productID, item.quantity, item.unitPrice))
		if err != nil {
			return nil, err
		}

		tag, err := tx.Exec(ctx, decrementStock, item.productID, item.quantity)
		if err != nil {
			return nil, err
		}

		if tag.RowsAffected() == 0 {
			return nil, domain.ErrInsufficientStock
		}

		order.Items = append(order.Items, orderItem)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return order, nil
}

// buildOrderFilters segue a mesma convenção de buildProductFilters e
// buildUserFilters: monta a cláusula WHERE e os argumentos correspondentes,
// para serem reutilizados tanto na contagem quanto na consulta paginada.
func buildOrderFilters(filter OrderFilter) (string, []any) {
	if filter.CustomerID == nil {
		return "", nil
	}

	return " WHERE customer_id = $1", []any{*filter.CustomerID}
}

func (r *PostgresOrderRepository) Search(ctx context.Context, filter OrderFilter) (*OrderSearchResult, error) {
	whereClause, args := buildOrderFilters(filter)

	// Contagem direta em orders (nunca sobre um JOIN com order_items, que
	// contaria um pedido uma vez por item).
	countQuery := `
		SELECT COUNT(*)
		FROM orders
	` + whereClause

	var total int64

	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	selectQuery := `
		SELECT id, customer_id, status, total_amount, paid_at, canceled_at, created_at, updated_at
		FROM orders
	` + whereClause + fmt.Sprintf(`
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
		OFFSET $%d
	`, len(args)+1, len(args)+2)

	selectArgs := append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, err
	}

	orders := make([]domain.Order, 0)
	orderIndex := make(map[string]int, len(orders))

	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}

		order.Items = make([]domain.OrderItem, 0)

		orderIndex[order.ID] = len(orders)
		orders = append(orders, *order)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	if len(orders) == 0 {
		return &OrderSearchResult{Orders: orders, Total: total}, nil
	}

	// Segunda consulta, restrita apenas aos pedidos desta página: a
	// paginação (LIMIT/OFFSET) nunca é aplicada sobre linhas de
	// order_items, só sobre orders.
	orderIDs := make([]string, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	const itemsQuery = `
		SELECT id, order_id, product_id, quantity, unit_price, created_at
		FROM order_items
		WHERE order_id = ANY($1)
		ORDER BY order_id, created_at, id
	`

	itemRows, err := r.db.Query(ctx, itemsQuery, orderIDs)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		item, err := scanOrderItem(itemRows)
		if err != nil {
			return nil, err
		}

		if idx, ok := orderIndex[item.OrderID]; ok {
			orders[idx].Items = append(orders[idx].Items, item)
		}
	}

	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	return &OrderSearchResult{Orders: orders, Total: total}, nil
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	const query = `
		SELECT id, customer_id, status, total_amount, paid_at, canceled_at, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order, err := scanOrder(r.db.QueryRow(ctx, query, id))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	items, err := r.findItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (r *PostgresOrderRepository) PayByID(ctx context.Context, id string, ownerID string) (*domain.Order, error) {
	const query = `
		UPDATE orders
		SET status = 'PAID', paid_at = NOW(), updated_at = NOW()
		WHERE id = $1
		  AND customer_id = $2
		  AND status = 'PENDING'
		RETURNING id, customer_id, status, total_amount, paid_at, canceled_at, created_at, updated_at
	`

	order, err := scanOrder(r.db.QueryRow(ctx, query, id, ownerID))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, r.diagnosePayFailure(ctx, id, ownerID)
	}
	if err != nil {
		return nil, err
	}

	items, err := r.findItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

// diagnosePayFailure roda depois que a atualização atômica de PayByID não
// afeta nenhuma linha, só para decidir qual erro (e status HTTP) devolver.
// Não concede nenhuma permissão extra: a autorização real já foi aplicada,
// e falhou, na cláusula WHERE do UPDATE acima.
func (r *PostgresOrderRepository) diagnosePayFailure(ctx context.Context, id string, ownerID string) error {
	const query = `SELECT customer_id, status FROM orders WHERE id = $1`

	var (
		customerID string
		status     string
	)

	err := r.db.QueryRow(ctx, query, id).Scan(&customerID, &status)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrOrderNotFound
	}
	if err != nil {
		return err
	}

	if customerID != ownerID {
		return domain.ErrOrderAccessDenied
	}

	return domain.ErrOrderCannotBePaid
}

func (r *PostgresOrderRepository) CancelByID(
	ctx context.Context,
	id string,
	requesterID string,
	isAdmin bool,
) (*domain.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var (
		customerID string
		status     string
	)

	err = tx.QueryRow(ctx, `SELECT customer_id, status FROM orders WHERE id = $1 FOR UPDATE`, id).
		Scan(&customerID, &status)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	if !isAdmin && customerID != requesterID {
		return nil, domain.ErrOrderAccessDenied
	}

	if status != string(domain.OrderStatusPending) {
		return nil, domain.ErrOrderCannotBeCanceled
	}

	const itemsQuery = `
		SELECT id, order_id, product_id, quantity, unit_price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY product_id
	`

	rows, err := tx.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}

	items := make([]domain.OrderItem, 0)

	for rows.Next() {
		item, err := scanOrderItem(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	const restoreStock = `
		UPDATE products
		SET stock = stock + $2, updated_at = NOW()
		WHERE id = $1
	`

	for _, item := range items {
		if _, err := tx.Exec(ctx, restoreStock, item.ProductID, item.Quantity); err != nil {
			return nil, err
		}
	}

	const updateOrder = `
		UPDATE orders
		SET status = 'CANCELED', canceled_at = NOW(), updated_at = NOW()
		WHERE id = $1
		  AND status = 'PENDING'
		RETURNING id, customer_id, status, total_amount, paid_at, canceled_at, created_at, updated_at
	`

	order, err := scanOrder(tx.QueryRow(ctx, updateOrder, id))
	if err != nil {
		return nil, err
	}
	order.Items = items

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return order, nil
}
