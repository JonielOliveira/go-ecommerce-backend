//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ecommerce/internal/domain"
)

// mustCreateUser cria um usuário válido diretamente pelo Repository, sem
// passar por Service/Handler. O hash de senha é uma string qualquer: o
// Repository só grava o que recebe em password_hash (TEXT), não valida
// formato bcrypt.
func mustCreateUser(t *testing.T, repo *PostgresUserRepository, name, email string, role domain.UserRole) *domain.User {
	t.Helper()

	user, err := domain.NewUser(name, email, role, nil)
	if err != nil {
		t.Fatalf("montar usuário de teste %q: %v", email, err)
	}

	created, err := repo.Create(context.Background(), user, "hash-fake-de-teste")
	if err != nil {
		t.Fatalf("criar usuário de teste %q: %v", email, err)
	}

	return created
}

// TestPostgresUserRepositoryCreateAndFindByID usa o PostgreSQL real para
// verificar em conjunto SQL, transação (usuário + credenciais em duas
// tabelas) e mapeamento da entidade.
func TestPostgresUserRepositoryCreateAndFindByID(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresUserRepository(db)

	created := mustCreateUser(t, repo, "Ana Souza", "ana.criar@example.com", domain.RoleCustomer)
	if created.ID() == "" {
		t.Fatalf("id do usuário criado está vazio")
	}
	if created.Role() != domain.RoleCustomer || !created.IsActive() || created.IsDeleted() {
		t.Errorf("usuário criado = %#v", created)
	}

	got, err := repo.FindByID(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("FindByID retornou erro: %v", err)
	}
	if got.ID() != created.ID() || got.Email() != created.Email() {
		t.Errorf("usuário lido = %#v; esperado = %#v", got, created)
	}

	_, err = repo.FindByID(context.Background(), "00000000-0000-7000-8000-000000000000")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("FindByID inexistente = %v; esperado = %v", err, domain.ErrUserNotFound)
	}

	t.Run("email duplicado retorna ErrUserEmailAlreadyExists", func(t *testing.T) {
		duplicate, err := domain.NewUser("Outra Ana", created.Email(), domain.RoleCustomer, nil)
		if err != nil {
			t.Fatalf("montar usuário duplicado: %v", err)
		}

		_, err = repo.Create(context.Background(), duplicate, "hash-fake-de-teste")
		if !errors.Is(err, domain.ErrUserEmailAlreadyExists) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrUserEmailAlreadyExists)
		}
	})
}

// TestPostgresUserRepositoryUpdate verifica os três caminhos de erro que só
// a query real distingue: usuário inexistente, usuário removido, e e-mail
// já usado por outro usuário (a constraint única também é verificada
// dentro do UPDATE, não só do INSERT).
func TestPostgresUserRepositoryUpdate(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresUserRepository(db)

	t.Run("atualiza usuário existente", func(t *testing.T) {
		created := mustCreateUser(t, repo, "Usuário Update", "usuario.update@example.com", domain.RoleCustomer)

		if err := created.Update("Usuário Atualizado", "usuario.update@example.com", domain.RoleAdmin, nil); err != nil {
			t.Fatalf("Update do domínio: %v", err)
		}

		updated, err := repo.Update(context.Background(), created, nil)
		if err != nil {
			t.Fatalf("Repository.Update retornou erro: %v", err)
		}
		if updated.Name() != "Usuário Atualizado" || updated.Role() != domain.RoleAdmin {
			t.Errorf("usuário atualizado = %#v", updated)
		}
	})

	t.Run("usuário inexistente retorna ErrUserNotFound", func(t *testing.T) {
		ghost, err := domain.RestoreUser("00000000-0000-7000-8000-000000000000", "Fantasma", "fantasma.update@example.com", nil, domain.RoleCustomer, true, nil, nil, time.Now(), time.Now(), nil)
		if err != nil {
			t.Fatalf("montar usuário fantasma: %v", err)
		}

		_, err = repo.Update(context.Background(), ghost, nil)
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Fatalf("Update inexistente = %v; esperado = %v", err, domain.ErrUserNotFound)
		}
	})

	t.Run("usuário removido retorna ErrUserAlreadyDeleted", func(t *testing.T) {
		created := mustCreateUser(t, repo, "Usuário Removido Update", "usuario.removido.update@example.com", domain.RoleCustomer)
		if err := repo.DeleteByID(context.Background(), created.ID()); err != nil {
			t.Fatalf("DeleteByID: %v", err)
		}

		if err := created.Update("Novo Nome", "usuario.removido.update@example.com", domain.RoleCustomer, nil); err != nil {
			t.Fatalf("Update do domínio: %v", err)
		}

		_, err := repo.Update(context.Background(), created, nil)
		if !errors.Is(err, domain.ErrUserAlreadyDeleted) {
			t.Fatalf("Update de removido = %v; esperado = %v", err, domain.ErrUserAlreadyDeleted)
		}
	})

	t.Run("email já usado por outro usuário retorna ErrUserEmailAlreadyExists", func(t *testing.T) {
		mustCreateUser(t, repo, "Primeiro", "primeiro.update@example.com", domain.RoleCustomer)
		second := mustCreateUser(t, repo, "Segundo", "segundo.update@example.com", domain.RoleCustomer)

		if err := second.Update("Segundo", "primeiro.update@example.com", domain.RoleCustomer, nil); err != nil {
			t.Fatalf("Update do domínio: %v", err)
		}

		_, err := repo.Update(context.Background(), second, nil)
		if !errors.Is(err, domain.ErrUserEmailAlreadyExists) {
			t.Fatalf("erro = %v; esperado = %v", err, domain.ErrUserEmailAlreadyExists)
		}
	})
}

