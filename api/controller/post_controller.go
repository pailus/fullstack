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

func (server *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	post := models.Post{}

	err = json.Unmarshal(body, &post)

	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	post.Prepare()
	err = post.Validate()
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		response.ERROR(w, http.StatusUnauthorized, errors.New("Unathorized"))
		return
	}
	if uid != post.AuthorID {
		response.ERROR(w, http.StatusUnauthorized, errors.New("Unathorized"))
		return
	}

	postCreated, err := post.SavePost(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		response.ERROR(w, http.StatusInternalServerError, formattedError)
	}

	w.Header().Set("Lacation", fmt.Sprintf("%s%s%s", r.Host, r.URL.Path, postCreated.ID))
	response.JSON(w, http.StatusCreated, postCreated)

}

func (server *Server) GetPosts(w http.ResponseWriter, r *http.Request) {

	post := models.Post{}

	posts, err := post.FindAllPosts(server.DB)
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	response.JSON(w, http.StatusOK, posts)
}

func (server *Server) GetPost(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	post := models.Post{}

	postReceived, err := post.FindPostByID(server.DB, pid)
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	response.JSON(w, http.StatusOK, postReceived)
}

func (server *Server) UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}

	uid, err := auth.ExtractTokenID(r)

	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
	}

	post := models.Post{}

	err = server.DB.Debug().Model(models.Post{}).Where("id=?", pid).Take(&post).Error

	if err != nil {
		response.ERROR(w, http.StatusNotFound, errors.New("Post Not Found"))
		return
	}

	if uid != post.AuthorID {
		response.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	postUpdate := models.Post{}

	err = json.Unmarshal(body, &postUpdate)
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if uid != postUpdate.AuthorID {
		response.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	postUpdate.Prepare()
	err = postUpdate.Validate()

	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	postUpdate.ID = post.ID

	postUpdated, err := postUpdate.UpdateAPost(server.DB)

	if err != nil {
		formaterror := formaterror.FormatError(err.Error())
		response.ERROR(w, http.StatusInternalServerError, formaterror)
		return
	}
	response.JSON(w, http.StatusOK, postUpdated)

}

func (server *Server) DeletePost(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// Is a valid post id given to us?
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Is this user authenticated?
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		response.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the post exist
	post := models.Post{}
	err = server.DB.Debug().Model(models.Post{}).Where("id = ?", pid).Take(&post).Error
	if err != nil {
		response.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}

	// Is the authenticated user, the owner of this post?
	if uid != post.AuthorID {
		response.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	_, err = post.DeleteAPost(server.DB, pid, uid)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", pid))
	response.JSON(w, http.StatusNoContent, "")
}
