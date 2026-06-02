package domain

type Seat struct {
	Grade string `json:"seat_grade"`
	Price int `json:"price"`
	Total int `json:"total_seats"`
	Available int `json:"available_seats"`
	Reserved int `json:"reserved_seats"`
	Sold int `json:"sold_seats"`
}