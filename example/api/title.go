package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
	"github.com/smallnest/gen/example/model"
)

func configTitlesRouter(router *httprouter.Router) {
	router.GET("/titles", GetAllTitles)
	router.POST("/titles", AddTitle)
	router.GET("/titles/:id", GetTitle)
	router.PUT("/titles/:id", UpdateTitle)
	router.DELETE("/titles/:id", DeleteTitle)
}

func GetAllTitles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 1)
	if err != nil || page < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	pagesize, err := readInt(r, "pagesize", 20)
	if err != nil || pagesize <= 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	offset := (page - 1) * pagesize

	order := r.FormValue("order")

	titles := []*model.Title{}

	if order != "" {
		err = DB.Model(&model.Title{}).Order(order).Offset(offset).Limit(pagesize).Find(&titles).Error
	} else {
		err = DB.Model(&model.Title{}).Offset(offset).Limit(pagesize).Find(&titles).Error
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func AddTitle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func UpdateTitle(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	if err := dbmeta.Copy(title, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
