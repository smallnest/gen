package api

import (
	"net/http"

	"generated/model"
	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
)

func configExamplesRouter(router *httprouter.Router) {
	router.GET("/examples", GetAllExamples)
	router.POST("/examples", AddExample)
	router.GET("/examples/:id", GetExample)
	router.PUT("/examples/:id", UpdateExample)
	router.DELETE("/examples/:id", DeleteExample)
}

func GetAllExamples(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	examples := []*model.Example{}

	if order != "" {
		err = DB.Model(&model.Example{}).Order(order).Offset(offset).Limit(pagesize).Find(&examples).Error
	} else {
		err = DB.Model(&model.Example{}).Offset(offset).Limit(pagesize).Find(&examples).Error
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, &examples)
}

func GetExample(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	example := &model.Example{}
	if DB.First(example, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, example)
}

func AddExample(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	example := &model.Example{}

	if err := readJSON(r, example); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save(example).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, example)
}

func UpdateExample(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	example := &model.Example{}
	if DB.First(example, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.Example{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := dbmeta.Copy(example, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := DB.Save(example).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, example)
}

func DeleteExample(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	example := &model.Example{}

	if DB.First(example, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete(example).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
