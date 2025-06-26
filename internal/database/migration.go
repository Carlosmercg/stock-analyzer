package database

import (
	"context"
	"log"

	"github.com/Carlosmercg/stock-analyzer/internal/models"
	"github.com/uptrace/bun"
)

func Migrate(db *bun.DB) {
	ctx := context.Background()

	_, err := db.NewCreateTable().
		Model((*models.StockItem)(nil)).
		IfNotExists().
		Exec(ctx)

	if err != nil {
		log.Fatalf("❌ Error creando tabla stock_items: %v", err)
	} else {
		log.Println("✅ Tabla stock_items creada o ya existía.")
	}
}

func TableExists(db *bun.DB, tableName string) bool {
	ctx := context.Background()

	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = ?
		)
	`

	err := db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	if err != nil {
		log.Printf("⚠️  Error verificando existencia de tabla %s: %v", tableName, err)
		return false
	}

	return exists
}
