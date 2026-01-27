package models

type DateTZ struct {
	Date         string `json:"date" db:"date"`
	TimezoneType int64  `json:"timezone_type" db:"timezone_type"`
	Timezone     string `json:"timezone" db:"timezone"`
}

type Rec struct {
	Date string `json:"date" db:"date"`
	Type string `json:"type" db:"type"`
}
