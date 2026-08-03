//go:build integration

package repository

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultRepositoryTestDatabaseURL = "postgres://postgres:postgres@localhost:5434/ecommerce_test?sslmode=disable"

// openTestDatabase prepara a fronteira real de persistência usada pelos
// testes deste pacote. É mais barata que backend/integration/setup_test.go
// porque não monta Service, Handler nem servidor HTTP — só o Repository
// direto contra o PostgreSQL de teste, o nível de isolamento mais baixo da
// suíte de integração.
func openTestDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultRepositoryTestDatabaseURL
	}
	databaseURL = requireTestDatabaseURL(t, databaseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("abrir PostgreSQL de teste: %v", err)
	}
	t.Cleanup(db.Close)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("conectar ao PostgreSQL de teste: %v", err)
	}

	ensureSchema(t, db)
	resetTables(t, db)
	t.Cleanup(func() { resetTables(t, db) })

	return db
}

// ensureSchema aplica a migration inicial só na primeira vez (o schema
// persiste no volume do container entre execuções): as migrations não são
// idempotentes — reaplicar um CREATE TABLE quebraria a segunda execução.
func ensureSchema(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	if err := db.QueryRow(ctx, `SELECT to_regclass('public.products') IS NOT NULL`).Scan(&exists); err != nil {
		t.Fatalf("verificar schema de teste: %v", err)
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

// resetTables esvazia todas as tabelas de dados do database exclusivo de
// teste, sempre nesta única instrução (o Postgres exige que tabelas ligadas
// por FK sejam truncadas juntas). O schema é preservado.
func resetTables(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const truncate = `TRUNCATE TABLE order_items, orders, user_password_credentials, users, products RESTART IDENTITY CASCADE`
	if _, err := db.Exec(ctx, truncate); err != nil {
		t.Fatalf("limpar dados de teste: %v", err)
	}
}

// requireTestDatabaseURL impede que setup ou TRUNCATE sejam executados por
// engano no database da aplicação. Além do container separado, o nome
// terminado em _test funciona como uma segunda barreira de segurança.
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
	// internal/repository/testhelpers_test.go -> internal -> backend
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	return filepath.Join(append([]string{projectRoot}, parts...)...)
}
