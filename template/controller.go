package template

var ControllerTmpl = `package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"{{.PackageName}}"
)

func config{{pluralize .StructName}}Router(router *httprouter.Router) {
	router.GET("/{{pluralize .StructName | toLower}}", GetAll{{pluralize .StructName}})
	router.POST("/{{pluralize .StructName | toLower}}", Post{{.StructName}})
	router.GET("/{{pluralize .StructName | toLower}}/:id", Get{{.StructName}})
	router.PUT("/{{pluralize .StructName | toLower}}/:id", Put{{.StructName}})
	router.DELETE("/{{pluralize .StructName | toLower}}/:id", Delete{{.StructName}})
}

func GetAll{{pluralize .StructName}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	{{pluralize .StructName | toLower}} := []model.{{.StructName}}{}
	DB.Find(&{{pluralize .StructName | toLower}})
	writeJSON(w, &{{pluralize .StructName | toLower}})
}

func Get{{.StructName}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	{{.StructName | toLower}} := &model.{{.StructName}}{}
	if DB.First({{.StructName | toLower}}, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, {{.StructName | toLower}})
}

func Post{{.StructName}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	{{.StructName | toLower}} := &model.{{.StructName}}{}

	if err := readJSON(r, {{.StructName | toLower}}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save({{.StructName | toLower}}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, {{.StructName | toLower}})
}

func Put{{.StructName}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	{{.StructName | toLower}} := &model.{{.StructName}}{}
	if DB.First({{.StructName | toLower}}, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.{{.StructName}}{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: copy necessary fields from updated to {{.StructName | toLower}}

	if err := DB.Save({{.StructName | toLower}}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, {{.StructName | toLower}})
}

func Delete{{.StructName}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	{{.StructName | toLower}} := &model.{{.StructName}}{}

	if DB.First({{.StructName | toLower}}, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete({{.StructName | toLower}}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
`
