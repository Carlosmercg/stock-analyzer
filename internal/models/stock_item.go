package models

import (
	"time"

	"github.com/uptrace/bun"
)

type StockItem struct {
	bun.BaseModel `bun:"table:stock_items"`

	ID         int64     `bun:",pk,autoincrement"` // Primary key
	Ticker     string    `bun:"ticker,notnull"`
	TargetFrom string    `bun:"target_from,notnull"`
	TargetTo   string    `bun:"target_to,notnull"`
	Company    string    `bun:"company,notnull"`
	Action     string    `bun:"action,notnull"`
	Brokerage  string    `bun:"brokerage,notnull"`
	RatingFrom string    `bun:"rating_from,notnull"`
	RatingTo   string    `bun:"rating_to,notnull"`
	Time       time.Time `bun:"time,notnull,type:timestamptz"` // Usa el tipo adecuado de CockroachDB
}
