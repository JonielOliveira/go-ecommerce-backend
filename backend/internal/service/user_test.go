package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// fakeUserRepository substitui o PostgreSQL e permite controlar cada
// operação no teste. Os métodos abaixo fazem este tipo satisfazer a
// interface repository.UserRepository.
type fakeUserRepository struct {
	createFunc         func(user *domain.User, passwordHash string) (*domain.User, error)
	updateFunc         func(user *domain.User, passwordHash *string) (*domain.User, error)
	findByIDFunc       func(id string) (*domain.User, error)
	searchFunc         func(filter repository.UserSearchFilter) (*repository.UserSearchResult, error)
	deleteByIDFunc     func(id string) error
	restoreByIDFunc    func(id string) error
	activateByIDFunc   func(id string) error
	deactivateByIDFunc func(id string) error
}

func (fake *fakeUserRepository) Create(_ context.Context, user *domain.User, passwordHash string) (*domain.User, error) {
	return fake.createFunc(user, passwordHash)
}

func (fake *fakeUserRepository) Update(_ context.Context, user *domain.User, passwordHash *string) (*domain.User, error) {
	return fake.updateFunc(user, passwordHash)
}

func (fake *fakeUserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	return fake.findByIDFunc(id)
}

func (fake *fakeUserRepository) Search(_ context.Context, filter repository.UserSearchFilter) (*repository.UserSearchResult, error) {
	return fake.searchFunc(filter)
}

func (fake *fakeUserRepository) DeleteByID(_ context.Context, id string) error {
	return fake.deleteByIDFunc(id)
}

func (fake *fakeUserRepository) RestoreByID(_ context.Context, id string) error {
	return fake.restoreByIDFunc(id)
}

func (fake *fakeUserRepository) ActivateByID(_ context.Context, id string) error {
	return fake.activateByIDFunc(id)
}

func (fake *fakeUserRepository) DeactivateByID(_ context.Context, id string) error {
	return fake.deactivateByIDFunc(id)
}

// mustNewUser monta um usuário ativo e não removido, com o id e o papel
// informados, pronto para ser devolvido pelos fakes de Repository.
func mustNewUser(t *testing.T, id string, role domain.UserRole) *domain.User {
	t.Helper()

	user, err := domain.RestoreUser(
		id,
		"Ana Souza",
		"ana@example.com",
		nil,
		role,
		true,
		nil,
		nil,
		time.Now(),
		time.Now(),
		nil,
	)
	if err != nil {
		t.Fatalf("montar usuário de teste: %v", err)
	}

	return user
}

// mustDeletedUser monta um usuário já removido (soft delete), usado para
// verificar que a Service recusa atualizações sobre ele.
func mustDeletedUser(t *testing.T, id string) *domain.User {
	t.Helper()

	deletedAt := time.Now()
	user, err := domain.RestoreUser(
		id,
		"Ana Souza",
		"ana@example.com",
		nil,
		domain.RoleCustomer,
		true,
		nil,
		nil,
		time.Now(),
		time.Now(),
		&deletedAt,
	)
	if err != nil {
		t.Fatalf("montar usuário removido de teste: %v", err)
	}

	return user
}

