package router

import (
	"github.com/Carlosmercg/stock-analyzer/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func RegisterStockRoutes(r *gin.RouterGroup, db *bun.DB) {
	stock := r.Group("/stocks")
	{
		stock.GET("/", handler.GetAllStocks(db))
		stock.GET("/filter", handler.GetFilteredStocks(db))
		stock.GET("/top", handler.GetTopInvestmentStocks(db))
		stock.GET("/top-by-brokerage", handler.GetTopStocksByBrokerage(db))
		stock.GET("/brokerages", handler.GetDistinctBrokerages(db))
		stock.GET("/ratings", handler.GetDistinctRatings(db))



	}
}
