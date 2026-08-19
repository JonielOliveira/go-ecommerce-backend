package main

import (
	"context"
	"errors"
	"log/slog"

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
func seedDefaultAdmin(logger *slog.Logger, userService service.UserService) {
	role := string(domain.RoleAdmin)

	_, err := userService.Create(context.Background(), dto.CreateUserRequest{
		Name:     defaultAdminName,
		Email:    defaultAdminEmail,
		Password: defaultAdminPassword,
		Role:     &role,
	})

	if err == nil {
		logger.Info("usuário admin padrão criado", slog.String("operation", "seed.default_admin"), slog.String("email", defaultAdminEmail))
		return
	}

	if errors.Is(err, domain.ErrUserEmailAlreadyExists) {
		return
	}

	logger.Error("não foi possível criar o usuário admin padrão", slog.String("operation", "seed.default_admin"), slog.String("error", err.Error()))
}
