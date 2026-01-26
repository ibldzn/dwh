package models

type DateTZ struct {
	Date         string `json:"date"`
	TimezoneType int64  `json:"timezone_type"`
	Timezone     string `json:"timezone"`
}

type Rec struct {
	Date string `json:"date"`
	Type string `json:"type"`
}
