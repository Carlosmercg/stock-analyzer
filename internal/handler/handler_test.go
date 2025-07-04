package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Carlosmercg/stock-analyzer/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

func setupTestDB(t *testing.T) *bun.DB {
	sqliteDB, err := sql.Open(sqliteshim.ShimName, ":memory:")
	assert.NoError(t, err)

	db := bun.NewDB(sqliteDB, sqlitedialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// Crear tabla
	err = db.ResetModel(contextBackground(), (*models.StockItem)(nil))
	assert.NoError(t, err)

	// Insertar datos de prueba
	_, err = db.NewInsert().Model(&[]models.StockItem{
		{
			Ticker:     "AAPL",
			Brokerage:  "Goldman",
			RatingTo:   "Buy",
			TargetFrom: "$150",
			TargetTo:   "$180",
			Action:     "Raised target",
			Company:    "Apple Inc.",
		},
		{
			Ticker:     "GOOG",
			Brokerage:  "Morgan",
			RatingTo:   "Neutral",
			TargetFrom: "$100",
			TargetTo:   "$120",
			Action:     "Initiated",
			Company:    "Alphabet Inc.",
		},
	}).
		Exec(contextBackground())
	assert.NoError(t, err)

	return db
}

func TestGetAllStocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	req, _ := http.NewRequest(http.MethodGet, "/stocks?page=1&limit=1", nil)
	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = req

	handler := GetAllStocks(db)
	handler(c)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Apple")
}

func contextBackground() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

func performRequest(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestGetFilteredStocks(t *testing.T) {
	db := setupTestDB(t)
	router := gin.Default()
	router.GET("/filtered", GetFilteredStocks(db))

	resp := performRequest(router, "GET", "/filtered?ticker=AAPL")
	assert.Equal(t, 200, resp.Code)

	var body map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &body)
	assert.NoError(t, err)

	assert.Equal(t, float64(1), body["total"])
}

func TestGetTopInvestmentStocks(t *testing.T) {
	db := setupTestDB(t)
	router := gin.Default()
	router.GET("/top", GetTopInvestmentStocks(db))

	resp := performRequest(router, "GET", "/top")
	assert.Equal(t, 200, resp.Code)

	var body []map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.True(t, len(body) > 0)
}

func TestGetTopStocksByBrokerage(t *testing.T) {
	db := setupTestDB(t)
	router := gin.Default()
	router.GET("/brokerage", GetTopStocksByBrokerage(db))

	resp := performRequest(router, "GET", "/brokerage?brokerage=goldman")
	assert.Equal(t, 200, resp.Code)

	var body []map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.True(t, len(body) > 0, "Debería haber al menos un resultado")
	assert.Equal(t, "AAPL", body[0]["Ticker"])
	assert.Contains(t, body[0], "score", "Debería incluir el campo score")
}

func TestGetDistinctBrokerages(t *testing.T) {
	db := setupTestDB(t)
	router := gin.Default()
	router.GET("/brokerages", GetDistinctBrokerages(db))

	resp := performRequest(router, "GET", "/brokerages")
	assert.Equal(t, 200, resp.Code)

	var body []string
	err := json.Unmarshal(resp.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Contains(t, body, "Goldman")
	assert.Contains(t, body, "Morgan")
}

func TestGetDistinctRatings(t *testing.T) {
	db := setupTestDB(t)
	router := gin.Default()
	router.GET("/ratings", GetDistinctRatings(db))

	resp := performRequest(router, "GET", "/ratings")
	assert.Equal(t, 200, resp.Code)

	var body []string
	err := json.Unmarshal(resp.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Contains(t, body, "Buy")
	assert.Contains(t, body, "Neutral")
}

func TestGetCompanyInfoFromFinnhub(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Simular variables de entorno
	os.Setenv("FINNHUB_APIKEY", "dummykey")
	os.Setenv("FINNHUB_URL", "https://finnhub.io/api/v1/stock/profile2?symbol=%s&token=%s")

	req, _ := http.NewRequest(http.MethodGet, "/?ticker=AAPL", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetCompanyInfoFromFinnhub()(c)

	// Como estamos usando una API key dummy, esperamos un error 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var res map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Contains(t, res, "error")
	assert.Contains(t, res["error"], "Error desde Finnhub")
}

func TestGetCompanyInfoFromFinnhub_MissingTicker(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	c.Request = req

	GetCompanyInfoFromFinnhub()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParseDollar(t *testing.T) {
	val, err := parseDollar("$1,234.56")
	assert.NoError(t, err)
	assert.Equal(t, 1234.56, val)
}
