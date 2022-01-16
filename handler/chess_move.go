package handler

import (
	"encoding/json"
	"net/http"

	"github.com/eco_codes/scraper"
	"github.com/gorilla/mux"
)

type ChessMoveHandler struct {
	scraper scraper.Scraper
}

func NewChessMoveHandler(scraper scraper.Scraper) *ChessMoveHandler {
	return &ChessMoveHandler{scraper: scraper}
}

func (h *ChessMoveHandler) Get(w http.ResponseWriter, r *http.Request) {
	res, err := h.scraper.GetAll()
	if err != nil {
		http.Error(w, "error scraping site for data", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func (h *ChessMoveHandler) GetMoves(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code, ok := vars["CODE"]
	if !ok {
		http.Error(w, "CODE is missing in path", http.StatusBadRequest)
		return
	}

	res, err := h.scraper.GetByCode(code)
	if err != nil {
		http.Error(w, "error scraping site for data", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}
