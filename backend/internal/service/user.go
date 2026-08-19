package service

import (
	"context"
	"log/slog"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/logging"
	"ecommerce/internal/mapper"
	"ecommerce/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserService declara os casos de uso de usuário consumidos pelo Handler.
// O tipo concreto abaixo satisfaz o contrato implicitamente.
type UserService interface {
	Register(ctx context.Context, request dto.RegisterRequest) (dto.UserResponse, error)
	Create(ctx context.Context, request dto.CreateUserRequest) (dto.UserResponse, error)
	Update(ctx context.Context, id string, request dto.UserUpdateRequest) (dto.UserResponse, error)
	FindByID(ctx context.Context, id string) (dto.UserResponse, error)
	Search(ctx context.Context, filter dto.UserSearchRequest) (dto.UserPageResponse, error)
	DeleteByID(ctx context.Context, id string) error
	RestoreByID(ctx context.Context, id string) error
	ActivateByID(ctx context.Context, id string) error
	DeactivateByID(ctx context.Context, id string) error
}

type userService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) UserService {
	return &userService{
		repository: repository,
	}
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CreateUserInput reúne os dados necessários para criar um usuário,
// independentemente de a chamada vir do autocadastro público ou da
// criação administrativa — ambas convergem em createUser.
type CreateUserInput struct {
	Name      string
	Email     string
	Password  string
	Role      domain.UserRole
	AvatarURL *string
}

func (s *userService) createUser(ctx context.Context, input CreateUserInput) (dto.UserResponse, error) {
	user, err := domain.NewUser(input.Name, input.Email, input.Role, input.AvatarURL)
	if err != nil {
		return dto.UserResponse{}, err
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return dto.UserResponse{}, err
	}

	createdUser, err := s.repository.Create(ctx, user, passwordHash)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return mapper.NewUserResponse(createdUser), nil
}

// Register é o autocadastro público (POST /auth/register): sempre cria um
// usuário "customer", sem exceção.
func (s *userService) Register(ctx context.Context, request dto.RegisterRequest) (dto.UserResponse, error) {
	response, err := s.createUser(ctx, CreateUserInput{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
		Role:     domain.RoleCustomer,
	})
	if err != nil {
		logging.FromContext(ctx).Warn("falha no autocadastro",
			slog.String("operation", "user.register"),
			slog.String("error", err.Error()),
		)
		return dto.UserResponse{}, err
	}

	logging.FromContext(ctx).Info("usuário registrado",
		slog.String("operation", "user.register"),
		slog.String("user_id", response.ID),
	)

	return response, nil
}

// Create é a criação administrativa (POST /users, restrita a admins):
// aceita um papel opcional, com "customer" como padrão quando omitido.
func (s *userService) Create(ctx context.Context, request dto.CreateUserRequest) (dto.UserResponse, error) {
	role := domain.RoleCustomer

	if request.Role != nil {
		role = domain.UserRole(*request.Role)
	}

	if !role.IsValid() {
		return dto.UserResponse{}, domain.ErrInvalidUserRole
	}

	response, err := s.createUser(ctx, CreateUserInput{
		Name:      request.Name,
		Email:     request.Email,
		Password:  request.Password,
		Role:      role,
		AvatarURL: request.AvatarURL,
	})
	if err != nil {
		logging.FromContext(ctx).Warn("falha ao criar usuário",
			slog.String("operation", "user.create"),
			slog.String("error", err.Error()),
		)
		return dto.UserResponse{}, err
	}

	logging.FromContext(ctx).Info("usuário criado",
		slog.String("operation", "user.create"),
		slog.String("user_id", response.ID),
		slog.String("role", response.Role),
	)

	return response, nil
}

func (s *userService) Update(ctx context.Context, id string, request dto.UserUpdateRequest) (dto.UserResponse, error) {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return dto.UserResponse{}, err
	}

	if user.IsDeleted() {
		return dto.UserResponse{}, domain.ErrUserAlreadyDeleted
	}

	if err := user.Update(
		request.Name,
		request.Email,
		domain.UserRole(request.Role),
		request.AvatarURL,
	); err != nil {
		return dto.UserResponse{}, err
	}

	var passwordHash *string

	if request.Password != "" {
		hash, err := hashPassword(request.Password)
		if err != nil {
			return dto.UserResponse{}, err
		}

		passwordHash = &hash
	}

	updatedUser, err := s.repository.Update(ctx, user, passwordHash)
	if err != nil {
		logging.FromContext(ctx).Warn("falha ao atualizar usuário",
			slog.String("operation", "user.update"),
			slog.String("user_id", id),
			slog.String("error", err.Error()),
		)
		return dto.UserResponse{}, err
	}

	logging.FromContext(ctx).Info("usuário atualizado",
		slog.String("operation", "user.update"),
		slog.String("user_id", updatedUser.ID()),
	)

	return mapper.NewUserResponse(updatedUser), nil
}

