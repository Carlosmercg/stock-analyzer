package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Carlosmercg/stock-analyzer/internal/dto"
	"github.com/uptrace/bun"
)


type APIResponse struct {
	Items    []dto.StockItem `json:"items"`
	NextPage string          `json:"next_page"`
}

// FetchAndStoreStocks descarga los datos y los guarda en la base de datos
func FetchAndStoreStocks(db *bun.DB) error {

	apiURL := os.Getenv("API_URL")
	authHeader := os.Getenv("AUTH_HEADER")

	if apiURL == "" || authHeader == "" {
	return fmt.Errorf("las variables de entorno API_URL o AUTH_HEADER no están definidas")
}

	url := apiURL
	page := 1
	ctx := context.Background()

	for {
		fmt.Printf("📦 Descargando página %d...\n", page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("error creando request: %v", err)
		}

		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error haciendo request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("respuesta no exitosa: %d\n%s", resp.StatusCode, string(body))
		}

		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			return fmt.Errorf("error decodificando JSON: %v", err)
		}

		// Insertar todos los elementos de la página de una vez (más eficiente)
		if len(apiResp.Items) > 0 {
			_, err := db.NewInsert().Model(&apiResp.Items).Exec(ctx)
			if err != nil {
				return fmt.Errorf("error insertando en DB: %v", err)
			}
		}

		if apiResp.NextPage == "" {
			break
		}
		url = apiURL + "?next_page=" + apiResp.NextPage
		page++
	}

	fmt.Println("✅ Datos descargados y almacenados con éxito.")
	return nil
}


