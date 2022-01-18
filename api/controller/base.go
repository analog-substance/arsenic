package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type baseController struct {
	router *mux.Router
}

func (controller baseController) useCorsMiddleware() {
	controller.router.Use(mux.CORSMethodMiddleware(controller.router))
}

func (controller baseController) setCorsHeaders(rw http.ResponseWriter) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "*")
}

func (controller baseController) genericError(rw http.ResponseWriter) {
	json.NewEncoder(rw).Encode(struct {
		Msg string
	}{
		Msg: "An error occurred",
	})
}

func (controller baseController) genericSuccess(rw http.ResponseWriter) {
	json.NewEncoder(rw).Encode(struct {
		Status string
	}{
		Status: "OK",
	})
}
