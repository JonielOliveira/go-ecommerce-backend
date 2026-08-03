package domain

import (
	"testing"
	"time"
)

// TestActivatable verifica os dois construtores e as transições de estado
// usados por Product e User (composição, não herança).
func TestActivatable(t *testing.T) {
	t.Run("NewActivatable começa ativo", func(t *testing.T) {
		activatable := NewActivatable()

		if !activatable.IsActive() {
			t.Errorf("IsActive() = false; esperado = true")
		}
	})

	t.Run("Deactivate desativa e Activate reativa", func(t *testing.T) {
		activatable := NewActivatable()

		activatable.Deactivate()
		if activatable.IsActive() {
			t.Fatalf("IsActive() = true após Deactivate; esperado = false")
		}

		activatable.Activate()
		if !activatable.IsActive() {
			t.Errorf("IsActive() = false após Activate; esperado = true")
		}
	})

	t.Run("NewActivatableFrom respeita o valor informado", func(t *testing.T) {
		testCases := []bool{true, false}

		for _, active := range testCases {
			activatable := NewActivatableFrom(active)
			if activatable.IsActive() != active {
				t.Errorf("NewActivatableFrom(%v).IsActive() = %v; esperado = %v", active, activatable.IsActive(), active)
			}
		}
	})
}

// TestSoftDelete verifica os dois construtores e as transições de estado
// usados por Product e User.
func TestSoftDelete(t *testing.T) {
	t.Run("NewSoftDelete começa não removido", func(t *testing.T) {
		softDelete := NewSoftDelete()

		if softDelete.IsDeleted() {
			t.Errorf("IsDeleted() = true; esperado = false")
		}
		if softDelete.DeletedAt() != nil {
			t.Errorf("DeletedAt() = %v; esperado = nil", softDelete.DeletedAt())
		}
	})

	t.Run("Delete marca removido e Restore desfaz", func(t *testing.T) {
		softDelete := NewSoftDelete()

		softDelete.Delete()
		if !softDelete.IsDeleted() || softDelete.DeletedAt() == nil {
			t.Fatalf("estado após Delete = IsDeleted:%v DeletedAt:%v; esperado removido", softDelete.IsDeleted(), softDelete.DeletedAt())
		}

		softDelete.Restore()
		if softDelete.IsDeleted() || softDelete.DeletedAt() != nil {
			t.Errorf("estado após Restore = IsDeleted:%v DeletedAt:%v; esperado não removido", softDelete.IsDeleted(), softDelete.DeletedAt())
		}
	})

	t.Run("NewSoftDeleteFrom respeita o valor informado", func(t *testing.T) {
		softDelete := NewSoftDeleteFrom(nil)
		if softDelete.IsDeleted() {
			t.Errorf("NewSoftDeleteFrom(nil).IsDeleted() = true; esperado = false")
		}

		deletedAt := time.Now()
		softDelete = NewSoftDeleteFrom(&deletedAt)
		if !softDelete.IsDeleted() || !softDelete.DeletedAt().Equal(deletedAt) {
			t.Errorf("NewSoftDeleteFrom(&deletedAt) = IsDeleted:%v DeletedAt:%v; esperado removido em %v", softDelete.IsDeleted(), softDelete.DeletedAt(), deletedAt)
		}
	})
}

// TestTimestamps verifica que createdAt/updatedAt nascem iguais e que Touch
// avança apenas updatedAt.
func TestTimestamps(t *testing.T) {
	t.Run("NewTimestamps define createdAt e updatedAt iguais", func(t *testing.T) {
		timestamps := NewTimestamps()

		if !timestamps.CreatedAt().Equal(timestamps.UpdatedAt()) {
			t.Errorf("createdAt = %v; updatedAt = %v; esperados iguais", timestamps.CreatedAt(), timestamps.UpdatedAt())
		}
	})

	t.Run("Touch atualiza apenas updatedAt", func(t *testing.T) {
		timestamps := NewTimestamps()
		createdAt := timestamps.CreatedAt()

		time.Sleep(time.Millisecond)
		timestamps.Touch()

		if !timestamps.CreatedAt().Equal(createdAt) {
			t.Errorf("createdAt mudou após Touch: %v -> %v", createdAt, timestamps.CreatedAt())
		}
		if !timestamps.UpdatedAt().After(createdAt) {
			t.Errorf("updatedAt = %v; esperado depois de %v", timestamps.UpdatedAt(), createdAt)
		}
	})

	t.Run("NewTimestampsFrom respeita os valores informados", func(t *testing.T) {
		createdAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
		updatedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)

		timestamps := NewTimestampsFrom(createdAt, updatedAt)

		if !timestamps.CreatedAt().Equal(createdAt) || !timestamps.UpdatedAt().Equal(updatedAt) {
			t.Errorf("timestamps = %v/%v; esperado = %v/%v", timestamps.CreatedAt(), timestamps.UpdatedAt(), createdAt, updatedAt)
		}
	})
}

// TestUserAuthenticationIsDeleted verifica a mesma regra de soft delete
// (deletedAt != nil) usada pelo registro de autenticação, que não compõe
// SoftDelete por trafegar campos diferentes (ver AuthService.Login).
func TestUserAuthenticationIsDeleted(t *testing.T) {
	t.Run("deletedAt nil não está removido", func(t *testing.T) {
		auth := &UserAuthentication{DeletedAt: nil}

		if auth.IsDeleted() {
			t.Errorf("IsDeleted() = true; esperado = false")
		}
	})

	t.Run("deletedAt preenchido está removido", func(t *testing.T) {
		deletedAt := time.Now()
		auth := &UserAuthentication{DeletedAt: &deletedAt}

		if !auth.IsDeleted() {
			t.Errorf("IsDeleted() = false; esperado = true")
		}
	})
}
