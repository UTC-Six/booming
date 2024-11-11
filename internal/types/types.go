package types

type Landmine struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius"`                 // 半径，单位米
	Location  string  `json:"location" db:"location"` // 地理空间点，格式为 "POINT(longitude latitude)"
}
