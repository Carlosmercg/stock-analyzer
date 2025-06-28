package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

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
		query := db.NewSelect().Model(&[]models.StockItem{})

		// Filtros opcionales
		if ticker := c.Query("ticker"); ticker != "" {
			query = query.Where("ticker = ?", strings.ToUpper(ticker))
		}
		if brokerage := c.Query("brokerage"); brokerage != "" {
			query = query.Where("brokerage = ?", brokerage)
		}
		if ratingTo := c.Query("rating_to"); ratingTo != "" {
			query = query.Where("rating_to = ?", ratingTo)
		}
		if action := c.Query("action"); action != "" {
			query = query.Where("action = ?", action)
		}

		// Rango de fechas
		if from := c.Query("from"); from != "" {
			if t, err := time.Parse("2006-01-02", from); err == nil {
				query = query.Where("time >= ?", t)
			}
		}
		if to := c.Query("to"); to != "" {
			if t, err := time.Parse("2006-01-02", to); err == nil {
				query = query.Where("time <= ?", t)
			}
		}

		// Precio objetivo (target_to) como float (quita el $)
		if min := c.Query("target_min"); min != "" {
			query = query.Where("CAST(REPLACE(target_to, '$', '') AS FLOAT) >= ?", min)
		}
		if max := c.Query("target_max"); max != "" {
			query = query.Where("CAST(REPLACE(target_to, '$', '') AS FLOAT) <= ?", max)
		}

		if company := c.Query("company"); company != "" {
			query = query.Where("company ILIKE ?", company+"%")
		}

		// Ejecutar query
		var results []models.StockItem
		if err := query.Order("time DESC").Scan(c, &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error consultando datos"})
			return
		}

		c.JSON(http.StatusOK, results)
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


func parseDollar(s string) (float64, error) {
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}