// TestPostgresUserRepositoryDeleteRestoreActivateDeactivate espelha a
// mesma bateria de transições de estado já usada em Product.
func TestPostgresUserRepositoryDeleteRestoreActivateDeactivate(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresUserRepository(db)

	t.Run("delete e restore", func(t *testing.T) {
		user := mustCreateUser(t, repo, "Usuário Delete", "usuario.delete@example.com", domain.RoleCustomer)

		if err := repo.DeleteByID(context.Background(), user.ID()); err != nil {
			t.Fatalf("DeleteByID: %v", err)
		}
		if err := repo.DeleteByID(context.Background(), user.ID()); !errors.Is(err, domain.ErrUserAlreadyDeleted) {
			t.Fatalf("segundo delete = %v; esperado = %v", err, domain.ErrUserAlreadyDeleted)
		}

		if err := repo.RestoreByID(context.Background(), user.ID()); err != nil {
			t.Fatalf("RestoreByID: %v", err)
		}
		if err := repo.RestoreByID(context.Background(), user.ID()); !errors.Is(err, domain.ErrUserNotDeleted) {
			t.Fatalf("segunda restauração = %v; esperado = %v", err, domain.ErrUserNotDeleted)
		}

		got, err := repo.FindByID(context.Background(), user.ID())
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if got.IsActive() {
			t.Errorf("restaurar não deveria reativar automaticamente; got = %#v", got)
		}
	})

	t.Run("activate e deactivate", func(t *testing.T) {
		user := mustCreateUser(t, repo, "Usuário Activate", "usuario.activate@example.com", domain.RoleCustomer)

		if err := repo.ActivateByID(context.Background(), user.ID()); !errors.Is(err, domain.ErrUserAlreadyActive) {
			t.Fatalf("ativar usuário já ativo = %v; esperado = %v", err, domain.ErrUserAlreadyActive)
		}

		if err := repo.DeactivateByID(context.Background(), user.ID()); err != nil {
			t.Fatalf("DeactivateByID: %v", err)
		}
		if err := repo.DeactivateByID(context.Background(), user.ID()); !errors.Is(err, domain.ErrUserAlreadyInactive) {
			t.Fatalf("desativar de novo = %v; esperado = %v", err, domain.ErrUserAlreadyInactive)
		}

		if err := repo.ActivateByID(context.Background(), user.ID()); err != nil {
			t.Fatalf("ActivateByID: %v", err)
		}
	})

	t.Run("ativar/desativar um usuário removido reporta a remoção", func(t *testing.T) {
		user := mustCreateUser(t, repo, "Usuário Removido", "usuario.removido.estado@example.com", domain.RoleCustomer)
		if err := repo.DeleteByID(context.Background(), user.ID()); err != nil {
			t.Fatalf("DeleteByID: %v", err)
		}

		if err := repo.ActivateByID(context.Background(), user.ID()); !errors.Is(err, domain.ErrUserAlreadyDeleted) {
			t.Fatalf("ativar usuário removido = %v; esperado = %v", err, domain.ErrUserAlreadyDeleted)
		}
		if err := repo.DeactivateByID(context.Background(), user.ID()); !errors.Is(err, domain.ErrUserAlreadyDeleted) {
			t.Fatalf("desativar usuário removido = %v; esperado = %v", err, domain.ErrUserAlreadyDeleted)
		}
	})

	t.Run("qualquer operação num id inexistente retorna ErrUserNotFound", func(t *testing.T) {
		const fakeID = "00000000-0000-7000-8000-000000000000"

		if err := repo.DeleteByID(context.Background(), fakeID); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("DeleteByID inexistente = %v; esperado = %v", err, domain.ErrUserNotFound)
		}
		if err := repo.RestoreByID(context.Background(), fakeID); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("RestoreByID inexistente = %v; esperado = %v", err, domain.ErrUserNotFound)
		}
		if err := repo.ActivateByID(context.Background(), fakeID); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("ActivateByID inexistente = %v; esperado = %v", err, domain.ErrUserNotFound)
		}
		if err := repo.DeactivateByID(context.Background(), fakeID); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("DeactivateByID inexistente = %v; esperado = %v", err, domain.ErrUserNotFound)
		}
	})
}

