package main

import (
	"errors"
	"log"

	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/service"
)

const (
	defaultAdminName     = "Administrador"
	defaultAdminEmail    = "admin@gmail.com"
	defaultAdminPassword = "senha123"
)

// seedDefaultAdmin garante que exista um usuário administrador padrão a
// cada inicialização da aplicação. Não há uma consulta prévia de
// existência: a própria tentativa de criação já resolve isso, já que o
// repository mapeia e-mail duplicado (constraint única do banco) para
// domain.ErrUserEmailAlreadyExists — nesse caso, não faz nada.
func seedDefaultAdmin(userService *service.UserService) {
	role := string(domain.RoleAdmin)

	_, err := userService.Create(dto.CreateUserRequest{
		Name:     defaultAdminName,
		Email:    defaultAdminEmail,
		Password: defaultAdminPassword,
		Role:     &role,
	})

	if err == nil {
		log.Printf("Usuário admin padrão criado: %s", defaultAdminEmail)
		return
	}

	if errors.Is(err, domain.ErrUserEmailAlreadyExists) {
		return
	}

	log.Printf("Não foi possível criar o usuário admin padrão: %v", err)
}
