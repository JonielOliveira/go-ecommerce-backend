ARG GO_VARIANT="1.26-bookworm"

FROM mcr.microsoft.com/devcontainers/go:${GO_VARIANT}

ARG INSTALL_POSTGRES_CLIENT="true"
ARG INSTALL_AIR="true"
ARG INSTALL_MIGRATE="true"
ARG INSTALL_SQLC="true"
ARG USERNAME="vscode"

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && if [ "${INSTALL_POSTGRES_CLIENT}" = "true" ]; then \
        apt-get -y install --no-install-recommends postgresql-client; \
    fi \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*

RUN if [ "${INSTALL_AIR}" = "true" ]; then \
        su ${USERNAME} -c "go install github.com/air-verse/air@latest"; \
    fi \
    && if [ "${INSTALL_MIGRATE}" = "true" ]; then \
        su ${USERNAME} -c "go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
    fi \
    && if [ "${INSTALL_SQLC}" = "true" ]; then \
        su ${USERNAME} -c "go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; \
    fi

ENV PATH="/home/${USERNAME}/go/bin:${PATH}"
