package domain

import (
	"errors"
	"testing"
	"time"
)

// TestNewUser verifica as regras aplicadas na criação: nome, e-mail e papel
// são validados juntos (errors.Join), e nome/e-mail são normalizados (trim)
// antes de virar usuário.
func TestNewUser(t *testing.T) {
	testCases := []struct {
		name      string
		userName  string
		userEmail string
		role      UserRole
		wantErr   error
	}{
		{
			name:      "cria usuário válido",
			userName:  "Ana Souza",
			userEmail: "ana@example.com",
			role:      RoleCustomer,
		},
		{
			name:      "aceita papel admin",
			userName:  "Ana Souza",
			userEmail: "ana@example.com",
			role:      RoleAdmin,
		},
		{
			name:      "rejeita nome vazio",
			userName:  "   ",
			userEmail: "ana@example.com",
			role:      RoleCustomer,
			wantErr:   ErrInvalidUserName,
		},
		{
			name:      "rejeita email sem arroba",
			userName:  "Ana Souza",
			userEmail: "anaexample.com",
			role:      RoleCustomer,
			wantErr:   ErrInvalidUserEmail,
		},
		{
			name:      "rejeita email sem domínio",
			userName:  "Ana Souza",
			userEmail: "ana@",
			role:      RoleCustomer,
			wantErr:   ErrInvalidUserEmail,
		},
		{
			name:      "rejeita email com espaço",
			userName:  "Ana Souza",
			userEmail: "ana souza@example.com",
			role:      RoleCustomer,
			wantErr:   ErrInvalidUserEmail,
		},
		{
			name:      "rejeita papel desconhecido",
			userName:  "Ana Souza",
			userEmail: "ana@example.com",
			role:      UserRole("superuser"),
			wantErr:   ErrInvalidUserRole,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			user, err := NewUser(testCase.userName, testCase.userEmail, testCase.role, nil)

			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("erro recebido = %v; esperado = %v", err, testCase.wantErr)
			}
			if testCase.wantErr != nil {
				if user != nil {
					t.Errorf("usuário retornado = %#v; esperado = nil", user)
				}
				return
			}
			if user.Role() != testCase.role {
				t.Errorf("papel = %q; esperado = %q", user.Role(), testCase.role)
			}
		})
	}

	t.Run("acumula todos os erros de validação simultaneamente", func(t *testing.T) {
		_, err := NewUser("   ", "email-invalido", UserRole("superuser"), nil)

		if !errors.Is(err, ErrInvalidUserName) {
			t.Errorf("erro = %v; esperado incluir %v", err, ErrInvalidUserName)
		}
		if !errors.Is(err, ErrInvalidUserEmail) {
			t.Errorf("erro = %v; esperado incluir %v", err, ErrInvalidUserEmail)
		}
		if !errors.Is(err, ErrInvalidUserRole) {
			t.Errorf("erro = %v; esperado incluir %v", err, ErrInvalidUserRole)
		}
	})

	t.Run("remove espaços de nome e email", func(t *testing.T) {
		user, err := NewUser("  Ana Souza  ", "  ana@example.com  ", RoleCustomer, nil)

		if err != nil {
			t.Fatalf("NewUser retornou erro inesperado: %v", err)
		}
		if user.Name() != "Ana Souza" {
			t.Errorf("nome = %q; esperado sem espaços = %q", user.Name(), "Ana Souza")
		}
		if user.Email() != "ana@example.com" {
			t.Errorf("email = %q; esperado sem espaços = %q", user.Email(), "ana@example.com")
		}
	})
}

// TestUserUpdate verifica que Update revalida os dados e que uma
// atualização inválida não deixa o usuário em estado parcialmente
// modificado — os dados originais devem ser preservados.
func TestUserUpdate(t *testing.T) {
	t.Run("atualiza campos válidos", func(t *testing.T) {
		user, err := NewUser("Ana Souza", "ana@example.com", RoleCustomer, nil)
		if err != nil {
			t.Fatalf("montar usuário de teste: %v", err)
		}

		err = user.Update("Ana Souza Silva", "ana.silva@example.com", RoleAdmin, nil)

		if err != nil {
			t.Fatalf("Update retornou erro inesperado: %v", err)
		}
		if user.Name() != "Ana Souza Silva" || user.Email() != "ana.silva@example.com" || user.Role() != RoleAdmin {
			t.Errorf("usuário atualizado = %#v", user)
		}
	})

	t.Run("mantém dados originais quando a atualização é inválida", func(t *testing.T) {
		user, err := NewUser("Ana Souza", "ana@example.com", RoleCustomer, nil)
		if err != nil {
			t.Fatalf("montar usuário de teste: %v", err)
		}

		err = user.Update("   ", "email-invalido", UserRole("superuser"), nil)

		if !errors.Is(err, ErrInvalidUserName) {
			t.Fatalf("erro recebido = %v; esperado incluir = %v", err, ErrInvalidUserName)
		}
		if user.Name() != "Ana Souza" || user.Email() != "ana@example.com" || user.Role() != RoleCustomer {
			t.Errorf("usuário deveria permanecer inalterado; ficou = %#v", user)
		}
	})
}

// TestRestoreUser verifica que a reidratação a partir do banco preenche
// todos os campos e continua aplicando a mesma validação usada por NewUser.
func TestRestoreUser(t *testing.T) {
	t.Run("restaura usuário com todos os campos", func(t *testing.T) {
		createdAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
		updatedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
		deletedAt := time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)

		user, err := RestoreUser("user-1", "Ana Souza", "ana@example.com", nil, RoleAdmin, false, nil, nil, createdAt, updatedAt, &deletedAt)

		if err != nil {
			t.Fatalf("RestoreUser retornou erro inesperado: %v", err)
		}
		if user.ID() != "user-1" {
			t.Errorf("id = %q; esperado = %q", user.ID(), "user-1")
		}
		if user.IsActive() {
			t.Errorf("usuário deveria estar inativo")
		}
		if !user.IsDeleted() || user.DeletedAt() == nil || !user.DeletedAt().Equal(deletedAt) {
			t.Errorf("deletedAt = %v; esperado = %v", user.DeletedAt(), deletedAt)
		}
		if !user.CreatedAt().Equal(createdAt) || !user.UpdatedAt().Equal(updatedAt) {
			t.Errorf("timestamps = %v/%v; esperado = %v/%v", user.CreatedAt(), user.UpdatedAt(), createdAt, updatedAt)
		}
	})

	t.Run("rejeita dados inválidos mesmo restaurando", func(t *testing.T) {
		_, err := RestoreUser("user-2", "Ana Souza", "email-invalido", nil, RoleCustomer, true, nil, nil, time.Now(), time.Now(), nil)

		if !errors.Is(err, ErrInvalidUserEmail) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, ErrInvalidUserEmail)
		}
	})
}

// TestUserRoleIsValid verifica que só "customer" e "admin" são papéis
// reconhecidos.
func TestUserRoleIsValid(t *testing.T) {
	testCases := []struct {
		role UserRole
		want bool
	}{
		{role: RoleCustomer, want: true},
		{role: RoleAdmin, want: true},
		{role: UserRole("superuser"), want: false},
		{role: UserRole(""), want: false},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.role), func(t *testing.T) {
			if got := testCase.role.IsValid(); got != testCase.want {
				t.Errorf("IsValid() de %q = %v; esperado = %v", testCase.role, got, testCase.want)
			}
		})
	}
}
