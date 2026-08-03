package domain

import (
	"errors"
	"testing"
	"time"
)

// TestOrderCanPayAndCanCancel verifica que só um pedido PENDING pode ser
// pago ou cancelado.
func TestOrderCanPayAndCanCancel(t *testing.T) {
	testCases := []struct {
		status OrderStatus
		want   bool
	}{
		{status: OrderStatusPending, want: true},
		{status: OrderStatusPaid, want: false},
		{status: OrderStatusCanceled, want: false},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.status), func(t *testing.T) {
			order := &Order{Status: testCase.status}

			if got := order.CanPay(); got != testCase.want {
				t.Errorf("CanPay() com status=%s = %v; esperado = %v", testCase.status, got, testCase.want)
			}
			if got := order.CanCancel(); got != testCase.want {
				t.Errorf("CanCancel() com status=%s = %v; esperado = %v", testCase.status, got, testCase.want)
			}
		})
	}
}

// TestOrderPay verifica que só um pedido PENDING é pago, e que tentar pagar
// um pedido em outro status não o modifica.
func TestOrderPay(t *testing.T) {
	t.Run("paga pedido pendente", func(t *testing.T) {
		order := &Order{Status: OrderStatusPending}
		now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

		err := order.Pay(now)

		if err != nil {
			t.Fatalf("Pay retornou erro inesperado: %v", err)
		}
		if order.Status != OrderStatusPaid {
			t.Errorf("status = %q; esperado = %q", order.Status, OrderStatusPaid)
		}
		if order.PaidAt == nil || !order.PaidAt.Equal(now) {
			t.Errorf("paidAt = %v; esperado = %v", order.PaidAt, now)
		}
		if !order.UpdatedAt.Equal(now) {
			t.Errorf("updatedAt = %v; esperado = %v", order.UpdatedAt, now)
		}
	})

	testCases := []struct {
		name   string
		status OrderStatus
	}{
		{name: "recusa pagar pedido já pago", status: OrderStatusPaid},
		{name: "recusa pagar pedido cancelado", status: OrderStatusCanceled},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			order := &Order{Status: testCase.status}

			err := order.Pay(time.Now())

			if !errors.Is(err, ErrOrderCannotBePaid) {
				t.Fatalf("erro recebido = %v; esperado = %v", err, ErrOrderCannotBePaid)
			}
			if order.Status != testCase.status || order.PaidAt != nil {
				t.Errorf("pedido deveria permanecer inalterado; ficou = %#v", order)
			}
		})
	}
}

// TestOrderCancel espelha TestOrderPay para a operação de cancelamento.
func TestOrderCancel(t *testing.T) {
	t.Run("cancela pedido pendente", func(t *testing.T) {
		order := &Order{Status: OrderStatusPending}
		now := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

		err := order.Cancel(now)

		if err != nil {
			t.Fatalf("Cancel retornou erro inesperado: %v", err)
		}
		if order.Status != OrderStatusCanceled {
			t.Errorf("status = %q; esperado = %q", order.Status, OrderStatusCanceled)
		}
		if order.CanceledAt == nil || !order.CanceledAt.Equal(now) {
			t.Errorf("canceledAt = %v; esperado = %v", order.CanceledAt, now)
		}
		if !order.UpdatedAt.Equal(now) {
			t.Errorf("updatedAt = %v; esperado = %v", order.UpdatedAt, now)
		}
	})

	testCases := []struct {
		name   string
		status OrderStatus
	}{
		{name: "recusa cancelar pedido já pago", status: OrderStatusPaid},
		{name: "recusa cancelar pedido já cancelado", status: OrderStatusCanceled},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			order := &Order{Status: testCase.status}

			err := order.Cancel(time.Now())

			if !errors.Is(err, ErrOrderCannotBeCanceled) {
				t.Fatalf("erro recebido = %v; esperado = %v", err, ErrOrderCannotBeCanceled)
			}
			if order.Status != testCase.status || order.CanceledAt != nil {
				t.Errorf("pedido deveria permanecer inalterado; ficou = %#v", order)
			}
		})
	}
}

// TestOrderItemSubtotal verifica o cálculo simples de quantidade × preço
// unitário, incluindo o caso de quantidade zero.
func TestOrderItemSubtotal(t *testing.T) {
	testCases := []struct {
		name      string
		quantity  int
		unitPrice float64
		want      float64
	}{
		{name: "quantidade zero", quantity: 0, unitPrice: 99.99, want: 0},
		{name: "quantidade positiva", quantity: 3, unitPrice: 10.5, want: 31.5},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			item := OrderItem{Quantity: testCase.quantity, UnitPrice: testCase.unitPrice}

			if got := item.Subtotal(); got != testCase.want {
				t.Errorf("Subtotal() = %v; esperado = %v", got, testCase.want)
			}
		})
	}
}

// TestOrderStatusIsValid verifica que só os três status conhecidos são
// aceitos.
func TestOrderStatusIsValid(t *testing.T) {
	testCases := []struct {
		status OrderStatus
		want   bool
	}{
		{status: OrderStatusPending, want: true},
		{status: OrderStatusPaid, want: true},
		{status: OrderStatusCanceled, want: true},
		{status: OrderStatus("UNKNOWN"), want: false},
		{status: OrderStatus(""), want: false},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.status), func(t *testing.T) {
			if got := testCase.status.IsValid(); got != testCase.want {
				t.Errorf("IsValid() de %q = %v; esperado = %v", testCase.status, got, testCase.want)
			}
		})
	}
}
