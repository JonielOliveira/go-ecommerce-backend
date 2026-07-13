# go-ecommerce-backend

Backend de e-commerce em Go (Gin + PostgreSQL), com autenticação JWT em cookie `HttpOnly`, autorização por papel (`customer`/`admin`) e os módulos de usuários, produtos e pedidos.

## Sobre o projeto

A API expõe:

- **Autenticação**: cadastro público (`/auth/register`), login/logout com JWT em cookie `HttpOnly`, e `/auth/me` para o usuário autenticado.
- **Usuários**: CRUD administrativo (exclusivo para `admin`), com soft delete e ativação/desativação.
- **Produtos**: leitura pública, escrita e gestão exclusivas para `admin`.
- **Pedidos**: criação, listagem paginada, pagamento e cancelamento, sempre atrelados ao dono autenticado (`customer` só vê/paga/cancela os próprios; `admin` pode listar e cancelar qualquer pedido, mas só paga os próprios).

Toda a documentação interativa dos endpoints fica disponível via **Swagger** depois que a aplicação sobe (veja o passo 6 abaixo).

## Stack

- Go 1.26
- [Gin](https://github.com/gin-gonic/gin)
- PostgreSQL 18+ (usa a função nativa `uuidv7()`)
- [pgx/v5](https://github.com/jackc/pgx) (`pgxpool`)
- JWT (`golang-jwt/jwt/v5`) + `bcrypt`
- [swaggo](https://github.com/swaggo/swag) para geração do Swagger/OpenAPI

## Pré-requisitos

- [Go 1.26+](https://go.dev/dl/)
- PostgreSQL 18+ (local ou via Docker)
- Opcional, mas recomendado: [Docker](https://www.docker.com/) + [VS Code Dev Containers](https://code.visualstudio.com/docs/devcontainers/containers) — o repositório já traz um Dev Container pronto (`.devcontainer/`), com Postgres, o binário `migrate` e o `air` (hot reload) já instalados.

## Como rodar (passo a passo)

### 1. Clonar o repositório

```bash
git clone <url-do-repositorio>
cd go-ecommerce-backend
```

### 2. Subir o banco de dados

**Opção A — Dev Container (mais simples):** abra a pasta no VS Code, instale a extensão *Dev Containers* e escolha "Reopen in Container". O Postgres sobe automaticamente junto com o container da aplicação (serviço `db`, porta `5432`), e as ferramentas `migrate`/`air`/`sqlc` já vêm instaladas. Todos os comandos abaixo (a partir do passo 4) devem ser rodados no terminal *dentro* do container.

**Opção B — Local:** suba só o banco via Docker Compose:

```bash
docker compose -f .devcontainer/docker-compose.yml up -d db
```

Isso deixa o Postgres disponível em `localhost:5432` (usuário/senha `postgres`, banco `postgres`). Se preferir, use uma instância Postgres 18+ já existente na sua máquina.

### 3. Configurar o `.env`

```bash
cd backend
cp .env.example .env
```

Ajuste os valores conforme seu ambiente (veja a tabela completa de variáveis mais abaixo). Os padrões do `.env.example` já funcionam com a Opção B do passo 2. Se estiver usando o Dev Container, troque `DB_HOST=db` (nome do serviço na rede Docker) em vez de `localhost`.

A aplicação **não sobe** sem `JWT_SECRET` preenchido — é validado no início e falha de propósito (`log.Fatal`) se estiver vazio.

### 4. Aplicar as migrations

Com o `migrate` (já instalado no Dev Container):

```bash
migrate -path database/migrations \
  -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
  up
```

Sem o `migrate`, aplique o SQL diretamente:

```bash
psql -h localhost -U postgres -d postgres -f database/migrations/000001_initial_schema.up.sql
```

(ajuste host/usuário/banco conforme o seu `.env`.)

### 5. Popular o banco com dados de exemplo (opcional)

```bash
go run ./cmd/seed
```

Executa `database/seeds/seed.sql`: cria 10 usuários (5 `admin` + 5 `customer`, senha `Senha@123` para todos), 20 produtos e 10 pedidos de exemplo (em `PENDING`, `PAID` e `CANCELED`). É idempotente — pode rodar quantas vezes quiser, os dados do seed são recriados a cada execução. **Não use em produção.**

Esse passo é opcional: mesmo sem rodá-lo, a aplicação sempre garante um usuário administrador padrão (veja o próximo passo).

### 6. Rodar a aplicação

```bash
go run ./cmd/api
```

A API sobe em `http://localhost:8080` (porta configurável por `SERVER_PORT`). Na inicialização, a aplicação também garante a existência de um usuário administrador padrão:

```text
email: admin@gmail.com
senha: senha123
papel: admin
```

Se esse e-mail já existir (por exemplo, depois de rodar o seed), nada é alterado — a criação só acontece uma vez.

### 7. Acessar o Swagger

Com a aplicação rodando, a documentação interativa de todos os endpoints fica em:

```text
http://localhost:8080/swagger/index.html
```

Lá é possível ver todas as rotas, os schemas de request/response e testar chamadas diretamente pelo navegador (as rotas protegidas exigem o cookie de sessão — faça login primeiro por `/api/v1/auth/login`, o próprio navegador guarda o cookie).

### 8. Testar

Faça login com o admin padrão (ou com qualquer usuário do seed) em `POST /api/v1/auth/login`, ou crie sua própria conta em `POST /api/v1/auth/register` (sempre criada como `customer`). A partir daí, explore os endpoints pelo Swagger ou por um cliente HTTP (Postman, Insomnia, `curl`).

## Estrutura do projeto

```text
backend/
├── cmd/
│   ├── api/          # ponto de entrada da API (main.go)
│   └── seed/         # comando para popular o banco com dados de exemplo
├── database/
│   ├── migrations/   # migrations SQL versionadas
│   └── seeds/        # seed.sql (embutido no binário via go:embed)
├── docs/             # Swagger/OpenAPI gerado (swag init)
└── internal/
    ├── config/       # carregamento e validação de variáveis de ambiente
    ├── database/     # conexão com PostgreSQL (pgxpool)
    ├── domain/       # entidades e regras de negócio
    ├── dto/          # contratos de request/response da API
    ├── handler/      # HTTP (Gin) — lê request, chama service, monta resposta
    ├── mapper/       # conversão entre domínio e DTOs
    ├── middleware/    # autenticação (JWT em cookie) e autorização por papel
    ├── repository/   # acesso a dados (SQL via pgx)
    ├── routes/       # composição das rotas e middlewares
    ├── security/     # geração/validação de JWT
    └── service/      # orquestração das regras de negócio
```

## Principais grupos de rotas

| Grupo | Prefixo | Observação |
|---|---|---|
| Autenticação | `/api/v1/auth` | `register` e `login`/`logout` são públicos; `me` exige login |
| Usuários | `/api/v1/users` | cadastro público fica em `/auth/register`; todo o grupo `/users` exige `admin` |
| Produtos | `/api/v1/products` | leitura pública; escrita exige `admin` |
| Pedidos | `/api/v1/orders` | todo o grupo exige login (`customer` ou `admin`); regras de propriedade por operação |
| Saúde | `/health` | status da aplicação, fora do prefixo `/api/v1` |

Detalhes completos (parâmetros, schemas, códigos de status) estão no Swagger — veja o passo 7.

## Variáveis de ambiente

Arquivo: `backend/.env` (copiado de `backend/.env.example`).

| Variável | Descrição | Padrão |
|---|---|---|
| `APP_NAME` / `APP_VERSION` / `APP_ENV` | Metadados da aplicação (aparecem em `/health`) | — |
| `SERVER_HOST` / `SERVER_PORT` | Endereço e porta em que a API escuta | `0.0.0.0` / `8080` |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` / `DB_SSLMODE` | Conexão com o PostgreSQL | — |
| `JWT_SECRET` | Chave de assinatura do JWT — **obrigatória**, a aplicação não sobe sem ela | — |
| `JWT_ISSUER` / `JWT_AUDIENCE` | Claims `iss`/`aud` do token | — |
| `JWT_ACCESS_TOKEN_TTL_MINUTES` | Duração do access token | `15` |
| `AUTH_COOKIE_NAME` | Nome do cookie que guarda o JWT | `access_token` |
| `AUTH_COOKIE_SECURE` | `true` em produção (HTTPS); `false` em desenvolvimento local | `false` |
| `AUTH_COOKIE_SAME_SITE` | `lax`, `strict` ou `none` | `lax` |
| `AUTH_COOKIE_DOMAIN` | Domínio do cookie (opcional) | — |
| `CORS_ALLOWED_ORIGINS` | Lista de origens permitidas, separadas por vírgula — obrigatória quando `AllowCredentials` está ativo | — |
---
