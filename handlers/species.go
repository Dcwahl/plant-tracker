package handlers

import (
	"html/template"
	"net/http"
	"plant-tracker/db"
	"strconv"
)

func ListSpecies(w http.ResponseWriter, r *http.Request) {
	species, err := db.GetAllSpecies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/species/list.html"))
	tmpl.Execute(w, species)
}

func NewSpecies(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/species/new.html"))
	tmpl.Execute(w, nil)
}

func CreateSpecies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/species", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	coldStratified := r.FormValue("cold_stratified") == "on"

	var stratificationDays *int
	if sd := r.FormValue("stratification_days"); sd != "" {
		val, _ := strconv.Atoi(sd)
		stratificationDays = &val
	}

	species := &db.Species{
		Name:               r.FormValue("name"),
		ColdStratified:     coldStratified,
		StratificationDays: stratificationDays,
	}

	if err := db.CreateSpecies(species); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/species", http.StatusSeeOther)
}

func EditSpecies(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	species, err := db.GetSpeciesByID(id)
	if err != nil {
		http.Error(w, "Species not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/species/edit.html"))
	tmpl.Execute(w, species)
}

func UpdateSpecies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/species", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	id, _ := strconv.Atoi(r.FormValue("id"))
	coldStratified := r.FormValue("cold_stratified") == "on"

	var stratificationDays *int
	if sd := r.FormValue("stratification_days"); sd != "" {
		val, _ := strconv.Atoi(sd)
		stratificationDays = &val
	}

	species := &db.Species{
		ID:                 id,
		Name:               r.FormValue("name"),
		ColdStratified:     coldStratified,
		StratificationDays: stratificationDays,
	}

	if err := db.UpdateSpecies(species); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/species", http.StatusSeeOther)
}

func DeleteSpecies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/species", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))

	if err := db.DeleteSpecies(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/species", http.StatusSeeOther)
}
