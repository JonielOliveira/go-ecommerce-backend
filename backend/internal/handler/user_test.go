package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
)

// fakeUserService substitui a Service real diretamente — UserService é uma
// interface, então o teste de Handler isola o Handler por completo, testando
// só binding e mapeamento de status, sem exercitar as regras da Service
// (essas já são cobertas em internal/service/user_test.go).
type fakeUserService struct {
	registerFunc       func(request dto.RegisterRequest) (dto.UserResponse, error)
	createFunc         func(request dto.CreateUserRequest) (dto.UserResponse, error)
	updateFunc         func(id string, request dto.UserUpdateRequest) (dto.UserResponse, error)
	findByIDFunc       func(id string) (dto.UserResponse, error)
	searchFunc         func(filter dto.UserSearchRequest) (dto.UserPageResponse, error)
	deleteByIDFunc     func(id string) error
	restoreByIDFunc    func(id string) error
	activateByIDFunc   func(id string) error
	deactivateByIDFunc func(id string) error
}

func (fake *fakeUserService) Register(_ context.Context, request dto.RegisterRequest) (dto.UserResponse, error) {
	return fake.registerFunc(request)
}

func (fake *fakeUserService) Create(_ context.Context, request dto.CreateUserRequest) (dto.UserResponse, error) {
	return fake.createFunc(request)
}

func (fake *fakeUserService) Update(_ context.Context, id string, request dto.UserUpdateRequest) (dto.UserResponse, error) {
	return fake.updateFunc(id, request)
}

func (fake *fakeUserService) FindByID(_ context.Context, id string) (dto.UserResponse, error) {
	return fake.findByIDFunc(id)
}

func (fake *fakeUserService) Search(_ context.Context, filter dto.UserSearchRequest) (dto.UserPageResponse, error) {
	return fake.searchFunc(filter)
}

func (fake *fakeUserService) DeleteByID(_ context.Context, id string) error {
	return fake.deleteByIDFunc(id)
}

func (fake *fakeUserService) RestoreByID(_ context.Context, id string) error {
	return fake.restoreByIDFunc(id)
}

func (fake *fakeUserService) ActivateByID(_ context.Context, id string) error {
	return fake.activateByIDFunc(id)
}

func (fake *fakeUserService) DeactivateByID(_ context.Context, id string) error {
	return fake.deactivateByIDFunc(id)
}

var validCreateUserRequest = dto.CreateUserRequest{
	Name: "Ana Souza", Email: "ana@example.com", Password: "segredo123",
}

var validUserUpdateRequest = dto.UserUpdateRequest{
	Name: "Ana Souza Silva", Email: "ana@example.com", Role: string(domain.RoleCustomer),
}

// TestUserHandlerCreate verifica o binding e o mapeamento completo do switch
// de erro: e-mail duplicado (409), dados inválidos em cada variante (400) e
// erro inesperado (500) — além do 400 de corpo malformado, que nem chega a
// chamar a Service.
func TestUserHandlerCreate(t *testing.T) {
	t.Run("corpo inválido retorna 400 sem chamar a Service", func(t *testing.T) {
		serviceCalls := 0
		fakeService := &fakeUserService{
			createFunc: func(dto.CreateUserRequest) (dto.UserResponse, error) {
				serviceCalls++
				return dto.UserResponse{}, nil
			},
		}
		userHandler := NewUserHandler(fakeService)

		context, recorder := newTestContext(http.MethodPost, "/api/v1/users", []byte("{invalid"))

		serve(context, userHandler.Create)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
		if serviceCalls != 0 {
			t.Errorf("chamadas à Service = %d; esperado = 0", serviceCalls)
		}
	})

	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "cria usuário e retorna 201", wantStatus: http.StatusCreated},
		{name: "email já existe retorna 409", serviceErr: domain.ErrUserEmailAlreadyExists, wantStatus: http.StatusConflict},
		{name: "nome inválido retorna 400", serviceErr: domain.ErrInvalidUserName, wantStatus: http.StatusBadRequest},
		{name: "email inválido retorna 400", serviceErr: domain.ErrInvalidUserEmail, wantStatus: http.StatusBadRequest},
		{name: "papel inválido retorna 400", serviceErr: domain.ErrInvalidUserRole, wantStatus: http.StatusBadRequest},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeUserService{
				createFunc: func(dto.CreateUserRequest) (dto.UserResponse, error) {
					return dto.UserResponse{ID: "user-1"}, testCase.serviceErr
				},
			}
			userHandler := NewUserHandler(fakeService)

			context, recorder := newJSONTestContext(t, http.MethodPost, "/api/v1/users", validCreateUserRequest)

			serve(context, userHandler.Create)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestUserHandlerUpdate verifica o mapeamento completo do switch de erro:
