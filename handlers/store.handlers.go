package handlers

import (
	"net/http"

	"github.com/labstack/gommon/log"
)

func HandleGetWebSocket(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	if _, err := w.Write([]byte("Hello")); err != nil {
		log.Error(err)
	}

}
