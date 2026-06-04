package handler


type SeatInfo struct {
	Grade string `json:"seat_grade"`
	Quantity int `json:"quantity"`
}
type ReservationRequest struct {
	GameID string `json:"game_id"`
	Seats []SeatInfo `json:"seats"`
}

func (h *Handler) HandleCreateReservation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var reqBody ReservationRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		h.respondError(w, err, http.StatusBadRequest)
		return
	}

	

}