// TestUserServiceRegister verifica o autocadastro público: o papel é sempre
// "customer", a senha chega ao Repository já em hash (nunca em texto puro),
// e dados inválidos nunca chegam ao Repository.
func TestUserServiceRegister(t *testing.T) {
	t.Run("cria cliente com senha em hash", func(t *testing.T) {
		var gotUser *domain.User
		var gotPasswordHash string
		fakeRepository := &fakeUserRepository{
			createFunc: func(user *domain.User, passwordHash string) (*domain.User, error) {
				gotUser = user
				gotPasswordHash = passwordHash
				return mustNewUser(t, "user-1", domain.RoleCustomer), nil
			},
		}
		userService := NewUserService(fakeRepository)

		response, err := userService.Register(context.Background(), dto.RegisterRequest{
			Name:     "Ana Souza",
			Email:    "ana@example.com",
			Password: "segredo123",
		})

		if err != nil {
			t.Fatalf("Register retornou erro inesperado: %v", err)
		}
		if gotUser.Role() != domain.RoleCustomer {
			t.Errorf("papel enviado ao Repository = %q; esperado = %q", gotUser.Role(), domain.RoleCustomer)
		}
		if gotPasswordHash == "segredo123" {
			t.Fatalf("senha enviada ao Repository em texto puro")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(gotPasswordHash), []byte("segredo123")); err != nil {
			t.Errorf("hash da senha não confere com a senha original: %v", err)
		}
		if response.ID != "user-1" {
			t.Errorf("usuário criado = %#v", response)
		}
	})

	t.Run("rejeita dados inválidos sem persistir", func(t *testing.T) {
		repositoryCalls := 0
		fakeRepository := &fakeUserRepository{
			createFunc: func(*domain.User, string) (*domain.User, error) {
				repositoryCalls++
				return nil, nil
			},
		}
		userService := NewUserService(fakeRepository)

		_, err := userService.Register(context.Background(), dto.RegisterRequest{
			Name:     "",
			Email:    "ana@example.com",
			Password: "segredo123",
		})

		if !errors.Is(err, domain.ErrInvalidUserName) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrInvalidUserName)
		}
		if repositoryCalls != 0 {
			t.Errorf("chamadas ao Repository = %d; esperado = 0", repositoryCalls)
		}
	})
}

// TestUserServiceCreate verifica a criação administrativa: papel omitido usa
// "customer" como padrão, papel explícito é respeitado, e papel inválido é
// recusado sem persistir.
func TestUserServiceCreate(t *testing.T) {
	t.Run("usa customer como papel padrão quando omitido", func(t *testing.T) {
		var gotUser *domain.User
		fakeRepository := &fakeUserRepository{
			createFunc: func(user *domain.User, _ string) (*domain.User, error) {
				gotUser = user
				return mustNewUser(t, "user-1", domain.RoleCustomer), nil
			},
		}
		userService := NewUserService(fakeRepository)

		_, err := userService.Create(context.Background(), dto.CreateUserRequest{
			Name:     "Ana Souza",
			Email:    "ana@example.com",
			Password: "segredo123",
		})

		if err != nil {
			t.Fatalf("Create retornou erro inesperado: %v", err)
		}
		if gotUser.Role() != domain.RoleCustomer {
			t.Errorf("papel enviado ao Repository = %q; esperado = %q", gotUser.Role(), domain.RoleCustomer)
		}
	})

	t.Run("respeita papel explícito", func(t *testing.T) {
		var gotUser *domain.User
		fakeRepository := &fakeUserRepository{
			createFunc: func(user *domain.User, _ string) (*domain.User, error) {
				gotUser = user
				return mustNewUser(t, "user-2", domain.RoleAdmin), nil
			},
		}
		userService := NewUserService(fakeRepository)

		role := string(domain.RoleAdmin)
		_, err := userService.Create(context.Background(), dto.CreateUserRequest{
			Name:     "Beto Lima",
			Email:    "beto@example.com",
			Password: "segredo123",
			Role:     &role,
		})

		if err != nil {
			t.Fatalf("Create retornou erro inesperado: %v", err)
		}
		if gotUser.Role() != domain.RoleAdmin {
			t.Errorf("papel enviado ao Repository = %q; esperado = %q", gotUser.Role(), domain.RoleAdmin)
		}
	})

	t.Run("rejeita papel inválido sem persistir", func(t *testing.T) {
		repositoryCalls := 0
		fakeRepository := &fakeUserRepository{
			createFunc: func(*domain.User, string) (*domain.User, error) {
				repositoryCalls++
				return nil, nil
			},
		}
		userService := NewUserService(fakeRepository)

		invalidRole := "superuser"
		_, err := userService.Create(context.Background(), dto.CreateUserRequest{
			Name:     "Beto Lima",
			Email:    "beto@example.com",
			Password: "segredo123",
			Role:     &invalidRole,
		})

		if !errors.Is(err, domain.ErrInvalidUserRole) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrInvalidUserRole)
		}
		if repositoryCalls != 0 {
			t.Errorf("chamadas ao Repository = %d; esperado = 0", repositoryCalls)
		}
	})
}

