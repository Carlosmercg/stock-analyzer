package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Carlosmercg/stock-analyzer/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type StockScore struct {
	models.StockItem
	Score float64 `json:"score"`
}

func GetAllStocks(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 20
		}
		offset := (page - 1) * limit

		// ✅ Corregido: asignar ambos valores de retorno
		total, err := db.NewSelect().Model((*models.StockItem)(nil)).Count(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo contar los registros"})
			return
		}

		var stocks []models.StockItem
		err = db.NewSelect().
			Model(&stocks).
			Order("time DESC").
			Limit(limit).
			Offset(offset).
			Scan(c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron obtener los datos"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  stocks,
			"total": total,
		})
	}
}

func GetFilteredStocks(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		baseQuery := db.NewSelect().Model(&[]models.StockItem{})

		// Filtros opcionales
		if ticker := c.Query("ticker"); ticker != "" {
			baseQuery = baseQuery.Where("ticker = ?", strings.ToUpper(ticker))
		}
		if brokerage := c.Query("brokerage"); brokerage != "" {
			baseQuery = baseQuery.Where("brokerage = ?", brokerage)
		}
		if ratingTo := c.Query("rating_to"); ratingTo != "" {
			baseQuery = baseQuery.Where("rating_to = ?", ratingTo)
		}
		if action := c.Query("action"); action != "" {
			baseQuery = baseQuery.Where("action = ?", action)
		}
		if min := c.Query("target_min"); min != "" {
			baseQuery = baseQuery.Where("CAST(REPLACE(target_to, '$', '') AS FLOAT) >= ?", min)
		}
		if max := c.Query("target_max"); max != "" {
			baseQuery = baseQuery.Where("CAST(REPLACE(target_to, '$', '') AS FLOAT) <= ?", max)
		}
		if company := c.Query("company"); company != "" {
			baseQuery = baseQuery.Where("company ILIKE ?", company+"%")
		}

		// 🔃 Orden dinámico por fecha
		order := strings.ToLower(c.DefaultQuery("order", "desc"))
		if order != "asc" && order != "desc" {
			order = "desc"
		}
		baseQuery = baseQuery.Order("time " + order)

		// Total de resultados antes de paginar
		totalQuery := baseQuery.Clone()
		total, err := totalQuery.Count(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error contando resultados"})
			return
		}

		// Paginación
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "21"))
		if page < 1 {
			page = 1
		}
		if limit <= 0 {
			limit = 21
		}
		offset := (page - 1) * limit

		var results []models.StockItem
		err = baseQuery.Limit(limit).Offset(offset).Scan(c, &results)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error consultando datos"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  results,
			"total": total,
		})
	}
}

func GetTopInvestmentStocks(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stocks []models.StockItem
		err := db.NewSelect().Model(&stocks).Scan(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cargando los datos"})
			return
		}

		// Calcular puntuaciones
		scored := make([]StockScore, 0, len(stocks))
		for _, s := range stocks {
			to, errTo := parseDollar(s.TargetTo)
			from, errFrom := parseDollar(s.TargetFrom)
			if errTo != nil || errFrom != nil || from == 0 {
				continue
			}

			growth := (to - from) / from * 100

			score := growth
			if strings.ToLower(s.RatingTo) == "buy" || strings.ToLower(s.RatingTo) == "outperform" {
				score += 10
			}
			if strings.Contains(strings.ToLower(s.Action), "raised") {
				score += 5
			} else if strings.Contains(strings.ToLower(s.Action), "initiated") {
				score += 2
			} else if strings.Contains(strings.ToLower(s.Action), "downgraded") {
				score -= 5
			}

			scored = append(scored, StockScore{StockItem: s, Score: score})
		}

		// Ordenar descendente por score
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].Score > scored[j].Score
		})

		// Tomar top 20
		limit := 20
		if len(scored) < 20 {
			limit = len(scored)
		}

		c.JSON(http.StatusOK, scored[:limit])
	}
}

func GetTopStocksByBrokerage(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		brokerageParam := c.Query("brokerage")
		if brokerageParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El parámetro 'brokerage' es requerido"})
			return
		}

		var stocks []models.StockItem
		err := db.NewSelect().
			Model(&stocks).
			Where("LOWER(brokerage) = LOWER(?)", brokerageParam).
			Scan(c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cargando los datos"})
			return
		}

		// Calcular puntuaciones
		scored := make([]StockScore, 0, len(stocks))
		for _, s := range stocks {
			to, errTo := parseDollar(s.TargetTo)
			from, errFrom := parseDollar(s.TargetFrom)
			if errTo != nil || errFrom != nil || from == 0 {
				continue
			}

			growth := (to - from) / from * 100

			score := growth
			if strings.ToLower(s.RatingTo) == "buy" || strings.ToLower(s.RatingTo) == "outperform" {
				score += 10
			}
			if strings.Contains(strings.ToLower(s.Action), "raised") {
				score += 5
			} else if strings.Contains(strings.ToLower(s.Action), "initiated") {
				score += 2
			} else if strings.Contains(strings.ToLower(s.Action), "downgraded") {
				score -= 5
			}

			scored = append(scored, StockScore{StockItem: s, Score: score})
		}

		// Ordenar descendente por score
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].Score > scored[j].Score
		})

		// Top 10 o menos si hay pocos
		limit := 10
		if len(scored) < 10 {
			limit = len(scored)
		}

		c.JSON(http.StatusOK, scored[:limit])
	}
}

func GetDistinctBrokerages(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var brokerages []string

		err := db.NewSelect().
			Model((*models.StockItem)(nil)).
			ColumnExpr("DISTINCT brokerage").
			OrderExpr("brokerage ASC").
			Scan(c, &brokerages)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron obtener las corredoras"})
			return
		}

		c.JSON(http.StatusOK, brokerages)
	}
}

func GetDistinctRatings(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ratings []string

		err := db.NewSelect().
			Model((*models.StockItem)(nil)).
			ColumnExpr("DISTINCT rating_to").
			OrderExpr("rating_to ASC").
			Scan(c, &ratings)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron obtener los ratings"})
			return
		}

		c.JSON(http.StatusOK, ratings)
	}
}

func GetCompanyInfoFromFinnhub() gin.HandlerFunc {
	return func(c *gin.Context) {
		ticker := c.Query("ticker")
		if ticker == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ticker requerido"})
			return
		}

		apiKey := os.Getenv("FINNHUB_APIKEY")
		urlTemplate := os.Getenv("FINNHUB_URL")

		if apiKey == "" || urlTemplate == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Falta configuración de Finnhub"})
			return
		}

		url := fmt.Sprintf(urlTemplate, ticker, apiKey)
		resp, err := http.Get(url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error haciendo la solicitud HTTP"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("Error desde Finnhub (código %d)", resp.StatusCode)})
			return
		}

		var profile struct {
			Name     string `json:"name"`
			WebURL   string `json:"weburl"`
			Logo     string `json:"logo"`
			Industry string `json:"finnhubIndustry"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decodificando JSON"})
			return
		}

		result := struct {
			Ticker      string `json:"ticker"`
			Description string `json:"description"`
			Domain      string `json:"domain"`
			LogoURL     string `json:"logo_url"`
		}{
			Ticker:      ticker,
			Description: profile.Industry,
			Domain:      profile.WebURL,
			LogoURL:     profile.Logo,
		}

		c.JSON(http.StatusOK, result)
	}
}

func parseDollar(s string) (float64, error) {
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}
