package api

import (
	"net/http"

	"github.com/smallnest/gen/example/model"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/dbmeta"
)

func configSalariesRouter(router *httprouter.Router) {
	router.GET("/salaries", GetAllSalaries)
	router.POST("/salaries", AddSalary)
	router.GET("/salaries/:id", GetSalary)
	router.PUT("/salaries/:id", UpdateSalary)
	router.DELETE("/salaries/:id", DeleteSalary)
}

func configGinSalariesRouter(router gin.IRoutes) {
	router.GET("/salaries", ConverHttprouterToGin(GetAllSalaries))
	router.POST("/salaries", ConverHttprouterToGin(AddSalary))
	router.GET("/salaries/:id", ConverHttprouterToGin(GetSalary))
	router.PUT("/salaries/:id", ConverHttprouterToGin(UpdateSalary))
	router.DELETE("/salaries/:id", ConverHttprouterToGin(DeleteSalary))
}

func GetAllSalaries(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, err := readInt(r, "page", 0)
	if err != nil || page < 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := r.FormValue("order")

	salaries := []*model.Salary{}

	salaries_orm := DB.Model(&model.Salary{})

	if page > 0 {
		pagesize, err := readInt(r, "pagesize", 20)
		if err != nil || pagesize <= 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offset := (page - 1) * pagesize

		salaries_orm = salaries_orm.Offset(offset).Limit(pagesize)
	}

	if order != "" {
		salaries_orm = salaries_orm.Order(order)
	}

	if err = salaries_orm.Find(&salaries).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, &salaries)
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
