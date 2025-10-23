package handlers

import (
	"html/template"
	"net/http"
	"plant-tracker/db"
	"strconv"
	"time"
)

func ListBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := db.GetAllBatches()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/batches/list.html"))
	tmpl.Execute(w, batches)
}

func NewBatch(w http.ResponseWriter, r *http.Request) {
	species, err := db.GetAllSpecies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Species   []db.Species
		TodayDate string
	}{
		Species:   species,
		TodayDate: time.Now().Format("2006-01-02"),
	}

	tmpl := template.Must(template.ParseFiles("templates/batches/new.html"))
	tmpl.Execute(w, data)
}

func CreateBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	speciesID, _ := strconv.Atoi(r.FormValue("species_id"))
	numCells, _ := strconv.Atoi(r.FormValue("num_cells"))
	seedsPerCell, _ := strconv.Atoi(r.FormValue("seeds_per_cell"))
	dateCreated := r.FormValue("date_created")

	var totalSeeds *int
	if ts := r.FormValue("total_seeds"); ts != "" {
		val, _ := strconv.Atoi(ts)
		totalSeeds = &val
	}

	batch := &db.Batch{
		SpeciesID:    speciesID,
		NumCells:     numCells,
		SeedsPerCell: seedsPerCell,
		TotalSeeds:   totalSeeds,
		DateCreated:  dateCreated,
	}

	if err := db.CreateBatch(batch); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func EditBatch(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	batch, err := db.GetBatchByID(id)
	if err != nil {
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}

	species, err := db.GetAllSpecies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Batch   *db.Batch
		Species []db.Species
	}{
		Batch:   batch,
		Species: species,
	}

	tmpl := template.Must(template.ParseFiles("templates/batches/edit.html"))
	tmpl.Execute(w, data)
}

func UpdateBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	id, _ := strconv.Atoi(r.FormValue("id"))
	speciesID, _ := strconv.Atoi(r.FormValue("species_id"))
	numCells, _ := strconv.Atoi(r.FormValue("num_cells"))
	seedsPerCell, _ := strconv.Atoi(r.FormValue("seeds_per_cell"))
	dateCreated := r.FormValue("date_created")

	var totalSeeds *int
	if ts := r.FormValue("total_seeds"); ts != "" {
		val, _ := strconv.Atoi(ts)
		totalSeeds = &val
	}

	batch := &db.Batch{
		ID:           id,
		SpeciesID:    speciesID,
		NumCells:     numCells,
		SeedsPerCell: seedsPerCell,
		TotalSeeds:   totalSeeds,
		DateCreated:  dateCreated,
	}

	if err := db.UpdateBatch(batch); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func DeleteBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))

	if err := db.DeleteBatch(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
