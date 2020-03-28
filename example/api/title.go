package api

import (
	"net/http"

	"github.com/smallnest/gen/example/model"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
)

func configTitlesRouter(router *httprouter.Router) {
	router.GET("/titles", GetAllTitles)
	router.POST("/titles", AddTitle)
	router.GET("/titles/:id", GetTitle)
	router.PUT("/titles/:id", UpdateTitle)
	router.DELETE("/titles/:id", DeleteTitle)
}

func configGinTitlesRouter(router gin.IRoutes) {
	router.GET("/titles", ConverHttprouterToGin(GetAllTitles))
	router.POST("/titles", ConverHttprouterToGin(AddTitle))
	router.GET("/titles/:id", ConverHttprouterToGin(GetTitle))
	router.PUT("/titles/:id", ConverHttprouterToGin(UpdateTitle))
	router.DELETE("/titles/:id", ConverHttprouterToGin(DeleteTitle))
}

func GetAllTitles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := r.FormValue("order")

	titles := []*model.Title{}

	titles_orm := DB.Model(&model.Title{})

	if page > 0 {
		pagesize, err := readInt(r, "pagesize", 20)
		if err != nil || pagesize <= 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset := (page - 1) * pagesize

		titles_orm = titles_orm.Offset(offset).Limit(pagesize)
	}

	if order != "" {
		titles_orm = titles_orm.Order(order)
	}

	if err = titles_orm.Find(&titles).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
