package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
	"github.com/smallnest/gen/example/model"
)

func configSalariesRouter(router *httprouter.Router) {
	router.GET("/salaries", GetAllSalaries)
	router.POST("/salaries", AddSalary)
	router.GET("/salaries/:id", GetSalary)
	router.PUT("/salaries/:id", UpdateSalary)
	router.DELETE("/salaries/:id", DeleteSalary)
}

func GetAllSalaries(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	salaries := []*model.Salary{}

	if order != "" {
		err = DB.Model(&model.Salary{}).Order(order).Offset(offset).Limit(pagesize).Find(&salaries).Error
	} else {
		err = DB.Model(&model.Salary{}).Offset(offset).Limit(pagesize).Find(&salaries).Error
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetSalary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	salary := &model.Salary{}
	if DB.First(salary, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, salary)
}

func AddSalary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	salary := &model.Salary{}

	if err := readJSON(r, salary); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := DB.Save(salary).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, salary)
}

func UpdateSalary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	salary := &model.Salary{}
	if DB.First(salary, id).Error != nil {
		http.NotFound(w, r)
		return
	}

	updated := &model.Salary{}
	if err := readJSON(r, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := dbmeta.Copy(salary, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := DB.Save(salary).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, salary)
}

func DeleteSalary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	salary := &model.Salary{}

	if DB.First(salary, id).Error != nil {
		http.NotFound(w, r)
		return
	}
	if err := DB.Delete(salary).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
