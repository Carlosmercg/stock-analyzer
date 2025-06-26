package main

import (
	"log"
	"os"

	"github.com/Carlosmercg/stock-analyzer/internal/database"
	"github.com/Carlosmercg/stock-analyzer/internal/router"
	"github.com/Carlosmercg/stock-analyzer/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Conectar a Cockroach y guardar instancia
	db := database.InitCockroach()

	// 2. Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Fatalf("❌ Error cargando .env: %v", err)
	}

	// 3. Crear tabla y cargar datos si no existe
	if !database.TableExists(db, "stock_items") {
		log.Println("🆕 Tabla no existe, creando tabla y cargando datos iniciales...")
		database.Migrate(db)

		if err := service.FetchAndStoreStocks(db); err != nil {
			log.Fatalf("❌ Error descargando y guardando datos: %v", err)
		}
	} else {
		log.Println("ℹ️  Tabla stock_items ya existe, no se realiza migración ni carga.")
	}

	// 4. Configurar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := router.SetupRouter(db)

	log.Printf("🚀 Servidor escuchando en http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}