func (s *userService) FindByID(ctx context.Context, id string) (dto.UserResponse, error) {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return mapper.NewUserResponse(user), nil
}

func mapUserDeletionFilter(state dto.DeletionState) repository.DeletionFilter {
	switch state {
	case dto.DeletionStateDeleted:
		return repository.DeletionFilterDeleted

	case dto.DeletionStateAll:
		return repository.DeletionFilterAll

	default:
		return repository.DeletionFilterNotDeleted
	}
}

func (s *userService) Search(ctx context.Context, filter dto.UserSearchRequest) (dto.UserPageResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	repositoryFilter := repository.UserSearchFilter{
		Name:           filter.Name,
		Email:          filter.Email,
		Role:           filter.Role,
		Active:         filter.Active,
		DeletionFilter: mapUserDeletionFilter(filter.DeletionState),
		Limit:          filter.PageSize,
		Offset:         (filter.Page - 1) * filter.PageSize,
	}

	result, err := s.repository.Search(ctx, repositoryFilter)
	if err != nil {
		return dto.UserPageResponse{}, err
	}

	items := make([]dto.UserResponse, 0, len(result.Users))

	for _, user := range result.Users {
		items = append(items, mapper.NewUserResponse(user))
	}

	totalPages := int(
		(result.Total + int64(filter.PageSize) - 1) /
			int64(filter.PageSize),
	)

	return dto.UserPageResponse{
		Items:      items,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalItems: result.Total,
		TotalPages: totalPages,
	}, nil
}

func (s *userService) DeleteByID(ctx context.Context, id string) error {
	if err := s.repository.DeleteByID(ctx, id); err != nil {
		logging.FromContext(ctx).Warn("falha ao excluir usuário",
			slog.String("operation", "user.delete"),
			slog.String("user_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	logging.FromContext(ctx).Info("usuário excluído",
		slog.String("operation", "user.delete"),
		slog.String("user_id", id),
	)

	return nil
}

func (s *userService) RestoreByID(ctx context.Context, id string) error {
	if err := s.repository.RestoreByID(ctx, id); err != nil {
		logging.FromContext(ctx).Warn("falha ao restaurar usuário",
			slog.String("operation", "user.restore"),
			slog.String("user_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	logging.FromContext(ctx).Info("usuário restaurado",
		slog.String("operation", "user.restore"),
		slog.String("user_id", id),
	)

	return nil
}

func (s *userService) ActivateByID(ctx context.Context, id string) error {
	if err := s.repository.ActivateByID(ctx, id); err != nil {
		logging.FromContext(ctx).Warn("falha ao ativar usuário",
			slog.String("operation", "user.activate"),
			slog.String("user_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	logging.FromContext(ctx).Info("usuário ativado",
		slog.String("operation", "user.activate"),
		slog.String("user_id", id),
	)

	return nil
}

func (s *userService) DeactivateByID(ctx context.Context, id string) error {
	if err := s.repository.DeactivateByID(ctx, id); err != nil {
		logging.FromContext(ctx).Warn("falha ao desativar usuário",
			slog.String("operation", "user.deactivate"),
			slog.String("user_id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	logging.FromContext(ctx).Info("usuário desativado",
		slog.String("operation", "user.deactivate"),
		slog.String("user_id", id),
	)

	return nil
}
