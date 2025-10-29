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
	tmpl := template.Must(template.ParseFiles("templates/species/form.html"))
	tmpl.Execute(w, &db.Species{})
}

func CreateSpecies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/species", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	// Parse stratification steps
	stepTypes := r.Form["step_type[]"]
	stepMoists := r.Form["step_moist[]"]
	stepDays := r.Form["step_days[]"]

	var steps []db.StratificationStep
	for i := range stepTypes {
		if stepTypes[i] == "" {
			continue // Skip empty steps
		}

		days, _ := strconv.Atoi(stepDays[i])
		moist := false
		// Check if this step index is in the moist array
		for _, moistVal := range stepMoists {
			if moistVal == strconv.Itoa(i) {
				moist = true
				break
			}
		}

		steps = append(steps, db.StratificationStep{
			Type:  stepTypes[i],
			Moist: moist,
			Days:  days,
		})
	}

	species := &db.Species{
		Name:                r.FormValue("name"),
		StratificationSteps: steps,
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

	tmpl := template.Must(template.ParseFiles("templates/species/form.html"))
	tmpl.Execute(w, species)
}

func UpdateSpecies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/species", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	id, _ := strconv.Atoi(r.FormValue("id"))

	// Parse stratification steps
	stepTypes := r.Form["step_type[]"]
	stepMoists := r.Form["step_moist[]"]
	stepDays := r.Form["step_days[]"]

	var steps []db.StratificationStep
	for i := range stepTypes {
		if stepTypes[i] == "" {
			continue // Skip empty steps
		}

		days, _ := strconv.Atoi(stepDays[i])
		moist := false
		// Check if this step index is in the moist array
		for _, moistVal := range stepMoists {
			if moistVal == strconv.Itoa(i) {
				moist = true
				break
			}
		}

		steps = append(steps, db.StratificationStep{
			Type:  stepTypes[i],
			Moist: moist,
			Days:  days,
		})
	}

	species := &db.Species{
		ID:                  id,
		Name:                r.FormValue("name"),
		StratificationSteps: steps,
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
