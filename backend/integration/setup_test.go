//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"ecommerce/internal/config"
	"ecommerce/internal/domain"
	"ecommerce/internal/dto"
	"ecommerce/internal/handler"
	"ecommerce/internal/middleware"
	"ecommerce/internal/repository"
	"ecommerce/internal/routes"
	"ecommerce/internal/security"
	"ecommerce/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultIntegrationDatabaseURL = "postgres://postgres:postgres@localhost:5434/ecommerce_test?sslmode=disable"
	integrationJWTSecret          = "integration-test-secret-nao-usar-em-producao-32bytes"
	integrationCookieName         = "access_token"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// testApp reúne o servidor HTTP real — Repository, Service, Handler, rotas
// e middlewares, exatamente como cmd/api/main.go monta — publicado em uma
// porta local, mais o UserService, exposto porque o autocadastro público
// (POST /auth/register) só cria "customer": preparar um admin para os
// testes exige chamar a Service diretamente, sem passar pelo HTTP.
type testApp struct {
	Server      *httptest.Server
	DB          *pgxpool.Pool
	UserService service.UserService
}

// newTestApp monta a aplicação inteira contra o PostgreSQL de integração e
// garante que cada teste comece com o banco limpo.
func newTestApp(t *testing.T) *testApp {
	t.Helper()

	db := openIntegrationDatabase(t)
	resetDatabase(t, db)

	productRepository := repository.NewPostgresProductRepository(db)
	productService := service.NewProductService(productRepository)
	productHandler := handler.NewProductHandler(productService)

	userRepository := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	jwtService := security.NewJWTService(integrationJWTSecret, "ecommerce-api", "ecommerce-clients", 15*time.Minute)

	authRepository := repository.NewPostgresAuthRepository(db)
	authService := service.NewAuthService(authRepository, jwtService)
	authHandler := handler.NewAuthHandler(authService, userService, handler.CookieConfig{
		Name:     integrationCookieName,
		SameSite: http.SameSiteLaxMode,
	})

	orderRepository := repository.NewPostgresOrderRepository(db)
	orderService := service.NewOrderService(orderRepository)
	orderHandler := handler.NewOrderHandler(orderService)

	healthHandler := handler.NewHealthHandler(&config.Config{
		AppName:    "ecommerce-integration",
		AppVersion: "test",
	})

	authenticateMiddleware := middleware.Authenticate(jwtService, authRepository, integrationCookieName)
	requireAdminMiddleware := middleware.RequireRole(domain.RoleAdmin)
	requireCustomerOrAdminMiddleware := middleware.RequireRole(domain.RoleCustomer, domain.RoleAdmin)

	router := gin.New()
	routes.Register(router, routes.Handlers{
		Product: productHandler,
		User:    userHandler,
		Auth:    authHandler,
		Order:   orderHandler,
		Health:  healthHandler,
	}, routes.Middlewares{
		Authenticate:           authenticateMiddleware,
		RequireAdmin:           requireAdminMiddleware,
		RequireCustomerOrAdmin: requireCustomerOrAdminMiddleware,
	})

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &testApp{Server: server, DB: db, UserService: userService}
}

// newClient devolve um *http.Client com cookie jar próprio — depois de um
// POST /auth/login bem-sucedido, as próximas chamadas desse client já
// carregam o cookie de sessão automaticamente, do mesmo jeito que um
// navegador faria.
func (app *testApp) newClient(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("criar cookie jar: %v", err)
	}

	return &http.Client{Jar: jar, Timeout: 5 * time.Second}
}

// registerAndLoginCustomer cria um usuário via autocadastro público (sempre
// "customer") e devolve um client já autenticado — usado sempre que um
// teste precisa de um cliente comum, sem privilégios administrativos.
func registerAndLoginCustomer(t *testing.T, app *testApp, name, email, password string) *http.Client {
	t.Helper()

	if _, err := app.UserService.Register(context.Background(), dto.RegisterRequest{Name: name, Email: email, Password: password}); err != nil {
		t.Fatalf("registrar cliente %q: %v", email, err)
	}

	client := app.newClient(t)
	loginResult := performRequest(t, client, http.MethodPost, app.Server.URL+"/api/v1/auth/login", dto.LoginRequest{
		Email:    email,
		Password: password,
	})
	if loginResult.status != http.StatusOK {
		t.Fatalf("login do cliente %q: status = %d; body = %s", email, loginResult.status, loginResult.body)
	}

	return client
}

// createAndLoginAdmin cria um usuário "admin" chamando UserService.Create
// diretamente (o autocadastro público só cria "customer") e devolve um
// client já autenticado.
func createAndLoginAdmin(t *testing.T, app *testApp, name, email, password string) *http.Client {
	t.Helper()

	role := "admin"
	if _, err := app.UserService.Create(context.Background(), dto.CreateUserRequest{Name: name, Email: email, Password: password, Role: &role}); err != nil {
		t.Fatalf("preparar admin %q: %v", email, err)
	}

	client := app.newClient(t)
	loginResult := performRequest(t, client, http.MethodPost, app.Server.URL+"/api/v1/auth/login", dto.LoginRequest{
		Email:    email,
		Password: password,
	})
	if loginResult.status != http.StatusOK {
		t.Fatalf("login do admin %q: status = %d; body = %s", email, loginResult.status, loginResult.body)
	}

	return client
}

