package domain

import (
	"errors"
	"testing"
	"time"
)

// TestNewProduct verifica as regras aplicadas na criação: nome, preço e
// estoque são validados juntos (errors.Join), e nome/descrição são
// normalizados (trim) antes de virar produto.
func TestNewProduct(t *testing.T) {
	testCases := []struct {
		name        string
		productName string
		description string
		price       float64
		stock       int
		wantErr     error
	}{
		{
			name:        "cria produto válido",
			productName: "Teclado Mecânico",
			description: "RGB, switches azuis",
			price:       199.90,
			stock:       10,
		},
		{
			name:        "aceita estoque zero",
			productName: "Teclado Mecânico",
			description: "RGB, switches azuis",
			price:       199.90,
			stock:       0,
		},
		{
			name:        "rejeita nome vazio",
			productName: "   ",
			description: "RGB, switches azuis",
			price:       199.90,
			stock:       10,
			wantErr:     ErrInvalidProductName,
		},
		{
			name:        "rejeita preço zero",
			productName: "Teclado Mecânico",
			description: "RGB, switches azuis",
			price:       0,
			stock:       10,
			wantErr:     ErrInvalidProductPrice,
		},
		{
			name:        "rejeita preço negativo",
			productName: "Teclado Mecânico",
			description: "RGB, switches azuis",
			price:       -5,
			stock:       10,
			wantErr:     ErrInvalidProductPrice,
		},
		{
			name:        "rejeita estoque negativo",
			productName: "Teclado Mecânico",
			description: "RGB, switches azuis",
			price:       199.90,
			stock:       -1,
			wantErr:     ErrInvalidProductStock,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			product, err := NewProduct(testCase.productName, testCase.description, testCase.price, testCase.stock, nil, nil)

			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("erro recebido = %v; esperado = %v", err, testCase.wantErr)
			}
			if testCase.wantErr != nil {
				if product != nil {
					t.Errorf("produto retornado = %#v; esperado = nil", product)
				}
				return
			}
			if product.Price() != testCase.price || product.Stock() != testCase.stock {
				t.Errorf("produto criado = %#v", product)
			}
		})
	}

	t.Run("acumula todos os erros de validação simultaneamente", func(t *testing.T) {
		_, err := NewProduct("   ", "desc", -5, -1, nil, nil)

		if !errors.Is(err, ErrInvalidProductName) {
			t.Errorf("erro = %v; esperado incluir %v", err, ErrInvalidProductName)
		}
		if !errors.Is(err, ErrInvalidProductPrice) {
			t.Errorf("erro = %v; esperado incluir %v", err, ErrInvalidProductPrice)
		}
		if !errors.Is(err, ErrInvalidProductStock) {
			t.Errorf("erro = %v; esperado incluir %v", err, ErrInvalidProductStock)
		}
	})

	t.Run("remove espaços de nome e descrição", func(t *testing.T) {
		product, err := NewProduct("  Teclado Mecânico  ", "  RGB  ", 10, 1, nil, nil)

		if err != nil {
			t.Fatalf("NewProduct retornou erro inesperado: %v", err)
		}
		if product.Name() != "Teclado Mecânico" {
			t.Errorf("nome = %q; esperado sem espaços = %q", product.Name(), "Teclado Mecânico")
		}
		if product.Description() != "RGB" {
			t.Errorf("descrição = %q; esperado sem espaços = %q", product.Description(), "RGB")
		}
	})
}

// TestProductUpdate verifica que Update revalida os dados e, principalmente,
// que uma atualização inválida não deixa o produto em estado parcialmente
// modificado — os dados originais devem ser preservados.
func TestProductUpdate(t *testing.T) {
	t.Run("atualiza campos válidos", func(t *testing.T) {
		product, err := NewProduct("Teclado Mecânico", "RGB", 199.90, 10, nil, nil)
		if err != nil {
			t.Fatalf("montar produto de teste: %v", err)
		}

		err = product.Update("Mouse Gamer", "Sensor óptico", 149.90, 5, nil, nil)

		if err != nil {
			t.Fatalf("Update retornou erro inesperado: %v", err)
		}
		if product.Name() != "Mouse Gamer" || product.Price() != 149.90 || product.Stock() != 5 {
			t.Errorf("produto atualizado = %#v", product)
		}
	})

	t.Run("mantém dados originais quando a atualização é inválida", func(t *testing.T) {
		product, err := NewProduct("Teclado Mecânico", "RGB", 199.90, 10, nil, nil)
		if err != nil {
			t.Fatalf("montar produto de teste: %v", err)
		}

		err = product.Update("   ", "Nova descrição", -10, -1, nil, nil)

		if !errors.Is(err, ErrInvalidProductName) {
			t.Fatalf("erro recebido = %v; esperado incluir = %v", err, ErrInvalidProductName)
		}
		if product.Name() != "Teclado Mecânico" || product.Price() != 199.90 || product.Stock() != 10 {
			t.Errorf("produto deveria permanecer inalterado; ficou = %#v", product)
		}
	})
}

