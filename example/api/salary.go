package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/smallnest/gen/example/model"
)

func configSalariesRouter(router *httprouter.Router) {
	router.GET("/salaries", GetAllSalaries)
	router.POST("/salaries", PostSalary)
	router.GET("/salaries/:id", GetSalary)
	router.PUT("/salaries/:id", PutSalary)
	router.DELETE("/salaries/:id", DeleteSalary)
}

func GetAllSalaries(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	salaries := []model.Salary{}
	DB.Find(&salaries)
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

func PostSalary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func PutSalary(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// TODO: copy necessary fields from updated to salary

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