// createTestProduct cria um produto com o estoque informado usando um
// client de admin já autenticado, e devolve o produto criado.
func createTestProduct(t *testing.T, adminClient *http.Client, serverURL, name string, price float64, stock int) dto.ProductResponse {
	t.Helper()

	result := performRequest(t, adminClient, http.MethodPost, serverURL+"/api/v1/products", dto.ProductRequest{
		Name:        name,
		Description: "produto de teste de integração",
		Price:       price,
		Stock:       stock,
	})
	if result.status != http.StatusCreated {
		t.Fatalf("criar produto de teste: status = %d; body = %s", result.status, result.body)
	}

	return decodeInto[dto.ProductResponse](t, result)
}

// openIntegrationDatabase prepara a fronteira real de persistência usada
// pelos testes. Exige um database terminado em _test, abre o pool, confirma
// que o PostgreSQL dedicado responde e garante que o schema existe.
func openIntegrationDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultIntegrationDatabaseURL
	}
	databaseURL = requireTestDatabaseURL(t, databaseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("abrir PostgreSQL de integração: %v", err)
	}
	t.Cleanup(db.Close)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("conectar ao PostgreSQL de integração: %v", err)
	}

	ensureSchema(t, db)

	return db
}

// ensureSchema aplica a migration inicial só na primeira vez (o schema
// persiste no volume do container entre execuções): ao contrário do seed,
// as migrations não são idempotentes — reaplicar um CREATE TABLE falharia.
func ensureSchema(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	if err := db.QueryRow(ctx, `SELECT to_regclass('public.products') IS NOT NULL`).Scan(&exists); err != nil {
		t.Fatalf("verificar schema de integração: %v", err)
	}
	if exists {
		return
	}

	script, err := os.ReadFile(projectFile(t, "database", "migrations", "000001_initial_schema.up.sql"))
	if err != nil {
		t.Fatalf("ler migration inicial: %v", err)
	}

	if _, err := db.Exec(ctx, string(script)); err != nil {
		t.Fatalf("aplicar migration inicial: %v", err)
	}
}

// resetDatabase esvazia todas as tabelas de dados do database exclusivo de
// integração, sempre nesta única instrução (Postgres exige que tabelas
// ligadas por FK sejam truncadas juntas). O schema é preservado.
func resetDatabase(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const truncate = `TRUNCATE TABLE order_items, orders, user_password_credentials, users, products RESTART IDENTITY CASCADE`
	if _, err := db.Exec(ctx, truncate); err != nil {
		t.Fatalf("limpar dados de integração: %v", err)
	}
}

// requireTestDatabaseURL impede que setup, migration ou TRUNCATE sejam
// executados por engano no database da aplicação. Além do container
// separado, o nome terminado em _test funciona como uma segunda barreira de
// segurança.
func requireTestDatabaseURL(t *testing.T, databaseURL string) string {
	t.Helper()

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("interpretar TEST_DATABASE_URL: %v", err)
	}
	databaseName := strings.TrimPrefix(parsedURL.Path, "/")
	if databaseName == "" || !strings.HasSuffix(databaseName, "_test") {
		t.Fatalf("TEST_DATABASE_URL deve apontar para um database terminado em _test; recebido %q", databaseName)
	}
	return parsedURL.String()
}

// projectFile monta um caminho absoluto a partir da raiz do módulo Go
// (backend/). A posição deste arquivo é usada como referência, portanto os
// testes encontram as migrations mesmo quando o comando é iniciado em outro
// diretório.
func projectFile(t *testing.T, parts ...string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("localizar diretório do projeto")
	}
	projectRoot := filepath.Dir(filepath.Dir(currentFile))
	return filepath.Join(append([]string{projectRoot}, parts...)...)
}

// httpResult concentra o resultado de uma chamada HTTP: status, Content-Type
// e o corpo já lido, para os testes decodificarem só o que precisam.
type httpResult struct {
	status      int
	contentType string
	body        []byte
}

// apiErrorResponse decodifica o formato usado por todos os Handlers para
// respostas de erro (gin.H{"error": ...}), quando um teste precisa
// verificar a mensagem, não só o status.
type apiErrorResponse struct {
	Error string `json:"error"`
}

// performRequest concentra o trabalho repetido de um cliente HTTP: serializa
// o payload (quando houver), define Content-Type, executa a chamada e lê o
// corpo da resposta.
func performRequest(t *testing.T, client *http.Client, method, target string, payload any) httpResult {
	t.Helper()

	var body io.Reader = http.NoBody
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("serializar corpo da requisição: %v", err)
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(t.Context(), method, target, body)
	if err != nil {
		t.Fatalf("criar request %s %s: %v", method, target, err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("executar request %s %s: %v", method, target, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ler response %s %s: %v", method, target, err)
	}

	return httpResult{
		status:      response.StatusCode,
		contentType: response.Header.Get("Content-Type"),
		body:        responseBody,
	}
}

// decodeInto transforma o corpo recebido no tipo esperado pelo cenário. O
// tipo genérico permite reutilizar o helper para qualquer DTO de resposta
// sem perder a verificação estática.
func decodeInto[T any](t *testing.T, result httpResult) T {
	t.Helper()

	var decoded T
	if err := json.Unmarshal(result.body, &decoded); err != nil {
		t.Fatalf("decodificar JSON %q: %v", result.body, err)
	}

	return decoded
}
