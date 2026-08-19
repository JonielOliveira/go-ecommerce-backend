# Desafio do Módulo 03: do monólito aos microsserviços

Este desafio é uma continuação do `desafio-aula07-final.md`. Primeiro, você precisa possuir a API monolítica de pedidos descrita abaixo. Depois, deverá refatorá-la e extrair uma parte do negócio para um microsserviço.

## Parte 1 — Projeto base: API monolítica de pedidos

### Contexto

Evolua o projeto de pedidos que antes funcionava apenas no terminal e utilizava repositories em memória. Ele deve se tornar uma API REST em Go, com persistência em PostgreSQL.

### Objetivo

Criar uma API de pedidos para uma loja.

A aplicação deve permitir:

- cadastrar clientes;
- listar clientes;
- buscar cliente por ID;
- cadastrar produtos;
- listar produtos;
- buscar produto por ID;
- criar pedido para um cliente;
- listar pedidos com paginação;
- buscar pedido por ID;
- pagar pedido;
- cancelar pedido.

### Regras de negócio

- cliente do pedido é obrigatório;
- cliente precisa existir;
- pedido precisa ter pelo menos um item;
- quantidade deve ser maior que zero;
- produto precisa existir;
- estoque precisa ser suficiente;
- ao criar pedido, o estoque deve diminuir;
- pedido nasce como `PENDING`;
- pedido pago vira `PAID`;
- pedido cancelado vira `CANCELED` e devolve o estoque;
- pedido pago ou cancelado não pode mudar de status.

### Banco de dados

Use PostgreSQL com migrations.

Crie tabelas interligadas para:

- clientes;
- produtos;
- pedidos;
- itens do pedido.

Relacionamentos obrigatórios:

- `pedidos.cliente_id` deve apontar para `clientes.id`;
- `itens_pedido.pedido_id` deve apontar para `pedidos.id`;
- `itens_pedido.produto_id` deve apontar para `produtos.id`.

Schema mínimo sugerido para clientes:

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

Não salve senha em texto puro. Gere o hash antes de gravar o cliente no banco.

Use `pgxpool` para conexão com o banco.

A criação do pedido deve usar uma transação, pois altera pedidos, itens e estoque.

### Endpoints mínimos

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

### Organização inicial

O projeto monolítico deve possuir responsabilidades separadas para configuração, banco de dados, modelos, regras de negócio, repositories, controllers e rotas.

O controller deve lidar com HTTP e JSON. O repository deve lidar com SQL. As regras de negócio não devem ficar dentro do controller.

### Erros e status HTTP

Trate erros sem usar `panic`.

No mínimo:

- dados inválidos: `400`;
- cliente, produto ou pedido não encontrado: `404`;
- estoque insuficiente: `409`;
- e-mail de cliente já cadastrado: `409`;
- transição inválida de status do pedido: `409`;
- erro inesperado: `500`.

### Testes automatizados

O projeto monolítico deve possuir testes automatizados com cobertura mínima total de 40%.

### Critérios de aceite da Parte 1

O projeto base será considerado completo se:

- compilar sem erros;
- subir uma API HTTP;
- usar PostgreSQL e migrations;
- possuir as tabelas e chaves estrangeiras exigidas;
- não armazenar senhas em texto puro;
- possuir controllers, rotas e repositories;
- criar o pedido dentro de uma transação;
- atualizar o estoque corretamente;
- listar pedidos com `limit` e `offset`;
- tratar os erros e status HTTP exigidos;
- demonstrar fluxos de sucesso e de erro via HTTP;
- alcançar no mínimo 40% de cobertura total nos testes automatizados.

## Parte 2 — Evolução arquitetural e extração do microsserviço

Se você chegou até aqui, espera-se que já tenha o monólito da API de pedidos funcionando, com tudo o que foi pedido na Parte 1 e com pelo menos 40% de cobertura de testes.

Agora o projeto deve evoluir. Suas responsabilidades internas precisam ficar mais bem separadas e uma parte do negócio deve passar a operar como um serviço independente.

### Objetivo

Refatorar o projeto com DDD e Clean Architecture e extrair pelo menos uma capacidade de negócio para um microsserviço.

Ao final, a solução deve possuir:

- camadas e dependências bem definidas;
- pelo menos dois serviços executáveis de forma independente;
- comunicação assíncrona por mensageria;
- uma Saga para coordenar um fluxo distribuído;
- logs estruturados que permitam acompanhar o fluxo entre os serviços.

### 1. Refatoração arquitetural

Refatore o projeto sem remover as regras de negócio e as operações já exigidas no desafio da Aula 07.

A organização deve respeitar os seguintes limites:

- o domínio representa as regras, invariantes, estados e comportamentos do negócio;
- a aplicação implementa e coordena os casos de uso;
- os componentes de entrada tratam protocolos e formatos externos, como HTTP e JSON;
- os componentes de infraestrutura tratam PostgreSQL, mensageria, configurações e outros mecanismos externos;
- as dependências de código apontam das partes externas para as políticas internas;
- o domínio não depende de HTTP, PostgreSQL, drivers, broker ou framework;
- decisões de negócio não ficam em controllers, handlers ou repositories concretos.

