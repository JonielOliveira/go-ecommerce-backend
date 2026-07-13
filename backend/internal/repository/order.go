package repository

import (
	"context"

	"ecommerce/internal/domain"
)

// OrderFilter segue a mesma convenção de filtro dos demais repositories
// (ProductSearchFilter, UserSearchFilter): fica em "repository", não em
// "domain", já que descreve uma consulta, não uma regra de negócio.
type OrderFilter struct {
	CustomerID *string
	Limit      int
	Offset     int
}

// OrderSearchResult segue a mesma convenção de ProductSearchResult e
// UserSearchResult: os itens da página atual mais o total de registros que
// batem com o filtro (para calcular totalPages).
type OrderSearchResult struct {
	Orders []domain.Order
	Total  int64
}

type OrderRepository interface {
	// Create consolida a criação do pedido: valida o proprietário, bloqueia
	// os produtos envolvidos, verifica disponibilidade/estoque, grava o
	// pedido e os itens e reduz o estoque — tudo em uma única transação.
	Create(
		ctx context.Context,
		ownerID string,
		items []domain.CreateOrderItem,
	) (*domain.Order, error)

	Search(
		ctx context.Context,
		filter OrderFilter,
	) (*OrderSearchResult, error)

	FindByID(
		ctx context.Context,
		id string,
	) (*domain.Order, error)

	// PayByID exige atomicamente id + ownerID + status PENDING na mesma
	// instrução SQL, sem exceção para admin.
	PayByID(
		ctx context.Context,
		id string,
		ownerID string,
	) (*domain.Order, error)

	// CancelByID aplica a autorização (customer só o próprio; admin
	// qualquer um) dentro da mesma transação/lock do pedido, no mesmo
	// espírito de PayByID.
	CancelByID(
		ctx context.Context,
		id string,
		requesterID string,
		isAdmin bool,
	) (*domain.Order, error)
}