// TestPostgresUserRepositorySearch prova que buildUserFilters monta o WHERE
// certo para nome, e-mail, papel, status de ativação e estado de exclusão
// — inclusive combinados — contra dados reais.
func TestPostgresUserRepositorySearch(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresUserRepository(db)

	ana := mustCreateUser(t, repo, "Ana Souza", "ana.busca.repo@example.com", domain.RoleCustomer)
	bruno := mustCreateUser(t, repo, "Bruno Costa", "bruno.busca.repo@example.com", domain.RoleAdmin)
	carla := mustCreateUser(t, repo, "Carla Dias", "carla.busca.repo@example.com", domain.RoleCustomer)
	anaPereira := mustCreateUser(t, repo, "Ana Pereira", "ana.pereira.busca.repo@example.com", domain.RoleCustomer)

	if err := repo.DeactivateByID(context.Background(), carla.ID()); err != nil {
		t.Fatalf("desativar carla: %v", err)
	}
	if err := repo.DeleteByID(context.Background(), anaPereira.ID()); err != nil {
		t.Fatalf("remover ana pereira: %v", err)
	}

	testCases := []struct {
		name    string
		filter  UserSearchFilter
		wantIDs []string
	}{
		{
			name:    "sem filtro extra usa o padrão (só não removidos)",
			filter:  UserSearchFilter{Limit: 10},
			wantIDs: []string{ana.ID(), bruno.ID(), carla.ID()},
		},
		{
			name:    "filtra por nome (ILIKE parcial e case-insensitive)",
			filter:  UserSearchFilter{Name: "ana", Limit: 10},
			wantIDs: []string{ana.ID()}, // "Ana Pereira" também bate no nome, mas está removida
		},
		{
			name:    "filtra por nome com deletionFilter=all",
			filter:  UserSearchFilter{Name: "ana", DeletionFilter: DeletionFilterAll, Limit: 10},
			wantIDs: []string{ana.ID(), anaPereira.ID()},
		},
		{
			name:    "filtra por email",
			filter:  UserSearchFilter{Email: "bruno.busca", Limit: 10},
			wantIDs: []string{bruno.ID()},
		},
		{
			name:    "filtra por papel",
			filter:  UserSearchFilter{Role: "admin", Limit: 10},
			wantIDs: []string{bruno.ID()},
		},
		{
			name:    "filtra por ativo=true",
			filter:  UserSearchFilter{Active: boolPtr(true), Limit: 10},
			wantIDs: []string{ana.ID(), bruno.ID()},
		},
		{
			name:    "filtra por ativo=false",
			filter:  UserSearchFilter{Active: boolPtr(false), Limit: 10},
			wantIDs: []string{carla.ID()},
		},
		{
			name:    "deletionFilter=deleted retorna só os removidos",
			filter:  UserSearchFilter{DeletionFilter: DeletionFilterDeleted, Limit: 10},
			wantIDs: []string{anaPereira.ID()},
		},
		{
			name:    "deletionFilter=all ignora o filtro de remoção",
			filter:  UserSearchFilter{DeletionFilter: DeletionFilterAll, Limit: 10},
			wantIDs: []string{ana.ID(), bruno.ID(), carla.ID(), anaPereira.ID()},
		},
		{
			name:    "combina papel e ativo",
			filter:  UserSearchFilter{Role: "customer", Active: boolPtr(true), Limit: 10},
			wantIDs: []string{ana.ID()},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := repo.Search(context.Background(), testCase.filter)
			if err != nil {
				t.Fatalf("Search retornou erro: %v", err)
			}

			gotIDs := make([]string, 0, len(result.Users))
			for _, user := range result.Users {
				gotIDs = append(gotIDs, user.ID())
			}

			if !sameSet(gotIDs, testCase.wantIDs) {
				t.Errorf("usuários = %v; esperado (sem ordem) = %v", gotIDs, testCase.wantIDs)
			}
			if int(result.Total) != len(testCase.wantIDs) {
				t.Errorf("total = %d; esperado = %d", result.Total, len(testCase.wantIDs))
			}
		})
	}
}

// TestPostgresUserRepositorySearchPagination espelha a mesma verificação
// de Limit/Offset já usada em Product.
func TestPostgresUserRepositorySearchPagination(t *testing.T) {
	db := openTestDatabase(t)
	repo := NewPostgresUserRepository(db)

	const totalUsers = 5
	for i := range totalUsers {
		mustCreateUser(t, repo, fmt.Sprintf("Usuário Paginação %02d", i), fmt.Sprintf("usuario.paginacao.%02d@example.com", i), domain.RoleCustomer)
	}

	seen := make(map[string]bool, totalUsers)
	const pageSize = 2

	for offset := 0; offset < totalUsers; offset += pageSize {
		page, err := repo.Search(context.Background(), UserSearchFilter{Limit: pageSize, Offset: offset})
		if err != nil {
			t.Fatalf("Search(offset=%d) retornou erro: %v", offset, err)
		}
		if int(page.Total) != totalUsers {
			t.Fatalf("total na página offset=%d = %d; esperado = %d", offset, page.Total, totalUsers)
		}

		for _, user := range page.Users {
			if seen[user.ID()] {
				t.Errorf("usuário %s apareceu em mais de uma página", user.ID())
			}
			seen[user.ID()] = true
		}
	}

	if len(seen) != totalUsers {
		t.Errorf("usuários únicos vistos = %d; esperado = %d", len(seen), totalUsers)
	}
}
