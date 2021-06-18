package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pailus/fullstack/api/auth"
	"github.com/pailus/fullstack/api/auth/formaterror"
	"github.com/pailus/fullstack/api/models"
	"github.com/pailus/fullstack/api/response"
)

func (server *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	user.Prepare()
	err = user.Validate("")
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	userCreated, err := user.SaveUser(server.DB)

	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		response.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s%s%d", r.Host, r.RequestURI, userCreated.ID))
	response.JSON(w, http.StatusCreated, userCreated)

}

func (server *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	users, err := user.FindAllUser(server.DB)
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	response.JSON(w, http.StatusOK, users)
}

func (server *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 30)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user := models.User{}
	userGotten, err := user.FindUserByID(server.DB, uint32(uid))

	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	response.JSON(w, http.StatusOK, userGotten)

}

func (server *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 30)
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	tokenID, err := auth.ExtractTokenID(r)

	if err != nil {
		response.ERROR(w, http.StatusUnauthorized, errors.New("Unathorized"))
		return
	}
	if tokenID != uint32(uid) {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	user.Prepare()
	err = user.Validate("update")
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	updatedUser, err := user.UpdateAUser(server.DB, uint32(uid))

	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		response.ERROR(w, http.StatusInternalServerError, formattedError)
	}
	response.JSON(w, http.StatusOK, updatedUser)

}

func (server *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := models.User{}
	uid, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		response.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if tokenID != 0 && tokenID != uint32(uid) {
		response.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	_, err = user.DeleteAuser(server.DB, uint32(uid))
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", uid))
	response.JSON(w, http.StatusNoContent, "")
}
