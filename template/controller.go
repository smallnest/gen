package template

var ControllerTmpl = `package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"../model"
)

func config{{pluralize .}}Router(router *httprouter.Router) {
	router.GET("/{{pluralize . | toLower}}", GetAll{{pluralize .}})
	router.POST("/{{pluralize . | toLower}}", Post{{.}})
	router.GET("/{{pluralize . | toLower}}/:id", Get{{.}})
	router.PUT("/{{pluralize . | toLower}}/:id", Put{{.}})
	router.DELETE("/{{pluralize . | toLower}}/:id", Delete{{.}})
}

func GetAll{{pluralize .}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	{{pluralize . | toLower}} := []model.{{.}}{}
	DB.Find(&{{pluralize . | toLower}})
	writeJSON(w, &{{pluralize . | toLower}})
}

func Get{{.}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	{{. | toLower}} := &model.{{.}}{}
	if DB.First({{. | toLower}}, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, {{. | toLower}})
}

func Post{{.}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	{{. | toLower}} := &model.{{.}}{}

	if err := readJSON(r, {{. | toLower}}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save({{. | toLower}}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, {{. | toLower}})
}

func Put{{.}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	{{. | toLower}} := &model.{{.}}{}
	if DB.First({{. | toLower}}, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.{{.}}{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: copy necessary fields from updated to {{. | toLower}}

	if err := DB.Save({{. | toLower}}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, {{. | toLower}})
}

func Delete{{.}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	{{. | toLower}} := &model.{{.}}{}

	if DB.First({{. | toLower}}, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete({{. | toLower}}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
`
