package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/example/model"
)

func configTitlesRouter(router *httprouter.Router) {
	router.GET("/titles", GetAllTitles)
	router.POST("/titles", PostTitle)
	router.GET("/titles/:id", GetTitle)
	router.PUT("/titles/:id", PutTitle)
	router.DELETE("/titles/:id", DeleteTitle)
}

func GetAllTitles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	titles := []model.Title{}
	DB.Find(&titles)
	writeJSON(w, &titles)
}

func GetTitle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	title := &model.Title{}
	if DB.First(title, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, title)
}

func PostTitle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	title := &model.Title{}

	if err := readJSON(r, title); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save(title).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, title)
}

func PutTitle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	title := &model.Title{}
	if DB.First(title, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.Title{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: copy necessary fields from updated to title

	if err := DB.Save(title).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, title)
}

func DeleteTitle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	title := &model.Title{}

	if DB.First(title, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete(title).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
