// Command seed executa database/seeds/seed.sql contra o banco configurado
// em .env — os mesmos dados de desenvolvimento (usuários, produtos e
// pedidos) usados para testar a API manualmente.
//
// Uso: go run ./cmd/seed
package main

import (
	"context"
	"log"

	"ecommerce/database/seeds"
	"ecommerce/internal/config"
	"ecommerce/internal/database"

	"github.com/jackc/pgx/v5"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Executando seed de desenvolvimento — não use em produção.")

	// seed.sql contém vários comandos separados por ";" (BEGIN/COMMIT,
	// INSERTs, um bloco DO $$ ... $$ e SELECTs de conferência). O protocolo
	// simples do Postgres é o único modo do pgx que executa uma string com
	// múltiplos comandos numa única chamada.
	if _, err := db.Exec(context.Background(), seeds.SeedSQL, pgx.QueryExecModeSimpleProtocol); err != nil {
		log.Fatalf("Falha ao executar o seed: %v", err)
	}

	log.Println("Seed executado com sucesso.")
}
