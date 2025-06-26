package dto

import "time"

type StockItem struct {
	Ticker     string `json:"ticker"`
	TargetFrom string `json:"target_from"`
	TargetTo   string `json:"target_to"`
	Company    string `json:"company"`
	Action     string `json:"action"`
	Brokerage  string `json:"brokerage"`
	RatingFrom string `json:"rating_from"`
	RatingTo   string `json:"rating_to"`
	Time       string `json:"time"`
}

func (s StockItem) ParseTime() time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s.Time)
	return t
}