// TestUserServiceUpdate verifica as regras aplicadas antes de persistir:
// usuário removido é recusado, dados inválidos nunca chegam ao Repository, e
// a senha só é re-hasheada quando informada.
func TestUserServiceUpdate(t *testing.T) {
	t.Run("atualiza usuário existente sem trocar senha", func(t *testing.T) {
		existing := mustNewUser(t, "user-1", domain.RoleCustomer)
		var gotPasswordHash *string
		updateCalls := 0
		fakeRepository := &fakeUserRepository{
			findByIDFunc: func(id string) (*domain.User, error) {
				if id != "user-1" {
					t.Fatalf("id recebido = %s; esperado = user-1", id)
				}
				return existing, nil
			},
			updateFunc: func(user *domain.User, passwordHash *string) (*domain.User, error) {
				updateCalls++
				gotPasswordHash = passwordHash
				return user, nil
			},
		}
		userService := NewUserService(fakeRepository)

		response, err := userService.Update(context.Background(), "user-1", dto.UserUpdateRequest{
			Name:  "Ana Souza Silva",
			Email: "ana@example.com",
			Role:  string(domain.RoleCustomer),
		})

		if err != nil {
			t.Fatalf("Update retornou erro inesperado: %v", err)
		}
		if updateCalls != 1 {
			t.Errorf("chamadas ao Repository.Update = %d; esperado = 1", updateCalls)
		}
		if gotPasswordHash != nil {
			t.Errorf("passwordHash enviado = %v; esperado = nil (senha não informada)", *gotPasswordHash)
		}
		if response.Name != "Ana Souza Silva" {
			t.Errorf("usuário atualizado = %#v", response)
		}
	})

	t.Run("re-hasheia a senha quando informada", func(t *testing.T) {
		existing := mustNewUser(t, "user-1", domain.RoleCustomer)
		var gotPasswordHash *string
		fakeRepository := &fakeUserRepository{
			findByIDFunc: func(string) (*domain.User, error) { return existing, nil },
			updateFunc: func(user *domain.User, passwordHash *string) (*domain.User, error) {
				gotPasswordHash = passwordHash
				return user, nil
			},
		}
		userService := NewUserService(fakeRepository)

		_, err := userService.Update(context.Background(), "user-1", dto.UserUpdateRequest{
			Name:     "Ana Souza",
			Email:    "ana@example.com",
			Role:     string(domain.RoleCustomer),
			Password: "novaSenha123",
		})

		if err != nil {
			t.Fatalf("Update retornou erro inesperado: %v", err)
		}
		if gotPasswordHash == nil {
			t.Fatalf("passwordHash enviado = nil; esperado hash não nulo")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*gotPasswordHash), []byte("novaSenha123")); err != nil {
			t.Errorf("hash da nova senha não confere: %v", err)
		}
	})

	t.Run("recusa atualizar usuário removido", func(t *testing.T) {
		deleted := mustDeletedUser(t, "user-2")
		updateCalls := 0
		fakeRepository := &fakeUserRepository{
			findByIDFunc: func(string) (*domain.User, error) { return deleted, nil },
			updateFunc: func(user *domain.User, _ *string) (*domain.User, error) {
				updateCalls++
				return user, nil
			},
		}
		userService := NewUserService(fakeRepository)

		_, err := userService.Update(context.Background(), "user-2", dto.UserUpdateRequest{
			Name:  "Novo nome",
			Email: "ana@example.com",
			Role:  string(domain.RoleCustomer),
		})

		if !errors.Is(err, domain.ErrUserAlreadyDeleted) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrUserAlreadyDeleted)
		}
		if updateCalls != 0 {
			t.Errorf("chamadas ao Repository.Update = %d; esperado = 0", updateCalls)
		}
	})

	t.Run("recusa dados inválidos sem persistir", func(t *testing.T) {
		existing := mustNewUser(t, "user-3", domain.RoleCustomer)
		updateCalls := 0
		fakeRepository := &fakeUserRepository{
			findByIDFunc: func(string) (*domain.User, error) { return existing, nil },
			updateFunc: func(user *domain.User, _ *string) (*domain.User, error) {
				updateCalls++
				return user, nil
			},
		}
		userService := NewUserService(fakeRepository)

		_, err := userService.Update(context.Background(), "user-3", dto.UserUpdateRequest{
			Name:  "Ana Souza",
			Email: "email-invalido",
			Role:  string(domain.RoleCustomer),
		})

		if !errors.Is(err, domain.ErrInvalidUserEmail) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrInvalidUserEmail)
		}
		if updateCalls != 0 {
			t.Errorf("chamadas ao Repository.Update = %d; esperado = 0", updateCalls)
		}
	})

	t.Run("propaga erro quando usuário não existe", func(t *testing.T) {
		fakeRepository := &fakeUserRepository{
			findByIDFunc: func(string) (*domain.User, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		userService := NewUserService(fakeRepository)

		_, err := userService.Update(context.Background(), "user-404", dto.UserUpdateRequest{
			Name:  "Ana Souza",
			Email: "ana@example.com",
			Role:  string(domain.RoleCustomer),
		})

		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Fatalf("erro recebido = %v; esperado = %v", err, domain.ErrUserNotFound)
		}
	})
}