// não encontrado (404), já removido (409), e-mail duplicado (409), dados
// inválidos em cada variante (400) e erro inesperado (500).
func TestUserHandlerUpdate(t *testing.T) {
	t.Run("corpo inválido retorna 400 sem chamar a Service", func(t *testing.T) {
		serviceCalls := 0
		fakeService := &fakeUserService{
			updateFunc: func(string, dto.UserUpdateRequest) (dto.UserResponse, error) {
				serviceCalls++
				return dto.UserResponse{}, nil
			},
		}
		userHandler := NewUserHandler(fakeService)

		context, recorder := newTestContext(http.MethodPut, "/api/v1/users/user-1", []byte("{invalid"))
		setIDParam(context, "user-1")

		serve(context, userHandler.Update)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
		if serviceCalls != 0 {
			t.Errorf("chamadas à Service = %d; esperado = 0", serviceCalls)
		}
	})

	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "atualiza usuário e retorna 200", wantStatus: http.StatusOK},
		{name: "usuário não encontrado retorna 404", serviceErr: domain.ErrUserNotFound, wantStatus: http.StatusNotFound},
		{name: "usuário removido retorna 409", serviceErr: domain.ErrUserAlreadyDeleted, wantStatus: http.StatusConflict},
		{name: "email já usado por outro usuário retorna 409", serviceErr: domain.ErrUserEmailAlreadyExists, wantStatus: http.StatusConflict},
		{name: "nome inválido retorna 400", serviceErr: domain.ErrInvalidUserName, wantStatus: http.StatusBadRequest},
		{name: "email inválido retorna 400", serviceErr: domain.ErrInvalidUserEmail, wantStatus: http.StatusBadRequest},
		{name: "papel inválido retorna 400", serviceErr: domain.ErrInvalidUserRole, wantStatus: http.StatusBadRequest},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeUserService{
				updateFunc: func(id string, _ dto.UserUpdateRequest) (dto.UserResponse, error) {
					return dto.UserResponse{ID: id}, testCase.serviceErr
				},
			}
			userHandler := NewUserHandler(fakeService)

			context, recorder := newJSONTestContext(t, http.MethodPut, "/api/v1/users/user-1", validUserUpdateRequest)
			setIDParam(context, "user-1")

			serve(context, userHandler.Update)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestUserHandlerFindByID verifica sucesso, não encontrado e erro inesperado.
func TestUserHandlerFindByID(t *testing.T) {
	t.Run("retorna usuário e 200", func(t *testing.T) {
		fakeService := &fakeUserService{
			findByIDFunc: func(id string) (dto.UserResponse, error) { return dto.UserResponse{ID: id}, nil },
		}
		userHandler := NewUserHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/users/user-1", nil)
		setIDParam(context, "user-1")

		serve(context, userHandler.FindByID)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusOK)
		}
		got := decodeJSONBody[dto.UserResponse](t, recorder)
		if got.ID != "user-1" {
			t.Errorf("usuário recebido = %#v", got)
		}
	})

	t.Run("usuário não encontrado retorna 404", func(t *testing.T) {
		fakeService := &fakeUserService{
			findByIDFunc: func(string) (dto.UserResponse, error) { return dto.UserResponse{}, domain.ErrUserNotFound },
		}
		userHandler := NewUserHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/users/user-404", nil)
		setIDParam(context, "user-404")

		serve(context, userHandler.FindByID)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusNotFound)
		}
	})

	t.Run("erro inesperado retorna 500", func(t *testing.T) {
		fakeService := &fakeUserService{
			findByIDFunc: func(string) (dto.UserResponse, error) {
				return dto.UserResponse{}, errors.New("db indisponível")
			},
		}
		userHandler := NewUserHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/users/user-1", nil)
		setIDParam(context, "user-1")

		serve(context, userHandler.FindByID)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusInternalServerError)
		}
	})
}