// TestProductHasStock verifica a regra de disponibilidade: quantidade
// precisa ser positiva e não pode exceder o estoque atual.
func TestProductHasStock(t *testing.T) {
	testCases := []struct {
		name     string
		stock    int
		quantity int
		want     bool
	}{
		{name: "quantidade zero nunca tem estoque", stock: 10, quantity: 0, want: false},
		{name: "quantidade negativa nunca tem estoque", stock: 10, quantity: -1, want: false},
		{name: "quantidade menor que o estoque", stock: 10, quantity: 5, want: true},
		{name: "quantidade igual ao estoque", stock: 10, quantity: 10, want: true},
		{name: "quantidade maior que o estoque", stock: 10, quantity: 11, want: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			product, err := NewProduct("Teclado Mecânico", "RGB", 199.90, testCase.stock, nil, nil)
			if err != nil {
				t.Fatalf("montar produto de teste: %v", err)
			}

			if got := product.HasStock(testCase.quantity); got != testCase.want {
				t.Errorf("HasStock(%d) com estoque=%d = %v; esperado = %v", testCase.quantity, testCase.stock, got, testCase.want)
			}
		})
	}
}

// TestProductReduceStock verifica que quantidade inválida e estoque
// insuficiente são recusados sem alterar o estoque, e que a redução válida
// subtrai corretamente.
func TestProductReduceStock(t *testing.T) {
	t.Run("quantidade inválida não reduz o estoque", func(t *testing.T) {
		product, err := NewProduct("Teclado Mecânico", "RGB", 199.90, 10, nil, nil)
		if err != nil {
			t.Fatalf("montar produto de teste: %v", err)
		}

		err = product.ReduceStock(0)

		if !errors.Is(err, ErrInvalidQuantity) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, ErrInvalidQuantity)
		}
		if product.Stock() != 10 {
			t.Errorf("estoque = %d; esperado permanecer = 10", product.Stock())
		}
	})

	t.Run("estoque insuficiente não reduz o estoque", func(t *testing.T) {
		product, err := NewProduct("Teclado Mecânico", "RGB", 199.90, 5, nil, nil)
		if err != nil {
			t.Fatalf("montar produto de teste: %v", err)
		}

		err = product.ReduceStock(10)

		if !errors.Is(err, ErrInsufficientStock) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, ErrInsufficientStock)
		}
		if product.Stock() != 5 {
			t.Errorf("estoque = %d; esperado permanecer = 5", product.Stock())
		}
	})

	t.Run("reduz o estoque disponível", func(t *testing.T) {
		product, err := NewProduct("Teclado Mecânico", "RGB", 199.90, 10, nil, nil)
		if err != nil {
			t.Fatalf("montar produto de teste: %v", err)
		}

		if err := product.ReduceStock(4); err != nil {
			t.Fatalf("ReduceStock retornou erro inesperado: %v", err)
		}
		if product.Stock() != 6 {
			t.Errorf("estoque = %d; esperado = 6", product.Stock())
		}
	})
}

// TestProductRestoreStock verifica que quantidade inválida é recusada sem
// alterar o estoque, e que a reposição válida soma corretamente.
func TestProductRestoreStock(t *testing.T) {
	t.Run("quantidade inválida não repõe o estoque", func(t *testing.T) {
		product, err := NewProduct("Teclado Mecânico", "RGB", 199.90, 10, nil, nil)
		if err != nil {
			t.Fatalf("montar produto de teste: %v", err)
		}

		err = product.RestoreStock(-1)

		if !errors.Is(err, ErrInvalidQuantity) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, ErrInvalidQuantity)
		}
		if product.Stock() != 10 {
			t.Errorf("estoque = %d; esperado permanecer = 10", product.Stock())
		}
	})

	t.Run("repõe o estoque", func(t *testing.T) {
		product, err := NewProduct("Teclado Mecânico", "RGB", 199.90, 5, nil, nil)
		if err != nil {
			t.Fatalf("montar produto de teste: %v", err)
		}

		if err := product.RestoreStock(3); err != nil {
			t.Fatalf("RestoreStock retornou erro inesperado: %v", err)
		}
		if product.Stock() != 8 {
			t.Errorf("estoque = %d; esperado = 8", product.Stock())
		}
	})
}

// TestRestoreProduct verifica que a reidratação a partir do banco preenche
// todos os campos (id, ativação, remoção) e continua aplicando a mesma
// validação usada por NewProduct.
func TestRestoreProduct(t *testing.T) {
	t.Run("restaura produto com todos os campos", func(t *testing.T) {
		createdAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
		updatedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
		deletedAt := time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)

		product, err := RestoreProduct("prod-1", "Teclado Mecânico", "RGB", 199.90, 10, nil, nil, false, createdAt, updatedAt, &deletedAt)

		if err != nil {
			t.Fatalf("RestoreProduct retornou erro inesperado: %v", err)
		}
		if product.ID() != "prod-1" {
			t.Errorf("id = %q; esperado = %q", product.ID(), "prod-1")
		}
		if product.IsActive() {
			t.Errorf("produto deveria estar inativo")
		}
		if !product.IsDeleted() || product.DeletedAt() == nil || !product.DeletedAt().Equal(deletedAt) {
			t.Errorf("deletedAt = %v; esperado = %v", product.DeletedAt(), deletedAt)
		}
		if !product.CreatedAt().Equal(createdAt) || !product.UpdatedAt().Equal(updatedAt) {
			t.Errorf("timestamps = %v/%v; esperado = %v/%v", product.CreatedAt(), product.UpdatedAt(), createdAt, updatedAt)
		}
	})

	t.Run("rejeita dados inválidos mesmo restaurando", func(t *testing.T) {
		_, err := RestoreProduct("prod-2", "   ", "RGB", 199.90, 10, nil, nil, true, time.Now(), time.Now(), nil)

		if !errors.Is(err, ErrInvalidProductName) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, ErrInvalidProductName)
		}
	})
}
