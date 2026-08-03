package service

import (
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/mapper"
	"ecommerce/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserService declara os casos de uso de usuário consumidos pelo Handler.
// O tipo concreto abaixo satisfaz o contrato implicitamente.
type UserService interface {
	Register(request dto.RegisterRequest) (dto.UserResponse, error)
	Create(request dto.CreateUserRequest) (dto.UserResponse, error)
	Update(id string, request dto.UserUpdateRequest) (dto.UserResponse, error)
	FindByID(id string) (dto.UserResponse, error)
	Search(filter dto.UserSearchRequest) (dto.UserPageResponse, error)
	DeleteByID(id string) error
	RestoreByID(id string) error
	ActivateByID(id string) error
	DeactivateByID(id string) error
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

func (s *userService) createUser(input CreateUserInput) (dto.UserResponse, error) {
	user, err := domain.NewUser(input.Name, input.Email, input.Role, input.AvatarURL)
	if err != nil {
		return dto.UserResponse{}, err
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return dto.UserResponse{}, err
	}

	createdUser, err := s.repository.Create(user, passwordHash)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return mapper.NewUserResponse(createdUser), nil
}

// Register é o autocadastro público (POST /auth/register): sempre cria um
// usuário "customer", sem exceção.
func (s *userService) Register(request dto.RegisterRequest) (dto.UserResponse, error) {
	return s.createUser(CreateUserInput{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
		Role:     domain.RoleCustomer,
	})
}

// Create é a criação administrativa (POST /users, restrita a admins):
// aceita um papel opcional, com "customer" como padrão quando omitido.
func (s *userService) Create(request dto.CreateUserRequest) (dto.UserResponse, error) {
	role := domain.RoleCustomer

	if request.Role != nil {
		role = domain.UserRole(*request.Role)
	}

	if !role.IsValid() {
		return dto.UserResponse{}, domain.ErrInvalidUserRole
	}

	return s.createUser(CreateUserInput{
		Name:      request.Name,
		Email:     request.Email,
		Password:  request.Password,
		Role:      role,
		AvatarURL: request.AvatarURL,
	})
}

func (s *userService) Update(id string, request dto.UserUpdateRequest) (dto.UserResponse, error) {
	user, err := s.repository.FindByID(id)
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

	updatedUser, err := s.repository.Update(user, passwordHash)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return mapper.NewUserResponse(updatedUser), nil
}

func (s *userService) FindByID(id string) (dto.UserResponse, error) {
	user, err := s.repository.FindByID(id)
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

func (s *userService) Search(filter dto.UserSearchRequest) (dto.UserPageResponse, error) {
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

	result, err := s.repository.Search(repositoryFilter)
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

func (s *userService) DeleteByID(id string) error {
	return s.repository.DeleteByID(id)
}

func (s *userService) RestoreByID(id string) error {
	return s.repository.RestoreByID(id)
}

func (s *userService) ActivateByID(id string) error {
	return s.repository.ActivateByID(id)
}

func (s *userService) DeactivateByID(id string) error {
	return s.repository.DeactivateByID(id)
}
