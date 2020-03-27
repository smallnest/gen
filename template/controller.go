package template

var ControllerTmpl = `package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
	"{{.PackageName}}"
)

func config{{pluralize .StructName}}Router(router *httprouter.Router) {
	router.GET("/{{pluralize .StructName | toLower}}", GetAll{{pluralize .StructName}})
	router.POST("/{{pluralize .StructName | toLower}}", Add{{.StructName}})
	router.GET("/{{pluralize .StructName | toLower}}/:id", Get{{.StructName}})
	router.PUT("/{{pluralize .StructName | toLower}}/:id", Update{{.StructName}})
	router.DELETE("/{{pluralize .StructName | toLower}}/:id", Delete{{.StructName}})
}

func GetAll{{pluralize .StructName}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	
	order := r.FormValue("order")

	{{pluralize .StructName | toLower}} := []*model.{{.StructName}}{}

	{{pluralize .StructName | toLower}}_orm := DB.Model(&model.{{.StructName}}{})

	if page > 0 {
		pagesize, err := readInt(r, "pagesize", 20)
		if err != nil || pagesize <= 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		offset := (page - 1) * pagesize

		{{pluralize .StructName | toLower}}_orm = {{pluralize .StructName | toLower}}_orm.Offset(offset).Limit(pagesize)
	}
	
	if order != "" {
		{{pluralize .StructName | toLower}}_orm = {{pluralize .StructName | toLower}}_orm.Order(order)
	}

	if err = {{pluralize .StructName | toLower}}_orm.Find(&{{pluralize .StructName | toLower}}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

func Add{{.StructName}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func Update{{.StructName}}(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	if err := dbmeta.Copy({{.StructName | toLower}}, updated); err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
