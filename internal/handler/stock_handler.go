package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/Carlosmercg/stock-analyzer/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func GetAllStocks(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stocks []models.StockItem

		err := db.NewSelect().
			Model(&stocks).
			Order("time DESC").
			Scan(c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron obtener los datos"})
			return
		}

		c.JSON(http.StatusOK, stocks)
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
		query = query.Where("company = ?", company)
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