Use conceitos de DDD onde eles forem relevantes, incluindo linguagem do negócio, entidades, objetos de valor, agregados, invariantes e erros de domínio.

### 2. Extração de um microsserviço

Escolha pelo menos uma capacidade do sistema para extrair. A escolha pode envolver, por exemplo:

- clientes;
- produtos ou estoque;
- pedidos;
- pagamentos;
- outra capacidade de negócio devidamente justificada.

O componente extraído deve:

- ser uma aplicação Go independente;
- possuir responsabilidade de negócio clara;
- poder ser compilado, iniciado e implantado separadamente;
- possuir configuração própria;
- ser responsável pelos próprios dados e migrations;
- não permitir que outro serviço leia ou altere diretamente suas tabelas;
- expor contratos claros para a comunicação necessária com os demais serviços.

Não é necessário transformar todo o sistema em microsserviços. O projeto final, porém, deve possuir pelo menos dois serviços participantes de um mesmo fluxo de negócio distribuído.

### 3. Mensageria

Escolha uma das tecnologias abaixo:

- Redpanda;
- Apache Kafka;
- RabbitMQ.

Implemente comunicação assíncrona entre os serviços em pelo menos um fluxo de negócio.

A solução deve possuir:

- publicação e consumo de mensagens reais pelo broker escolhido;
- eventos ou comandos com nomes e finalidades claros;
- identificação da operação ou entidade relacionada à mensagem;
- tratamento de falhas de publicação e consumo;
- estratégia para evitar que o processamento repetido da mesma mensagem corrompa o estado do sistema;
- configuração do broker fora do código-fonte.

### 4. Saga

Implemente uma Saga em um fluxo que envolva pelo menos dois serviços.

Escolha e implemente uma das abordagens:

- Saga coreografada;
- Saga orquestrada.

O fluxo deve contemplar:

- início da operação distribuída;
- participação de pelo menos dois serviços;
- caminho de sucesso;
- falha em pelo menos uma etapa;
- ação compensatória para desfazer ou neutralizar etapas já concluídas;
- correlação das mensagens pertencentes à mesma Saga;
- estado final consistente e identificável após sucesso ou falha.

A escolha entre coreografia e orquestração deve estar registrada no README, junto com os participantes, as etapas e as compensações do fluxo.

### 5. Logs estruturados

Implemente logs estruturados com o pacote `log/slog`.

Os logs devem:

- ser emitidos em formato JSON;
- identificar o serviço que gerou o registro;
- registrar operação, resultado e erro quando houver;
- incluir os identificadores relevantes do negócio;
- incluir um identificador de correlação ou da Saga nos registros do fluxo distribuído;
- permitir acompanhar uma mesma operação enquanto ela passa por serviços diferentes;
- evitar senhas, tokens e outros dados sensíveis.

Não use apenas mensagens de texto soltas para registrar os acontecimentos principais do fluxo distribuído.

### 6. Requisitos funcionais preservados

Todas as operações e regras de negócio exigidas na Parte 1 continuam válidas depois da refatoração. A separação em serviços não pode remover comportamentos nem permitir estados que eram inválidos no monólito.

O comportamento final também deve permanecer correto quando ocorrer uma falha durante a comunicação entre os serviços.

### 7. Documentação e execução

Entregue um README contendo:

- a capacidade escolhida para extração e a justificativa da fronteira;
- a responsabilidade de cada serviço;
- um diagrama simples da solução;
- a direção das dependências dentro de cada serviço;
- o broker escolhido;
- a abordagem de Saga escolhida;
- as etapas, mensagens e compensações da Saga;
- as variáveis de ambiente necessárias;
- os comandos para iniciar a infraestrutura e os serviços;
- instruções para demonstrar o caminho de sucesso e o caminho com falha e compensação.

Forneça também os arquivos necessários para executar localmente os bancos, o broker e os serviços. As migrations devem fazer parte da entrega.

### Critérios de aceite da Parte 2

O desafio será considerado completo se:

- o código compilar e os serviços iniciarem sem erros;
- as funcionalidades exigidas no desafio da Aula 07 continuarem disponíveis;
- domínio, aplicação, entrada e infraestrutura possuírem responsabilidades identificáveis;
- as dependências respeitarem as fronteiras definidas;
- pelo menos uma capacidade tiver sido extraída para um serviço independente;
- cada serviço for responsável pelos próprios dados;
- não houver acesso direto ao banco de outro serviço;
- Redpanda, Kafka ou RabbitMQ estiver integrado à aplicação;
- mensagens forem realmente produzidas e consumidas;
- uma Saga coreografada ou orquestrada estiver implementada;
- o caminho de sucesso da Saga puder ser demonstrado;
- uma falha e sua ação compensatória puderem ser demonstradas;
- mensagens repetidas não corromperem o estado do sistema;
- logs JSON permitirem correlacionar o fluxo entre os serviços;
- nenhuma informação sensível aparecer nos logs;
- o README permitir que outra pessoa execute e compreenda a solução.

### Entrega

Entregue o código-fonte completo, migrations, arquivos de configuração da infraestrutura, documentação e exemplos de requisições necessários para avaliar os critérios de aceite.
