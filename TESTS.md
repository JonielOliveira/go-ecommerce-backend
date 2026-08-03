# Testes do backend

Este documento descreve os testes já implementados no `/backend` e o padrão
seguido para escrevê-los. A estrutura é inspirada no projeto didático
`api-tarefas-testavel` (go-tests): repositórios são interfaces consumidas
pela Service, o que permite testar as regras de negócio com um Repository
falso, sem PostgreSQL.

## Status atual

- [x] Testes unitários de `domain` (Product, User, Order e tipos compostos)
- [x] Testes unitários de `ProductService`
- [x] Testes unitários de `OrderService`
- [x] Testes unitários de `UserService`
- [x] Testes unitários de `AuthService`
- [x] Testes de Handler (Gin + `httptest`)
- [x] Testes unitários de `security.JWTService`
- [x] Testes unitários de `middleware` (`Authenticate`, `RequireRole`)
- [x] Testes de integração full-stack HTTP com PostgreSQL real cobrindo os
      quatro módulos com lógica de negócio: autenticação, produtos, pedidos
      (criação com lock de estoque, pagamento com propriedade atômica,
      cancelamento com restore de estoque, busca por propriedade) e
      usuários (CRUD administrativo, incluindo o restore-não-reativa-sozinho)
- [x] Testes de integração no nível de Repository (`internal/repository`,
      isolando só a camada SQL, sem Service/Handler/HTTP) para
      `PostgresProductRepository`, `PostgresUserRepository` e
      `PostgresOrderRepository` — este último cobrindo `ErrOrderOwnerNotFound`/
      `ErrOrderOwnerUnavailable`, inalcançáveis pelos testes HTTP

`internal/routes` e `middleware.CORS` ficaram de fora de propósito: são
configuração/wiring declarativo (registro de rota, config estática de uma
lib de terceiros), sem nenhuma lógica condicional nossa para testar — ver
a discussão registrada no histórico do projeto.

## Cobertura por módulo (somente testes unitários)

Medida isolando só a suíte unitária (sem subir Postgres nem rodar a tag
`integration`), usando `-coverpkg=./internal/...` para contar cobertura de
todo o código de produção, não só do pacote de cada teste:

```bash
cd backend
go test ./... -coverprofile=unit.cov -coverpkg=./internal/...
go tool cover -func=unit.cov | tail -1   # total
```

| Módulo (`internal/...`) | Cobertura | Statements cobertos |
|---|---:|---:|
| `domain` | 100,0% | 115/115 |
| `mapper` | 100,0% | 7/7 |
| `handler` | 97,8% | 269/275 |
| `middleware` | 97,2% | 35/36 |
| `service` | 92,0% | 172/187 |
| `security` | 87,5% | 21/24 |
| `config` | 0,0% | 0/29 |
| `database` | 0,0% | 0/7 |
| `repository` | 0,0% | 0/530 |
| `routes` | 0,0% | 0/43 |
| **Total (`./internal/...`)** | **49,4%** | **626/1253** |

