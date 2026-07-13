# Desafio Aula 06: Servico de Pedidos com API e Postgres

## Contexto

Evolua o desafio da aula 03.

Antes o sistema rodava apenas no terminal, com repositories em memoria. Agora ele deve ser uma API REST em Go, com controllers, rotas HTTP e persistencia em PostgreSQL, seguindo a ideia do projeto `cliente`.

## Objetivo

Criar uma API de pedidos para uma loja.

A aplicacao deve permitir:

- cadastrar clientes;
- listar clientes;
- buscar cliente por id;
- cadastrar produtos;
- listar produtos;
- buscar produto por id;
- criar pedido para um cliente;
- listar pedidos com paginacao;
- buscar pedido por id;
- pagar pedido;
- cancelar pedido.

## Regras de negocio

Mantenha as regras do desafio da aula 03:

- cliente do pedido e obrigatorio;
- cliente precisa existir;
- pedido precisa ter pelo menos um item;
- quantidade deve ser maior que zero;
- produto precisa existir;
- estoque precisa ser suficiente;
- ao criar pedido, o estoque deve diminuir;
- pedido nasce como `PENDING`;
- pedido pago vira `PAID`;
- pedido cancelado vira `CANCELED` e devolve estoque;
- pedido pago ou cancelado nao pode mudar de status.

## Banco de dados

Use PostgreSQL com migrations.

Crie tabelas interligadas para:

- clientes;
- produtos;
- pedidos;
- itens do pedido.

Relacionamentos obrigatorios:

- `pedidos.cliente_id` deve apontar para `clientes.id`;
- `itens_pedido.pedido_id` deve apontar para `pedidos.id`;
- `itens_pedido.produto_id` deve apontar para `produtos.id`.

Schema minimo sugerido:

```sql
CREATE TABLE clientes (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

No JSON da API, use o nome `passwordHash`. No banco, use `password_hash`.

Nao salve senha em texto puro. Gere o hash antes de gravar o cliente no banco.

Use `pgxpool` para conexao com o banco.

A criacao do pedido deve usar transacao, pois altera pedidos, itens e estoque.

## Endpoints minimos

```text
POST   /clientes
GET    /clientes
GET    /clientes/{id}

POST   /produtos
GET    /produtos
GET    /produtos/{id}

POST   /pedidos
GET    /pedidos?limit=10&offset=0
GET    /pedidos/{id}
POST   /pedidos/{id}/pagar
POST   /pedidos/{id}/cancelar
```

## Estrutura esperada

Organize o projeto em camadas, como no projeto `cliente`:

```text
cmd ou main.go
config/
database/
model/
repository/
controllers/
routes/
migrations/
```

O controller deve lidar com HTTP e JSON.

O repository deve lidar com SQL.

As regras de negocio devem ficar no dominio/service, nao dentro do controller.

## Erros e status HTTP

Trate erros sem usar `panic`.

No minimo:

- dados invalidos: `400`
- cliente, produto ou pedido nao encontrado: `404`
- estoque insuficiente: `409`
- email de cliente ja cadastrado: `409`
- status invalido do pedido: `409`
- erro inesperado: `500`

## Criterios de aceite

O projeto sera considerado completo se:

- compilar sem erros;
- subir uma API HTTP;
- usar PostgreSQL;
- tiver migrations;
- tiver tabela de clientes com `password_hash`;
- tiver chaves estrangeiras ligando clientes, pedidos, itens e produtos;
- tiver controllers e rotas;
- tiver repositories usando SQL;
- criar pedido dentro de transacao;
- atualizar estoque corretamente;
- listar pedidos com `limit` e `offset`;
- demonstrar um fluxo feliz e alguns fluxos de erro via HTTP.
