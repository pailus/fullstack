package controller

import (
	"net/http"

	"github.com/pailus/fullstack/api/response"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, "Wellcome This Api")
}