// TestUserServiceSearch verifica os valores padrão de paginação, o limite
// máximo de PageSize e o cálculo de TotalPages a partir do total devolvido
// pelo Repository.
func TestUserServiceSearch(t *testing.T) {
	testCases := []struct {
		name           string
		request        dto.UserSearchRequest
		total          int64
		wantPage       int
		wantPageSize   int
		wantTotalPages int
	}{
		{
			name:           "aplica página e tamanho padrão quando ausentes",
			request:        dto.UserSearchRequest{},
			total:          25,
			wantPage:       1,
			wantPageSize:   20,
			wantTotalPages: 2,
		},
		{
			name:           "limita tamanho de página a 100",
			request:        dto.UserSearchRequest{Page: 2, PageSize: 500},
			total:          150,
			wantPage:       2,
			wantPageSize:   100,
			wantTotalPages: 2,
		},
		{
			name:           "sem resultados retorna zero páginas",
			request:        dto.UserSearchRequest{Page: 1, PageSize: 10},
			total:          0,
			wantPage:       1,
			wantPageSize:   10,
			wantTotalPages: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var gotFilter repository.UserSearchFilter
			fakeRepository := &fakeUserRepository{
				searchFunc: func(filter repository.UserSearchFilter) (*repository.UserSearchResult, error) {
					gotFilter = filter
					return &repository.UserSearchResult{Users: nil, Total: testCase.total}, nil
				},
			}
			userService := NewUserService(fakeRepository)

			response, err := userService.Search(context.Background(), testCase.request)

			if err != nil {
				t.Fatalf("Search retornou erro inesperado: %v", err)
			}
			if response.Page != testCase.wantPage || response.PageSize != testCase.wantPageSize {
				t.Errorf("paginação = %#v; esperado page=%d pageSize=%d", response, testCase.wantPage, testCase.wantPageSize)
			}
			if response.TotalPages != testCase.wantTotalPages {
				t.Errorf("totalPages = %d; esperado = %d", response.TotalPages, testCase.wantTotalPages)
			}
			wantOffset := (testCase.wantPage - 1) * testCase.wantPageSize
			if gotFilter.Limit != testCase.wantPageSize || gotFilter.Offset != wantOffset {
				t.Errorf("filtro enviado ao Repository = %#v; esperado limit=%d offset=%d", gotFilter, testCase.wantPageSize, wantOffset)
			}
		})
	}
}

