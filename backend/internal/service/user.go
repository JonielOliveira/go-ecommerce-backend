package service

import (
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/mapper"
	"ecommerce/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) *UserService {
	return &UserService{
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

func (s *UserService) Create(request dto.UserRequest) (dto.UserResponse, error) {
	user, err := mapper.NewUser(request)
	if err != nil {
		return dto.UserResponse{}, err
	}

	passwordHash, err := hashPassword(request.Password)
	if err != nil {
		return dto.UserResponse{}, err
	}

	createdUser, err := s.repository.Create(user, passwordHash)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return mapper.NewUserResponse(createdUser), nil
}

func (s *UserService) Update(id string, request dto.UserUpdateRequest) (dto.UserResponse, error) {
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

func (s *UserService) FindByID(id string) (dto.UserResponse, error) {
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

func (s *UserService) Search(filter dto.UserSearchRequest) (dto.UserPageResponse, error) {
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

func (s *UserService) DeleteByID(id string) error {
	return s.repository.DeleteByID(id)
}

func (s *UserService) RestoreByID(id string) error {
	return s.repository.RestoreByID(id)
}

func (s *UserService) ActivateByID(id string) error {
	return s.repository.ActivateByID(id)
}

func (s *UserService) DeactivateByID(id string) error {
	return s.repository.DeactivateByID(id)
}