// TestUserHandlerSearch verifica o binding dos parâmetros de query e a
// propagação de erro da Service.
func TestUserHandlerSearch(t *testing.T) {
	t.Run("retorna página de usuários e 200", func(t *testing.T) {
		fakeService := &fakeUserService{
			searchFunc: func(dto.UserSearchRequest) (dto.UserPageResponse, error) { return dto.UserPageResponse{}, nil },
		}
		userHandler := NewUserHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/users?page=1&pageSize=10", nil)

		serve(context, userHandler.Search)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, http.StatusOK, recorder.Body)
		}
	})

	t.Run("parâmetros inválidos retornam 400 sem chamar a Service", func(t *testing.T) {
		serviceCalls := 0
		fakeService := &fakeUserService{
			searchFunc: func(dto.UserSearchRequest) (dto.UserPageResponse, error) {
				serviceCalls++
				return dto.UserPageResponse{}, nil
			},
		}
		userHandler := NewUserHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/users?page=abc", nil)

		serve(context, userHandler.Search)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusBadRequest)
		}
		if serviceCalls != 0 {
			t.Errorf("chamadas à Service = %d; esperado = 0", serviceCalls)
		}
	})

	t.Run("erro da Service retorna 500", func(t *testing.T) {
		fakeService := &fakeUserService{
			searchFunc: func(dto.UserSearchRequest) (dto.UserPageResponse, error) {
				return dto.UserPageResponse{}, errors.New("db indisponível")
			},
		}
		userHandler := NewUserHandler(fakeService)

		context, recorder := newTestContext(http.MethodGet, "/api/v1/users", nil)

		serve(context, userHandler.Search)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d; esperado = %d", recorder.Code, http.StatusInternalServerError)
		}
	})
}

// TestUserHandlerDeleteByID cobre 204, 404, 409 e 500.
func TestUserHandlerDeleteByID(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "sucesso retorna 204", wantStatus: http.StatusNoContent},
		{name: "usuário não encontrado retorna 404", serviceErr: domain.ErrUserNotFound, wantStatus: http.StatusNotFound},
		{name: "usuário já removido retorna 409", serviceErr: domain.ErrUserAlreadyDeleted, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeUserService{
				deleteByIDFunc: func(string) error { return testCase.serviceErr },
			}
			userHandler := NewUserHandler(fakeService)

			context, recorder := newTestContext(http.MethodDelete, "/api/v1/users/user-1", nil)
			setIDParam(context, "user-1")

			serve(context, userHandler.DeleteByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestUserHandlerRestoreByID cobre 204, 404, 409 (não removido) e 500.
func TestUserHandlerRestoreByID(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "sucesso retorna 204", wantStatus: http.StatusNoContent},
		{name: "usuário não encontrado retorna 404", serviceErr: domain.ErrUserNotFound, wantStatus: http.StatusNotFound},
		{name: "usuário não removido retorna 409", serviceErr: domain.ErrUserNotDeleted, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeUserService{
				restoreByIDFunc: func(string) error { return testCase.serviceErr },
			}
			userHandler := NewUserHandler(fakeService)

			context, recorder := newTestContext(http.MethodPatch, "/api/v1/users/user-1/restore", nil)
			setIDParam(context, "user-1")

			serve(context, userHandler.RestoreByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestUserHandlerActivateByID cobre 204, 404, 409 (removido ou já ativo) e 500.
func TestUserHandlerActivateByID(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "sucesso retorna 204", wantStatus: http.StatusNoContent},
		{name: "usuário não encontrado retorna 404", serviceErr: domain.ErrUserNotFound, wantStatus: http.StatusNotFound},
		{name: "usuário removido retorna 409", serviceErr: domain.ErrUserAlreadyDeleted, wantStatus: http.StatusConflict},
		{name: "usuário já ativo retorna 409", serviceErr: domain.ErrUserAlreadyActive, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeUserService{
				activateByIDFunc: func(string) error { return testCase.serviceErr },
			}
			userHandler := NewUserHandler(fakeService)

			context, recorder := newTestContext(http.MethodPatch, "/api/v1/users/user-1/activate", nil)
			setIDParam(context, "user-1")

			serve(context, userHandler.ActivateByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}

// TestUserHandlerDeactivateByID cobre 204, 404, 409 (removido ou já inativo)
// e 500.
func TestUserHandlerDeactivateByID(t *testing.T) {
	testCases := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "sucesso retorna 204", wantStatus: http.StatusNoContent},
		{name: "usuário não encontrado retorna 404", serviceErr: domain.ErrUserNotFound, wantStatus: http.StatusNotFound},
		{name: "usuário removido retorna 409", serviceErr: domain.ErrUserAlreadyDeleted, wantStatus: http.StatusConflict},
		{name: "usuário já inativo retorna 409", serviceErr: domain.ErrUserAlreadyInactive, wantStatus: http.StatusConflict},
		{name: "erro inesperado retorna 500", serviceErr: errors.New("db indisponível"), wantStatus: http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fakeService := &fakeUserService{
				deactivateByIDFunc: func(string) error { return testCase.serviceErr },
			}
			userHandler := NewUserHandler(fakeService)

			context, recorder := newTestContext(http.MethodPatch, "/api/v1/users/user-1/deactivate", nil)
			setIDParam(context, "user-1")

			serve(context, userHandler.DeactivateByID)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d; esperado = %d; body = %s", recorder.Code, testCase.wantStatus, recorder.Body)
			}
		})
	}
}