// TestMapUserDeletionFilter verifica a tradução do estado de exclusão do DTO
// para o filtro usado pelo Repository, inclusive o valor padrão.
func TestMapUserDeletionFilter(t *testing.T) {
	testCases := []struct {
		name  string
		state dto.DeletionState
		want  repository.DeletionFilter
	}{
		{name: "estado vazio usa não removidos", state: "", want: repository.DeletionFilterNotDeleted},
		{name: "removidos", state: dto.DeletionStateDeleted, want: repository.DeletionFilterDeleted},
		{name: "todos", state: dto.DeletionStateAll, want: repository.DeletionFilterAll},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := mapUserDeletionFilter(testCase.state)
			if got != testCase.want {
				t.Errorf("filtro = %v; esperado = %v", got, testCase.want)
			}
		})
	}
}

// TestUserServiceDelegatesOperations organiza em subtests os casos de uso
// sem regra adicional, apenas repasse ao Repository.
func TestUserServiceDelegatesOperations(t *testing.T) {
	t.Run("busca usuário por id", func(t *testing.T) {
		want := mustNewUser(t, "user-1", domain.RoleCustomer)
		fakeRepository := &fakeUserRepository{
			findByIDFunc: func(id string) (*domain.User, error) {
				if id != "user-1" {
					t.Fatalf("id recebido = %s; esperado = user-1", id)
				}
				return want, nil
			},
		}

		got, err := NewUserService(fakeRepository).FindByID(context.Background(), "user-1")

		if err != nil {
			t.Fatalf("FindByID retornou erro inesperado: %v", err)
		}
		if got.ID != want.ID() {
			t.Errorf("usuário recebido = %#v", got)
		}
	})

	t.Run("exclui usuário", func(t *testing.T) {
		deletedID := ""
		fakeRepository := &fakeUserRepository{
			deleteByIDFunc: func(id string) error {
				deletedID = id
				return nil
			},
		}

		err := NewUserService(fakeRepository).DeleteByID(context.Background(), "user-2")

		if err != nil {
			t.Fatalf("DeleteByID retornou erro inesperado: %v", err)
		}
		if deletedID != "user-2" {
			t.Errorf("id excluído = %s; esperado = user-2", deletedID)
		}
	})

	t.Run("restaura usuário", func(t *testing.T) {
		restoredID := ""
		fakeRepository := &fakeUserRepository{
			restoreByIDFunc: func(id string) error {
				restoredID = id
				return nil
			},
		}

		err := NewUserService(fakeRepository).RestoreByID(context.Background(), "user-3")

		if err != nil {
			t.Fatalf("RestoreByID retornou erro inesperado: %v", err)
		}
		if restoredID != "user-3" {
			t.Errorf("id restaurado = %s; esperado = user-3", restoredID)
		}
	})

	t.Run("ativa usuário", func(t *testing.T) {
		activatedID := ""
		fakeRepository := &fakeUserRepository{
			activateByIDFunc: func(id string) error {
				activatedID = id
				return nil
			},
		}

		err := NewUserService(fakeRepository).ActivateByID(context.Background(), "user-4")

		if err != nil {
			t.Fatalf("ActivateByID retornou erro inesperado: %v", err)
		}
		if activatedID != "user-4" {
			t.Errorf("id ativado = %s; esperado = user-4", activatedID)
		}
	})

	t.Run("desativa usuário", func(t *testing.T) {
		deactivatedID := ""
		fakeRepository := &fakeUserRepository{
			deactivateByIDFunc: func(id string) error {
				deactivatedID = id
				return nil
			},
		}

		err := NewUserService(fakeRepository).DeactivateByID(context.Background(), "user-5")

		if err != nil {
			t.Fatalf("DeactivateByID retornou erro inesperado: %v", err)
		}
		if deactivatedID != "user-5" {
			t.Errorf("id desativado = %s; esperado = user-5", deactivatedID)
		}
	})
}