`repository` e `routes` ficam em 0% **de propósito** nessa medição: nos
testes unitários, `repository` é sempre substituído por um fake (é
exatamente o ponto de ter a interface) e `routes.Register` só é chamado de
verdade dentro de `backend/integration/setup_test.go` — nenhum teste
unitário sobe um `gin.Engine` real. Ambos saem de 0% para 100%/83,6%
assim que os testes de integração entram na conta (ver a seção "Executar
os testes de integração" abaixo); a tabela acima é só a fatia que **não**
depende de Docker/Postgres para rodar. `config`/`database` continuam em 0%
mesmo somando integração — são wrappers triviais (`godotenv.Load`,
`pgxpool.New`) que nenhum teste chama diretamente.

## Executar

Todos os testes atuais são unitários e não exigem PostgreSQL:

```bash
go test ./internal/domain/... ./internal/service/... ./internal/handler/... ./internal/security/... ./internal/middleware/... -v
```

Para rodar só um arquivo/pacote:

```bash
go test ./internal/domain/... -run TestNewProduct -v
go test ./internal/domain/... -run TestNewUser -v
go test ./internal/domain/... -run TestOrder -v
go test ./internal/service/... -run TestProductService -v
go test ./internal/service/... -run TestOrderService -v
go test ./internal/service/... -run TestUserService -v
go test ./internal/service/... -run TestAuthService -v
go test ./internal/handler/... -run TestProductHandler -v
go test ./internal/handler/... -run TestOrderHandler -v
go test ./internal/handler/... -run TestUserHandler -v
go test ./internal/handler/... -run TestAuthHandler -v
go test ./internal/security/... -run TestJWTService -v
go test ./internal/middleware/... -run TestAuthenticate -v
go test ./internal/middleware/... -run TestRequireRole -v
```

Cobertura:

```bash
go test ./internal/domain/... ./internal/service/... ./internal/handler/... ./internal/security/... ./internal/middleware/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Executar os testes de integração

Diferente de tudo acima, esses testes exigem PostgreSQL real e a tag de
build `integration` — por isso `go test ./...` (sem a tag) nunca os
executa, e `go build`/`go vet` normais nem tentam compilar os pacotes
`integration` e `internal/repository` (a parte com essa tag).

Suba o banco dedicado (porta, database e volume isolados do banco de
desenvolvimento) e aguarde o healthcheck:

```bash
docker compose -f .devcontainer/docker-compose.yml --profile test up -d --wait postgres-test
```

Execute as duas suítes — full-stack HTTP (`integration`) e Repository
isolado (`internal/repository`):

```bash
go test -tags=integration -count=1 -v ./integration/... ./internal/repository/...
```

O endereço pode ser substituído quando necessário (por exemplo, para rodar
contra um Postgres já existente em outra porta):

```bash
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5434/ecommerce_test?sslmode=disable' \
  go test -tags=integration -count=1 -v ./integration/... ./internal/repository/...
```

O helper `requireTestDatabaseURL` (reimplementado em cada um dos dois
pacotes) recusa qualquer URL cujo database não termine em `_test` —
barreira extra contra rodar `TRUNCATE` no banco de desenvolvimento por
engano, mesmo com a variável de ambiente errada. Em `integration`, cada
teste roda `newTestApp(t)`, que:

```text
SETUP:   conecta ao postgres-test → aplica a migration inicial (só na
         primeira vez; checa se a tabela já existe) → TRUNCATE
TESTE:   sobe a aplicação inteira (Repository+Service+Handler+rotas+
         middlewares reais) via httptest.NewServer e faz requisições HTTP
         de verdade, com cookie jar próprio por client
CLEANUP: fecha o pool de conexões (t.Cleanup)
```

Em `internal/repository`, `openTestDatabase(t)` faz o mesmo setup, mas o
teste chama o Repository direto — sem Service, Handler nem HTTP.

Para parar o banco de teste depois (preserva o volume, os dados somem só no
próximo `TRUNCATE`):

```bash
docker compose -f .devcontainer/docker-compose.yml --profile test stop postgres-test
```

## Padrão usado

- **Fake de Repository por Service**: cada arquivo de teste declara um
  `fakeXxxRepository` com um campo `func` por método da interface (`Create`,
  `Update`, `FindByID`, ...). O teste configura só os métodos que o cenário
  usa; os demais ficam `nil` e panicam se chamados por engano — um sinal
  explícito de que a Service tentou fazer algo além do esperado.
- **Table-driven tests + subtests** (`t.Run`): usados tanto para casos de
  validação (título/preço/e-mail inválido) quanto para variações de um mesmo
  fluxo (paginação, papéis de usuário).
- **Contadores de chamada**: quando o importante é provar que o Repository
  _não_ foi chamado (dado inválido, produto já removido, etc.), o fake
  incrementa um contador local e o teste verifica que ficou em zero.
- **`errors.Is`**: todas as comparações de erro usam `errors.Is`, compatível
  com os `errors.Join` usados nas validações de domínio (`domain.NewProduct`,
  `domain.NewUser`).
- **Handler chamado diretamente** (sem `gin.Engine`/rotas): cada teste monta
  um `*gin.Context` isolado com `gin.CreateTestContext` + `httptest`, injeta
  corpo/params/usuário autenticado à mão, e chama o método do Handler. É o
  nível mais barato de teste de Controller — equivalente ao "Controller
  chamado diretamente" do go-tests — e não depende de middlewares reais
  (`Authenticate`/`RequireRole`) nem de JWT.
- **`serve(context, handler.Method)`**: como os testes não passam pelo
  `gin.Engine`, `c.Status(204)` sozinho nunca é gravado no
  `ResponseRecorder` — só `WriteHeaderNow()` (chamada pelo engine de
  roteamento) faz esse flush. O helper `serve` chama o Handler e força esse
  flush logo em seguida.
- **Os quatro Handlers dependem de interface de Service**
  (`service.ProductService`, `service.OrderService`, `service.UserService`,
  `service.AuthService`). Os tipos concretos (`productService`, `orderService`,
  `userService`, `authService`) são não exportados e satisfazem o contrato
  implicitamente — o mesmo padrão usado nos Repositories. `OrderHandler` e
  `AuthHandler` já nasceram assim; `ProductService`/`UserService` eram
  structs exportados até essa consistência ser aplicada.
- **Todo teste de Handler usa uma Service falsa direta** (`fakeProductService`,
  `fakeOrderService`, `fakeUserService`, `fakeAuthService`) — nunca a Service
  real. Isso isola o Handler por completo: o teste cobre só binding de
  JSON/query/params e o mapeamento de erro→status HTTP, sem reexercitar
  validação de domínio (já coberta em `internal/service/*_test.go`). Cada
  fake segue o mesmo formato dos fakes de Repository: um campo `func` por
  método da interface.
- **Testes de domínio não usam nenhum double**: `Product`, `User` e `Order`
  não têm dependências (nem Repository, nem Service), então os testes
  chamam os construtores/métodos diretamente. O padrão recorrente é o par
  "atualização válida muda os campos" / "atualização inválida preserva os
  campos originais" — importante porque `Update()` só chama `setData` depois
  de validar, então um erro nunca deve deixar a entidade parcialmente
  modificada.
- **Testes de integração não usam nenhum double**: é o oposto do resto da
  suíte — a aplicação inteira é montada com as implementações reais
  (`repository.NewPostgresXxxRepository`, os Services, os Handlers, as
  rotas e os middlewares `Authenticate`/`RequireRole`) contra um PostgreSQL
  de verdade, e as requisições atravessam um `httptest.Server` por socket
  HTTP local. O único atalho é criar o usuário admin de teste chamando
  `UserService.Create` direto em Go (sem HTTP), porque o autocadastro
  público só cria `customer` — o mesmo atalho que `cmd/api/seed.go` usa
  para o admin padrão em produção.

## `internal/domain/product_test.go`

Cobre a entidade `Product`: construção, atualização e as regras de estoque.

- `TestNewProduct` — nome/preço/estoque validados juntos via `errors.Join`
  (um subteste confirma que os três erros aparecem simultaneamente com
  `errors.Is`); nome e descrição são normalizados (trim).
- `TestProductUpdate` — atualização válida muda os campos; atualização
  inválida preserva os campos originais (a entidade nunca fica
  parcialmente modificada).
- `TestProductHasStock` — quantidade precisa ser positiva e não pode
  exceder o estoque atual (tabela cobrindo os limites).
- `TestProductReduceStock` / `TestProductRestoreStock` — quantidade inválida
  e estoque insuficiente são recusados sem alterar o estoque; a operação
  válida soma/subtrai corretamente.
- `TestRestoreProduct` — reidratação a partir do banco preenche id,
  ativação, remoção e timestamps, e continua validando os mesmos dados que
  `NewProduct`.

## `internal/domain/user_test.go`

Cobre a entidade `User`: construção, atualização e o papel (role).

- `TestNewUser` — nome/e-mail/papel validados juntos via `errors.Join`
  (regex de e-mail exercitada com várias entradas malformadas); nome e
  e-mail são normalizados (trim).
- `TestUserUpdate` — mesmo par "válido muda / inválido preserva" de Product.
- `TestRestoreUser` — reidratação a partir do banco, mesmo formato de
  `TestRestoreProduct`.
- `TestUserRoleIsValid` — só `customer` e `admin` são papéis reconhecidos.

## `internal/domain/order_test.go`

Cobre `Order`, `OrderItem` e `OrderStatus` — sem Repository, direto nos
métodos que a Service delega (`Pay`, `Cancel`).

- `TestOrderCanPayAndCanCancel` — só um pedido `PENDING` pode ser pago ou
  cancelado.
- `TestOrderPay` / `TestOrderCancel` — a transição válida define `Status`,
  `PaidAt` ou `CanceledAt` e `UpdatedAt`; tentar pagar ou cancelar um pedido
  em outro status recusa **sem mutar o pedido** (o `CanPay`/`CanCancel` é
  checado antes de qualquer atribuição).
- `TestOrderItemSubtotal` — quantidade × preço unitário, incluindo
  quantidade zero.
- `TestOrderStatusIsValid` — só os três status conhecidos são aceitos.

## `internal/domain/shared_test.go`

Cobre os tipos compostos por `Product`/`User` via embedding
(`Activatable`, `SoftDelete`, `Timestamps`) e `UserAuthentication.IsDeleted`
(usado por `AuthService.Login`, que não compõe `SoftDelete` porque trafega
campos diferentes do repository de autenticação).

- `TestActivatable` / `TestSoftDelete` — os dois construtores (`New*` e
  `New*From`) e as transições de estado (`Activate`/`Deactivate`,
  `Delete`/`Restore`).
- `TestTimestamps` — `NewTimestamps` cria `createdAt`/`updatedAt` iguais;
  `Touch` avança só `updatedAt`.
- `TestUserAuthenticationIsDeleted` — mesma regra de `deletedAt != nil`
  usada por `SoftDelete`, testada isoladamente por não reusar o tipo.

## `internal/service/product_test.go`

Cobre `ProductService`, incluindo soft delete e ativação/desativação.

- `TestProductServiceCreate` — delega ao Repository apenas quando o Model
  aceita os dados; nome vazio e preço inválido são rejeitados antes disso.
- `TestProductServiceUpdate` — atualização normal; recusa produto já
  removido (`ErrProductAlreadyDeleted`) sem persistir; recusa dados
  inválidos sem persistir; propaga erro quando o produto não existe.
- `TestProductServiceSearch` — paginação padrão (`page=1`, `pageSize=20`),
  teto de `pageSize=100` e cálculo de `TotalPages`.
- `TestMapDeletionFilter` — tradução do `dto.DeletionState` para
  `repository.DeletionFilter`.
- `TestProductServiceDelegatesOperations` — repasse simples de `FindByID`,
  `DeleteByID`, `RestoreByID`, `ActivateByID`, `DeactivateByID`.

## `internal/service/order_test.go`

Cobre `OrderService`, com foco na regra de propriedade do pedido (o
`customer_id` nunca vem do corpo da requisição, só do usuário autenticado).

- `TestConsolidateOrderItems` — função pura não exportada: soma quantidades
  de `product_id` repetidos preservando a ordem da primeira ocorrência, e
  detecta overflow ao somar (`math.MaxInt`).
- `TestOrderServiceCreate` — consolidação delegada ao Repository com o
  `ownerID` vindo do usuário autenticado; rejeita pedido sem itens; rejeita
  overflow sem persistir; propaga erro do Repository.
- `TestOrderServiceSearch` — customer recebe `filter.CustomerID` preenchido
  com o próprio id; admin não recebe filtro de propriedade; mesma paginação
  padrão/teto usada em Product.
- `TestOrderServiceFindByID` — matriz de autorização: admin acessa qualquer
  pedido, customer só o próprio (`ErrOrderAccessDenied` caso contrário),
  propagação de `ErrOrderNotFound`.
- `TestOrderServicePayByID` / `TestOrderServiceCancelByID` — confirmam que
  `id`, `ownerID`/`requesterID` e o booleano `isAdmin` (derivado do papel
  autenticado) chegam corretos ao Repository, já que a autorização real
  acontece atomicamente no SQL, não na Service.

## `internal/service/user_test.go`

Cobre `UserService`, incluindo o hashing de senha com bcrypt.

- `TestUserServiceRegister` — autocadastro público sempre cria
  `RoleCustomer`; a senha chega ao Repository em hash (nunca texto puro,
  verificado com `bcrypt.CompareHashAndPassword`); dados inválidos não
  persistem.
- `TestUserServiceCreate` — criação administrativa: papel omitido usa
  `customer` como padrão; papel explícito é respeitado; papel inválido é
  rejeitado sem persistir.
- `TestUserServiceUpdate` — atualização sem trocar senha envia
  `passwordHash=nil`; quando a senha é informada, é re-hasheada
  corretamente; usuário removido é recusado (`ErrUserAlreadyDeleted`); dados
  inválidos (e-mail malformado) não persistem; erro de `FindByID` é
  propagado.
- `TestUserServiceSearch` / `TestMapUserDeletionFilter` — mesmo padrão de
  paginação e mapeamento de filtro usado em Product.
- `TestUserServiceDelegatesOperations` — repasse simples de `FindByID`,
  `DeleteByID`, `RestoreByID`, `ActivateByID`, `DeactivateByID`.

## `internal/service/auth_test.go`

Cobre `AuthService`, que depende de duas interfaces (`repository.AuthRepository`
e `security.JWTService`) — cada uma com seu próprio fake.

- `TestAuthServiceLogin`:
  - autentica, normaliza o e-mail (trim) antes da busca, gera token e
    atualiza o último login;
  - recusa senha incorreta sem gerar token nem atualizar login;
  - recusa usuário removido mesmo com senha correta — mesmo erro
    (`ErrInvalidCredentials`) do caso de senha errada, para não vazar qual
    foi o motivo real;
  - recusa usuário inativo (`ErrUserInactive`, erro distinto do anterior);
  - propaga erro quando o Repository não encontra o e-mail;
  - propaga erro ao gerar token, sem chegar a atualizar o último login;
  - propaga erro ao atualizar o último login (retorno final zerado).
- `TestAuthServiceFindAuthenticatedUserByID` — repasse direto ao Repository,
  usado pelo middleware de autenticação a cada requisição.

## `internal/handler/testhelpers_test.go`

Helpers compartilhados por todos os testes de Handler: `TestMain` (liga
`gin.TestMode`), `newTestContext`/`newJSONTestContext` (monta `*gin.Context`

- `httptest.ResponseRecorder` sem depender de rotas), `setIDParam` (simula
  `:id`), `setAuthenticatedUser` (simula o trabalho do middleware
  `Authenticate`, colocando o usuário no contexto), `serve` (chama o Handler e
  força o flush do status/headers) e `decodeJSONBody` (decodifica a resposta).

## `internal/handler/product_test.go`

Cobre `ProductHandler` usando `fakeProductService` (Service falsa direta).

- `TestProductHandlerCreate` — 201 no caminho feliz; 400 para corpo
  malformado sem chamar a Service; 400 quando a Service devolve qualquer
  erro — esse endpoint não tem `switch`, qualquer falha vira 400.
- `TestProductHandlerUpdate` — table-driven cobrindo o `switch` inteiro: 200,
  404 não encontrado, 409 já removido, 400 em cada variante de dado inválido
  (`ErrInvalidProductName`/`Description`/`Price`/`Stock`) e 500 genérico —
  mais o 400 de corpo malformado, testado à parte por não chamar a Service.
- `TestProductHandlerFindByID` / `TestProductHandlerSearch` — rota pública:
  sucesso, 404/400 e 500 para erro inesperado da Service.
- `TestProductHandlerDeleteByID` / `RestoreByID` / `ActivateByID` /
  `DeactivateByID` — table-driven cobrindo todo o mapeamento de status desses
  endpoints administrativos (204, 404, 409 em cada variação, 500).

## `internal/handler/order_test.go`

Cobre `OrderHandler` usando `fakeOrderService` (Service falsa direta).

- `TestOrderHandlerCreate` — 401 sem usuário autenticado (o dono do pedido
  só existe no contexto); 400 corpo inválido; repasse do usuário autenticado
  à Service; mapeamento completo de erro (`ErrOrderMustHaveItems`/
  `ErrInvalidOrderQuantity`→400, `ErrProductNotFound`→404,
  `ErrProductUnavailable`/`ErrInsufficientStock`→409, genérico→500).
- `TestOrderHandlerSearch` — 401, 400 (`page` inválido), repasse do usuário, 500.
- `TestOrderHandlerFindByID` — 401, 400 (id fora do formato UUID), 200, e o
  ponto mais importante: `ErrOrderNotFound` e `ErrOrderAccessDenied` geram a
  **mesma** resposta 404, para não vazar se o pedido existe e pertence a
  outro usuário.
- `TestOrderHandlerPayByID` / `TestOrderHandlerCancelByID` — mesmo formato,
  com o conflito específico de cada operação (`ErrOrderCannotBePaid` /
  `ErrOrderCannotBeCanceled` → 409).

## `internal/handler/user_test.go`

Cobre `UserHandler` usando `fakeUserService` (Service falsa direta).

- `TestUserHandlerCreate` — table-driven cobrindo o `switch` inteiro: 201,
  409 e-mail duplicado, 400 em cada variante de dado inválido
  (`ErrInvalidUserName`/`Email`/`Role`) e 500 genérico — mais o 400 de corpo
  malformado, testado à parte por não chamar a Service.
- `TestUserHandlerUpdate` — mesmo formato: 200, 404, 409 já removido, 409
  e-mail já usado por outro usuário, 400 em cada variante de dado inválido e
  500 genérico.
- `TestUserHandlerFindByID` / `TestUserHandlerSearch` — sucesso, 404/400 e 500.
- `TestUserHandlerDeleteByID` / `RestoreByID` / `ActivateByID` /
  `DeactivateByID` — table-driven, mesmo formato usado em Product.

## `internal/handler/auth_test.go`

Cobre `AuthHandler` usando `fakeAuthService` para Login/Me e `fakeUserService`
(o mesmo tipo declarado em `user_test.go`, reaproveitado por estarem no
mesmo pacote) para Register.

- `TestAuthHandlerRegister` — table-driven: 201, 409 e-mail duplicado, 400 em
  cada variante de dado inválido e 500 genérico — mais o 400 de corpo
  malformado sem chamar a Service. Mesmo mapeamento de `UserHandler.Create`,
  já que Register chama `UserService.Register`.
- `TestAuthHandlerLogin` — sucesso define o cookie `access_token` como
  `HttpOnly` com o token/expiração devolvidos pela Service; 400 corpo
  inválido; 401 credenciais inválidas; 403 usuário inativo; 500 genérico.
- `TestAuthHandlerLogout` — sempre 204 e sempre limpa o cookie (`MaxAge`
  negativo), mesmo sem sessão válida — rota pública.
- `TestAuthHandlerMe` — 200 com o usuário já colocado no contexto pelo
  middleware; 401 quando não há usuário autenticado.

## `internal/security/jwt_test.go`

Cobre `jwtService` (implementação real de `security.JWTService` usada por
`AuthService`/`middleware.Authenticate`) — sem fake, é o único componente de
segurança que vale testar contra a biblioteca `golang-jwt/jwt/v5` de
verdade, já que nos testes de `AuthService`/Handler ele sempre foi
substituído por `fakeJWTService`.

- `TestJWTServiceGenerateAndValidateAccessToken` — ciclo completo: o token
  gerado é aceito de volta e devolve o mesmo `userID`; a expiração respeita
  o TTL configurado (tolerância de 2s).
- `TestJWTServiceValidateAccessToken` — table-driven cobrindo formas de
  token inválido, cada uma construída fora do serviço via
  `mustSignToken`/`jwt.NewWithClaims` para produzir combinações de claims
  que `GenerateAccessToken` nunca geraria: token vazio/malformado, assinado
  com outro segredo, `issuer`/`audience` divergentes, sem `exp`, `subject`
  vazio ou fora do formato UUID, e **algoritmo `none`** (ataque clássico de
  confusão de algoritmo — a lib rejeita porque `WithValidMethods` restringe
  a HS256). Todos esses casos retornam `domain.ErrInvalidToken`.
- `TestJWTServiceValidateAccessTokenExpired` — um token expirado retorna
  `domain.ErrExpiredToken`, erro distinto dos demais — é essa distinção que
  `middleware.Authenticate` usa para escolher a mensagem de erro (401
  "expirado" vs. "inválido").

## `internal/middleware/authorize_test.go`

Cobre `RequireRole` — lógica pura, sem dependências, chamada diretamente
sobre um `*gin.Context` isolado (sem `serve`, pois esse middleware nunca usa
`c.Status` sozinho: cada erro escreve JSON via `AbortWithStatusJSON`, que já
dispara o flush).

- `TestRequireRole` — sem usuário autenticado retorna 401 (tratado como
  falha de _autenticação_, não de permissão — `RequireRole` pressupõe que
  `Authenticate` já rodou antes); papel fora da lista permitida retorna 403;
  papel permitido segue adiante sem abortar; aceita qualquer um dos papéis
  informados (`customer` ou `admin`).

## `internal/middleware/auth_test.go`

Cobre `Authenticate` e `GetAuthenticatedUser`, usando `fakeJWTService`
(interface `security.JWTService`) e `fakeAuthRepository` (interface
`repository.AuthRepository`) — os mesmos dois pontos de extensão já
fakeados nos testes de `AuthService`, agora dublados aqui para isolar o
middleware.

- `TestGetAuthenticatedUser` — contrato de "nunca causa panic": contexto
  vazio ou com valor de tipo inesperado retornam `ok=false` em vez de um
  type assertion sem verificação.
- `TestAuthenticate` — as três etapas em sequência, cada uma com sua própria
  mensagem de 401 (verificada no corpo da resposta, não só no status): sem
  cookie / cookie vazio (sem chegar a validar o token — contador de
  chamadas comprova), token inválido, token expirado (mensagem distinta do
  anterior), usuário não encontrado no Repository. O último caso confirma
  que só a combinação das três etapas injeta o `*domain.AuthenticatedUser`
  no contexto e libera a requisição (`context.IsAborted() == false`).

## `.devcontainer/docker-compose.yml` — serviço `postgres-test`

Banco dedicado aos testes de integração, atrás do profile `test` (não sobe
com `docker compose up` normal nem com o Dev Container). Porta (`5434` por
padrão), database (`ecommerce_test`) e volume (`postgres-test-data`) são
isolados do banco de desenvolvimento (`db`, porta `5432`, database
`postgres`) — os nomes das variáveis de ambiente também são distintos
(`POSTGRES_TEST_*`) para não colidir com as já usadas por `app`/`db` no
mesmo arquivo.

## `backend/integration/setup_test.go`

Helpers compartilhados por toda a suíte de integração, no mesmo espírito do
`integration/setup_test.go` do go-tests:

- `openIntegrationDatabase` — conexão fail-closed (`requireTestDatabaseURL`
  recusa qualquer database que não termine em `_test`) e `ensureSchema`,
  que só aplica a migration inicial se a tabela `products` ainda não
  existir (as migrations não são idempotentes como o seed; reaplicar um
  `CREATE TABLE` quebraria a segunda execução).
- `resetDatabase` — um único `TRUNCATE` cobrindo as cinco tabelas de dados
  (exigência do Postgres para tabelas ligadas por FK), preservando o schema.
- `newTestApp` — monta a aplicação inteira (todas as `PostgresXxxRepository`,
  Services, Handlers, rotas e middlewares) exatamente como `cmd/api/main.go`,
  publica em `httptest.NewServer`, e expõe `UserService` para os testes
  seedarem um admin sem passar pelo HTTP.
- `newClient` — `*http.Client` com cookie jar próprio por chamada: depois de
  um login bem-sucedido, as requisições seguintes desse client já carregam o
  cookie de sessão automaticamente.
- `performRequest`/`decodeInto` — mesmo padrão de helpers HTTP do
  `integration/http_test.go` do go-tests.
- `registerAndLoginCustomer`/`createAndLoginAdmin` — atalhos usados por
  praticamente todo teste: registram (ou criam, no caso do admin) e já
  devolvem um client autenticado. `createAndLoginAdmin` existe porque o
  autocadastro público só cria `customer` — criar um admin de teste exige
  chamar `UserService.Create` direto, fora do HTTP.
- `createTestProduct` — cria um produto via HTTP usando um client de admin
  já autenticado; usado pelos testes de pedido para ter um produto com
  estoque conhecido.
- `apiErrorResponse` — decodifica `gin.H{"error": ...}`, o formato usado por
  todos os Handlers de erro, para os poucos testes que precisam conferir a
  mensagem, não só o status.

## `backend/integration/auth_test.go`

- `TestAuthRegisterLoginMeLogout` — fluxo completo contra Postgres real:
  autocadastro (sempre `customer`, nunca configurável pelo cliente), `/me`
  sem sessão (401), login com senha errada (401), login válido (cookie
  `HttpOnly` de verdade, devolvido automaticamente pelo cookie jar), `/me`
  autenticado, logout, e `/me` voltando a 401 depois. bcrypt e JWT rodam de
  verdade — nada aqui é fake.
- `TestAuthRegisterDuplicateEmail` — a constraint única do banco
  (`users.email CITEXT UNIQUE`) chega como `domain.ErrUserEmailAlreadyExists`
  e vira 409 na resposta.

## `backend/integration/product_test.go`

- `TestProductAdminFlow` — gestão de produtos de ponta a ponta, com foco na
  autorização real (não fakeada): visitante anônimo não pode criar (401);
  cliente autenticado não pode criar (403 — `RequireRole` de verdade);
  admin cria, lê (rota pública), atualiza e remove; remover de novo um
  produto já removido dá 409; buscar por ID um produto removido continua
  retornando 200 (`FindByID` não filtra soft-deleted, só `Search` filtra);
  ID inexistente dá 404.

## `backend/integration/order_test.go`

O mais valioso dos três: exercita transação, lock de linha (`FOR UPDATE`) e
baixa/reposição de estoque no `PostgresOrderRepository`, nada disso
verificável com Repository falso.

- `TestOrderCreateFlow` — anônimo não pode criar (401); produto inexistente
  (404); **estoque insuficiente devolve 409 e não altera o estoque**
  (prova de que a transação é revertida antes do commit); produto inativo
  (409); criação válida decrementa o estoque atomicamente e calcula
  `total_amount` corretamente; dono lê o próprio pedido, outro cliente
  recebe 404 (não 403 — para não vazar a existência do pedido).
- `TestOrderPayOwnership` — a regra mais sutil do módulo:
  **`PayByID` exige atomicamente id + ownerID + status PENDING na mesma
  instrução SQL, sem exceção para admin.** O teste prova isso na prática:
  outro cliente não paga (404), e um **admin também não** (404) — só o
  dono. Paga uma vez com sucesso; pagar de novo (409); cancelar um pedido
  já pago (409).
- `TestOrderCancelRestoresStockAndAllowsAdmin` — a assimetria complementar:
  `CancelByID` recebe `isAdmin` explicitamente, então **admin pode cancelar
  o pedido de qualquer cliente** (diferente de pagar). Cancelar um pedido
  `PENDING` restaura a quantidade ao estoque do produto (verificado
  lendo o produto antes/durante/depois via a rota pública).
- `TestOrderSearchOwnership` — customer só vê e só conta os próprios
  pedidos; admin vê e conta todos — a mesma regra já teria sido "provada"
  nos testes de `OrderService` com Repository falso, mas aqui é a
  cláusula `WHERE customer_id = $1` de verdade que está sendo exercitada.

## `backend/integration/user_test.go`

Diferente de `Product`, todo o grupo `/users` exige `admin` — não existe
rota pública de leitura aqui (`RegisterUserRoutes` aplica
`authenticate+requireAdmin` ao grupo inteiro, sem exceção).

- `TestUserAdminFlow` — anônimo não pode nem listar (401); cliente
  autenticado não pode criar (403); admin cria, busca por ID, filtra por
  nome (`ILIKE` real via query `?name=`), atualiza, remove; e-mail
  duplicado dá 409; usuário inexistente dá 404. A parte mais interessante é
  a sequência remove→restaura→ativa: **`RestoreByID` tira o `deleted_at`
  mas devolve o usuário inativo** (`active` continua `FALSE` — não
  reativa sozinho, é preciso um `ActivateByID` explícito depois), e o
  teste verifica isso lendo o usuário via GET logo após restaurar. Por
  fim, prova que ativar um usuário removido reporta especificamente
  "usuário já está removido" (não "já está ativo") — a ordem das
  checagens em `ActivateByID` importa, e só dá pra ver isso na mensagem
  real da resposta, não só no status 409.

## `internal/repository/testhelpers_test.go`

O nível mais baixo da suíte de integração: os três arquivos de teste deste
pacote chamam o `PostgresXxxRepository` direto contra o Postgres de teste,
sem Service, Handler, rotas nem servidor HTTP no meio — no formato de
`internal/repository/task_integration_test.go` do go-tests. Este arquivo
reimplementa a própria conexão/schema/reset (`openTestDatabase`,
`ensureSchema`, `resetTables`, `requireTestDatabaseURL`, `projectFile`) em
vez de reaproveitar `backend/integration/setup_test.go`: são pacotes Go
diferentes (`repository` vs. `integration_test`), e importar um do outro
criaria uma dependência de teste desnecessária entre pacotes de produção.
Também define `mustCreateProduct`/`mustCreateUser` (reaproveitados pelos
três arquivos de teste, inclusive por `order_integration_test.go` para
preparar dono e produtos) e `boolPtr`/`floatPtr`/`sameSet`, usados pelos
testes de `Search`.

## `internal/repository/product_integration_test.go`

- `TestPostgresProductRepositoryCreateAndFindByID` — round trip completo:
  `Create` grava e devolve o produto com ID gerado pelo Postgres
  (`uuidv7()`), `FindByID` lê de volta os mesmos valores; ID inexistente dá
  `ErrProductNotFound`.
- `TestPostgresProductRepositoryUpdate` — os dois caminhos de erro que só a
  query real distingue: produto inexistente (nenhuma linha, em nenhum
  estado) vs. produto removido (a linha existe, mas o `UPDATE` tem
  `AND deleted_at IS NULL` — o Repository só sabe qual dos dois aconteceu
  consultando o estado depois que o `UPDATE` afeta zero linhas).
- `TestPostgresProductRepositoryDeleteRestoreActivateDeactivate` — as quatro
  transições de estado e o mapeamento de erro de cada uma (já removido,
  não removido, já ativo, já inativo, removido ao tentar ativar/desativar,
  inexistente) — a mesma lógica de "consulta o estado só quando o UPDATE
  não afeta nada" usada em `Update`.
- `TestPostgresProductRepositorySearch` — o motivo principal de ter essa
  camada de teste: prova que `buildProductFilters` monta o `WHERE` certo
  para nome (`ILIKE` parcial e case-insensitive), categoria, `active`,
  faixa de preço, estado de exclusão (`not_deleted`/`deleted`/`all`) e
  combinações desses filtros — contra dados reais, sem precisar de
  admin/HTTP para isso.
- `TestPostgresProductRepositorySearchPagination` — `Limit`/`Offset`
  isolados do filtro: percorre todas as páginas e confirma que não há
  sobreposição nem lacuna.

## `internal/repository/user_integration_test.go`

Mesmo formato de `product_integration_test.go`, adaptado às
particularidades de `User`.

- `TestPostgresUserRepositoryCreateAndFindByID` — round trip (`Create`
  grava usuário **e** credencial em duas tabelas, numa transação);
  e-mail duplicado retorna `ErrUserEmailAlreadyExists` — a mesma
  constraint única (`users.email CITEXT UNIQUE`) já vista no nível HTTP,
  aqui exercitada diretamente.
- `TestPostgresUserRepositoryUpdate` — inexistente, removido, e um caso que
  `product_integration_test.go` não tinha: **e-mail já usado por outro
  usuário durante um `Update`** (não só na criação) também retorna
  `ErrUserEmailAlreadyExists`.
- `TestPostgresUserRepositoryDeleteRestoreActivateDeactivate` — mesma
  bateria de transições de Product, e confirma de novo (agora no nível
  Repository, não HTTP) que `RestoreByID` devolve o usuário sem
  `deleted_at` mas ainda inativo.
- `TestPostgresUserRepositorySearch` — `buildUserFilters` para nome, e-mail,
  papel, `active` e estado de exclusão, sozinhos e combinados.
- `TestPostgresUserRepositorySearchPagination` — igual à de Product.

## `internal/repository/order_integration_test.go`

O mais valioso dos três, porque foca exatamente no que os testes HTTP
(`backend/integration`) **não conseguem** alcançar.

- `TestPostgresOrderRepositoryCreate` — **`ErrOrderOwnerNotFound` e
  `ErrOrderOwnerUnavailable` (proprietário inativo ou removido) nunca
  tinham sido exercitados em nenhum teste até agora**: no nível HTTP, o
  middleware `Authenticate` já rejeita (401) um usuário inativo/removido
  antes da requisição chegar ao Service, então esse trecho de
  `validateOwner` — que o próprio comentário do código descreve como
  cobertura para "a janela entre a validação do JWT e a efetivação do
  pedido" — só é alcançável chamando o Repository direto. O teste também
  cria um pedido com **dois produtos diferentes** numa única chamada
  (soma do total e dois decrementos de estoque na mesma transação), caso
  que os testes HTTP nunca montaram (sempre usaram um item por pedido).
- `TestPostgresOrderRepositoryFindByIDAndPayByID` — leitura com itens
  anexados, e a exigência atômica de `PayByID` (outro dono → acesso
  negado; pagar de novo o mesmo pedido, já pago, pelo próprio dono →
  `ErrOrderCannotBePaid`).
- `TestPostgresOrderRepositoryCancelByID` — a assimetria com `PayByID`
  (`CancelByID` recebe `isAdmin` explicitamente) e a restauração exata da
  quantidade ao estoque.
- `TestPostgresOrderRepositorySearch` — `buildOrderFilters` com e sem
  `CustomerID`.

---
