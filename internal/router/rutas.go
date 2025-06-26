package router

import (
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func SetupRouter(db *bun.DB) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	RegisterStockRoutes(api, db)

	return router
}
