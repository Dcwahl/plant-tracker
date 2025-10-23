package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Species struct {
	ID                 int
	Name               string
	ColdStratified     bool
	StratificationDays *int
}

type Batch struct {
	ID           int
	SpeciesID    int
	SpeciesName  string
	NumCells     int
	SeedsPerCell int
	TotalSeeds   *int
	DateCreated  time.Time
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./plant_tracker.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables
	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	// Routes
	http.HandleFunc("/", listBatchesHandler)
	http.HandleFunc("/batches/new", newBatchHandler)
	http.HandleFunc("/batches/create", createBatchHandler)
	http.HandleFunc("/batches/edit", editBatchHandler)
	http.HandleFunc("/batches/update", updateBatchHandler)
	http.HandleFunc("/batches/delete", deleteBatchHandler)

	http.HandleFunc("/species", listSpeciesHandler)
	http.HandleFunc("/species/new", newSpeciesHandler)
	http.HandleFunc("/species/create", createSpeciesHandler)
	http.HandleFunc("/species/edit", editSpeciesHandler)
	http.HandleFunc("/species/update", updateSpeciesHandler)
	http.HandleFunc("/species/delete", deleteSpeciesHandler)

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() error {
	// Create species table
	speciesSchema := `
	CREATE TABLE IF NOT EXISTS species (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		cold_stratified BOOLEAN NOT NULL DEFAULT 0,
		stratification_days INTEGER
	);`

	if _, err := db.Exec(speciesSchema); err != nil {
		return err
	}

	// Check if batches table needs migration
	var columnExists bool
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM pragma_table_info('batches')
		WHERE name='species_id'
	`).Scan(&columnExists)

	if err == nil && !columnExists {
		// Old table exists, need to migrate
		// For now, just drop and recreate (in production you'd migrate data)
		db.Exec("DROP TABLE IF EXISTS batches")
	}

	// Create batches table with foreign key
	batchesSchema := `
	CREATE TABLE IF NOT EXISTS batches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		species_id INTEGER NOT NULL,
		num_cells INTEGER NOT NULL,
		seeds_per_cell INTEGER NOT NULL,
		total_seeds INTEGER,
		date_created DATE NOT NULL,
		FOREIGN KEY (species_id) REFERENCES species(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(batchesSchema)
	return err
}

// Species Database functions
func getAllSpecies() ([]Species, error) {
	rows, err := db.Query(`
		SELECT id, name, cold_stratified, stratification_days
		FROM species
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var species []Species
	for rows.Next() {
		var s Species
		if err := rows.Scan(&s.ID, &s.Name, &s.ColdStratified, &s.StratificationDays); err != nil {
			return nil, err
		}
		species = append(species, s)
	}
	return species, nil
}

func getSpeciesByID(id int) (*Species, error) {
	var s Species
	err := db.QueryRow(`
		SELECT id, name, cold_stratified, stratification_days
		FROM species
		WHERE id = ?
	`, id).Scan(&s.ID, &s.Name, &s.ColdStratified, &s.StratificationDays)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func createSpecies(s *Species) error {
	_, err := db.Exec(`
		INSERT INTO species (name, cold_stratified, stratification_days)
		VALUES (?, ?, ?)
	`, s.Name, s.ColdStratified, s.StratificationDays)
	return err
}

func updateSpecies(s *Species) error {
	_, err := db.Exec(`
		UPDATE species
		SET name = ?, cold_stratified = ?, stratification_days = ?
		WHERE id = ?
	`, s.Name, s.ColdStratified, s.StratificationDays, s.ID)
	return err
}

func deleteSpecies(id int) error {
	_, err := db.Exec("DELETE FROM species WHERE id = ?", id)
	return err
}

// Batch Database functions
func getAllBatches() ([]Batch, error) {
	rows, err := db.Query(`
		SELECT b.id, b.species_id, s.name, b.num_cells, b.seeds_per_cell, b.total_seeds, b.date_created
		FROM batches b
		JOIN species s ON b.species_id = s.id
		ORDER BY b.date_created DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []Batch
	for rows.Next() {
		var b Batch
		if err := rows.Scan(&b.ID, &b.SpeciesID, &b.SpeciesName, &b.NumCells, &b.SeedsPerCell, &b.TotalSeeds, &b.DateCreated); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, nil
}

func getBatchByID(id int) (*Batch, error) {
	var b Batch
	err := db.QueryRow(`
		SELECT b.id, b.species_id, s.name, b.num_cells, b.seeds_per_cell, b.total_seeds, b.date_created
		FROM batches b
		JOIN species s ON b.species_id = s.id
		WHERE b.id = ?
	`, id).Scan(&b.ID, &b.SpeciesID, &b.SpeciesName, &b.NumCells, &b.SeedsPerCell, &b.TotalSeeds, &b.DateCreated)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func createBatch(b *Batch) error {
	_, err := db.Exec(`
		INSERT INTO batches (species_id, num_cells, seeds_per_cell, total_seeds, date_created)
		VALUES (?, ?, ?, ?, ?)
	`, b.SpeciesID, b.NumCells, b.SeedsPerCell, b.TotalSeeds, b.DateCreated)
	return err
}

func updateBatch(b *Batch) error {
	_, err := db.Exec(`
		UPDATE batches
		SET species_id = ?, num_cells = ?, seeds_per_cell = ?, total_seeds = ?, date_created = ?
		WHERE id = ?
	`, b.SpeciesID, b.NumCells, b.SeedsPerCell, b.TotalSeeds, b.DateCreated, b.ID)
	return err
}

func deleteBatch(id int) error {
	_, err := db.Exec("DELETE FROM batches WHERE id = ?", id)
	return err
}

// Species Handlers
func listSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	species, err := getAllSpecies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("list").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Plant Tracker - Species</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 1000px; margin: 0 auto; padding: 20px; }
		h1 { color: #2d5016; }
		.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
		.nav { margin-bottom: 20px; }
		.nav a { margin-right: 15px; padding: 8px 15px; text-decoration: none; background: #f0f0f0; border-radius: 4px; }
		.nav a.active { background: #4CAF50; color: white; }
		.btn { padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; }
		.btn-primary { background: #4CAF50; color: white; }
		.btn-secondary { background: #2196F3; color: white; font-size: 12px; padding: 5px 10px; }
		.btn-danger { background: #f44336; color: white; font-size: 12px; padding: 5px 10px; }
		table { width: 100%; border-collapse: collapse; margin-top: 20px; }
		th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
		th { background-color: #4CAF50; color: white; }
		tr:hover { background-color: #f5f5f5; }
		.actions { display: flex; gap: 10px; }
		.badge { background: #2196F3; color: white; padding: 4px 8px; border-radius: 3px; font-size: 11px; }
	</style>
</head>
<body>
	<div class="nav">
		<a href="/">Batches</a>
		<a href="/species" class="active">Species</a>
	</div>

	<div class="header">
		<h1>Species</h1>
		<a href="/species/new" class="btn btn-primary">+ New Species</a>
	</div>

	{{if .}}
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Cold Stratified</th>
				<th>Days to Stratify</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			{{range .}}
			<tr>
				<td>{{.Name}}</td>
				<td>{{if .ColdStratified}}<span class="badge">Yes</span>{{else}}-{{end}}</td>
				<td>{{if .StratificationDays}}{{.StratificationDays}} days{{else}}-{{end}}</td>
				<td>
					<div class="actions">
						<a href="/species/edit?id={{.ID}}" class="btn btn-secondary">Edit</a>
						<form method="POST" action="/species/delete" style="margin: 0;">
							<input type="hidden" name="id" value="{{.ID}}">
							<button type="submit" class="btn btn-danger" onclick="return confirm('Delete this species? This will also delete all associated batches.')">Delete</button>
						</form>
					</div>
				</td>
			</tr>
			{{end}}
		</tbody>
	</table>
	{{else}}
	<p>No species yet. Create your first species to get started!</p>
	{{end}}
</body>
</html>
	`))

	tmpl.Execute(w, species)
}

func newSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>New Species - Plant Tracker</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
		h1 { color: #2d5016; }
		.nav { margin-bottom: 20px; }
		.nav a { margin-right: 15px; padding: 8px 15px; text-decoration: none; background: #f0f0f0; border-radius: 4px; }
		form { background: #f9f9f9; padding: 20px; border-radius: 5px; }
		.form-group { margin-bottom: 15px; }
		label { display: block; margin-bottom: 5px; font-weight: bold; }
		input[type="text"], input[type="number"] { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
		input[type="checkbox"] { width: auto; margin-right: 5px; }
		.checkbox-group { display: flex; align-items: center; }
		.btn { padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; }
		.btn-primary { background: #4CAF50; color: white; }
		.btn-secondary { background: #999; color: white; margin-left: 10px; }
		.form-actions { margin-top: 20px; }
		#stratification_days_group { display: none; }
	</style>
	<script>
		function toggleStratificationDays() {
			const checkbox = document.getElementById('cold_stratified');
			const daysGroup = document.getElementById('stratification_days_group');
			daysGroup.style.display = checkbox.checked ? 'block' : 'none';
		}
	</script>
</head>
<body>
	<div class="nav">
		<a href="/">Batches</a>
		<a href="/species">Species</a>
	</div>

	<h1>New Species</h1>
	<form method="POST" action="/species/create">
		<div class="form-group">
			<label for="name">Species Name *</label>
			<input type="text" id="name" name="name" required>
		</div>

		<div class="form-group">
			<div class="checkbox-group">
				<input type="checkbox" id="cold_stratified" name="cold_stratified" onchange="toggleStratificationDays()">
				<label for="cold_stratified" style="margin-bottom: 0;">Requires Cold Stratification</label>
			</div>
		</div>

		<div class="form-group" id="stratification_days_group">
			<label for="stratification_days">Days to Stratify</label>
			<input type="number" id="stratification_days" name="stratification_days" min="1">
		</div>

		<div class="form-actions">
			<button type="submit" class="btn btn-primary">Create Species</button>
			<a href="/species" class="btn btn-secondary">Cancel</a>
		</div>
	</form>
</body>
</html>
	`))

	tmpl.Execute(w, nil)
}

func createSpeciesHandler(w http.ResponseWriter, r *http.Request) {
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

	species := &Species{
		Name:               r.FormValue("name"),
		ColdStratified:     coldStratified,
		StratificationDays: stratificationDays,
	}

	if err := createSpecies(species); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/species", http.StatusSeeOther)
}

func editSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	species, err := getSpeciesByID(id)
	if err != nil {
		http.Error(w, "Species not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Edit Species - Plant Tracker</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
		h1 { color: #2d5016; }
		.nav { margin-bottom: 20px; }
		.nav a { margin-right: 15px; padding: 8px 15px; text-decoration: none; background: #f0f0f0; border-radius: 4px; }
		form { background: #f9f9f9; padding: 20px; border-radius: 5px; }
		.form-group { margin-bottom: 15px; }
		label { display: block; margin-bottom: 5px; font-weight: bold; }
		input[type="text"], input[type="number"] { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
		input[type="checkbox"] { width: auto; margin-right: 5px; }
		.checkbox-group { display: flex; align-items: center; }
		.btn { padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; }
		.btn-primary { background: #4CAF50; color: white; }
		.btn-secondary { background: #999; color: white; margin-left: 10px; }
		.form-actions { margin-top: 20px; }
		#stratification_days_group { display: {{if .ColdStratified}}block{{else}}none{{end}}; }
	</style>
	<script>
		function toggleStratificationDays() {
			const checkbox = document.getElementById('cold_stratified');
			const daysGroup = document.getElementById('stratification_days_group');
			daysGroup.style.display = checkbox.checked ? 'block' : 'none';
		}
	</script>
</head>
<body>
	<div class="nav">
		<a href="/">Batches</a>
		<a href="/species">Species</a>
	</div>

	<h1>Edit Species</h1>
	<form method="POST" action="/species/update">
		<input type="hidden" name="id" value="{{.ID}}">

		<div class="form-group">
			<label for="name">Species Name *</label>
			<input type="text" id="name" name="name" value="{{.Name}}" required>
		</div>

		<div class="form-group">
			<div class="checkbox-group">
				<input type="checkbox" id="cold_stratified" name="cold_stratified" {{if .ColdStratified}}checked{{end}} onchange="toggleStratificationDays()">
				<label for="cold_stratified" style="margin-bottom: 0;">Requires Cold Stratification</label>
			</div>
		</div>

		<div class="form-group" id="stratification_days_group">
			<label for="stratification_days">Days to Stratify</label>
			<input type="number" id="stratification_days" name="stratification_days" value="{{if .StratificationDays}}{{.StratificationDays}}{{end}}" min="1">
		</div>

		<div class="form-actions">
			<button type="submit" class="btn btn-primary">Update Species</button>
			<a href="/species" class="btn btn-secondary">Cancel</a>
		</div>
	</form>
</body>
</html>
	`))

	tmpl.Execute(w, species)
}

func updateSpeciesHandler(w http.ResponseWriter, r *http.Request) {
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

	species := &Species{
		ID:                 id,
		Name:               r.FormValue("name"),
		ColdStratified:     coldStratified,
		StratificationDays: stratificationDays,
	}

	if err := updateSpecies(species); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/species", http.StatusSeeOther)
}

func deleteSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/species", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))

	if err := deleteSpecies(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/species", http.StatusSeeOther)
}

// Batch Handlers
func listBatchesHandler(w http.ResponseWriter, r *http.Request) {
	batches, err := getAllBatches()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("list").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Plant Tracker - Germination Batches</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 1000px; margin: 0 auto; padding: 20px; }
		h1 { color: #2d5016; }
		.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
		.nav { margin-bottom: 20px; }
		.nav a { margin-right: 15px; padding: 8px 15px; text-decoration: none; background: #f0f0f0; border-radius: 4px; }
		.nav a.active { background: #4CAF50; color: white; }
		.btn { padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; }
		.btn-primary { background: #4CAF50; color: white; }
		.btn-secondary { background: #2196F3; color: white; font-size: 12px; padding: 5px 10px; }
		.btn-danger { background: #f44336; color: white; font-size: 12px; padding: 5px 10px; }
		table { width: 100%; border-collapse: collapse; margin-top: 20px; }
		th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
		th { background-color: #4CAF50; color: white; }
		tr:hover { background-color: #f5f5f5; }
		.actions { display: flex; gap: 10px; }
	</style>
</head>
<body>
	<div class="nav">
		<a href="/" class="active">Batches</a>
		<a href="/species">Species</a>
	</div>

	<div class="header">
		<h1>Germination Batches</h1>
		<a href="/batches/new" class="btn btn-primary">+ New Batch</a>
	</div>

	{{if .}}
	<table>
		<thead>
			<tr>
				<th>Species</th>
				<th>Cells</th>
				<th>Seeds/Cell</th>
				<th>Total Seeds</th>
				<th>Date</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			{{range .}}
			<tr>
				<td>{{.SpeciesName}}</td>
				<td>{{.NumCells}}</td>
				<td>{{.SeedsPerCell}}</td>
				<td>{{if .TotalSeeds}}{{.TotalSeeds}}{{else}}-{{end}}</td>
				<td>{{.DateCreated.Format "2006-01-02"}}</td>
				<td>
					<div class="actions">
						<a href="/batches/edit?id={{.ID}}" class="btn btn-secondary">Edit</a>
						<form method="POST" action="/batches/delete" style="margin: 0;">
							<input type="hidden" name="id" value="{{.ID}}">
							<button type="submit" class="btn btn-danger" onclick="return confirm('Delete this batch?')">Delete</button>
						</form>
					</div>
				</td>
			</tr>
			{{end}}
		</tbody>
	</table>
	{{else}}
	<p>No batches yet. Create your first batch to get started!</p>
	<p>Note: You need to <a href="/species">create at least one species</a> before you can create a batch.</p>
	{{end}}
</body>
</html>
	`))

	tmpl.Execute(w, batches)
}

func newBatchHandler(w http.ResponseWriter, r *http.Request) {
	species, err := getAllSpecies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Species   []Species
		TodayDate string
	}{
		Species:   species,
		TodayDate: time.Now().Format("2006-01-02"),
	}

	tmpl := template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>New Batch - Plant Tracker</title>
	<link href="https://cdn.jsdelivr.net/npm/tom-select@2.3.1/dist/css/tom-select.css" rel="stylesheet">
	<style>
		body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
		h1 { color: #2d5016; }
		.nav { margin-bottom: 20px; }
		.nav a { margin-right: 15px; padding: 8px 15px; text-decoration: none; background: #f0f0f0; border-radius: 4px; }
		form { background: #f9f9f9; padding: 20px; border-radius: 5px; }
		.form-group { margin-bottom: 15px; }
		label { display: block; margin-bottom: 5px; font-weight: bold; }
		input, select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
		.btn { padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; }
		.btn-primary { background: #4CAF50; color: white; }
		.btn-secondary { background: #999; color: white; margin-left: 10px; }
		.form-actions { margin-top: 20px; }
		.species-detail { font-size: 12px; color: #666; }
		/* Hide Tom Select input when item is selected */
		.ts-control .item ~ input { display: none !important; }
	</style>
</head>
<body>
	<div class="nav">
		<a href="/">Batches</a>
		<a href="/species">Species</a>
	</div>

	<h1>New Germination Batch</h1>

	{{if .Species}}
	<form method="POST" action="/batches/create">
		<div class="form-group">
			<label for="species_id">Species *</label>
			<select id="species_id" name="species_id" required>
				<option value="">Select a species...</option>
				{{range .Species}}
				<option value="{{.ID}}"
					data-stratified="{{.ColdStratified}}"
					data-days="{{if .StratificationDays}}{{.StratificationDays}}{{end}}">
					{{.Name}}{{if .ColdStratified}} 🧊{{end}}{{if .StratificationDays}} ({{.StratificationDays}} days){{end}}
				</option>
				{{end}}
			</select>
		</div>

		<div class="form-group">
			<label for="num_cells">Number of Cells *</label>
			<input type="number" id="num_cells" name="num_cells" min="1" required>
		</div>

		<div class="form-group">
			<label for="seeds_per_cell">Seeds per Cell *</label>
			<input type="number" id="seeds_per_cell" name="seeds_per_cell" min="1" required>
		</div>

		<div class="form-group">
			<label for="total_seeds">Total Seeds (optional)</label>
			<input type="number" id="total_seeds" name="total_seeds" min="1">
		</div>

		<div class="form-group">
			<label for="date_created">Date *</label>
			<input type="date" id="date_created" name="date_created" value="{{.TodayDate}}" required>
		</div>

		<div class="form-actions">
			<button type="submit" class="btn btn-primary">Create Batch</button>
			<a href="/" class="btn btn-secondary">Cancel</a>
		</div>
	</form>
	{{else}}
	<p>You need to <a href="/species/new">create at least one species</a> before you can create a batch.</p>
	<a href="/species/new" class="btn btn-primary">Create Species</a>
	{{end}}

	<script src="https://cdn.jsdelivr.net/npm/tom-select@2.3.1/dist/js/tom-select.complete.min.js"></script>
	<script>
		{{if .Species}}
		new TomSelect('#species_id', {
			placeholder: 'Search for a species...',
			create: false,
			maxItems: 1,
			maxOptions: 50,
			closeAfterSelect: true
		});
		{{end}}
	</script>
</body>
</html>
	`))

	tmpl.Execute(w, data)
}

func createBatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	speciesID, _ := strconv.Atoi(r.FormValue("species_id"))
	numCells, _ := strconv.Atoi(r.FormValue("num_cells"))
	seedsPerCell, _ := strconv.Atoi(r.FormValue("seeds_per_cell"))
	dateCreated, _ := time.Parse("2006-01-02", r.FormValue("date_created"))

	var totalSeeds *int
	if ts := r.FormValue("total_seeds"); ts != "" {
		val, _ := strconv.Atoi(ts)
		totalSeeds = &val
	}

	batch := &Batch{
		SpeciesID:    speciesID,
		NumCells:     numCells,
		SeedsPerCell: seedsPerCell,
		TotalSeeds:   totalSeeds,
		DateCreated:  dateCreated,
	}

	if err := createBatch(batch); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editBatchHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	batch, err := getBatchByID(id)
	if err != nil {
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}

	species, err := getAllSpecies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Batch   *Batch
		Species []Species
	}{
		Batch:   batch,
		Species: species,
	}

	tmpl := template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Edit Batch - Plant Tracker</title>
	<link href="https://cdn.jsdelivr.net/npm/tom-select@2.3.1/dist/css/tom-select.css" rel="stylesheet">
	<style>
		body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
		h1 { color: #2d5016; }
		.nav { margin-bottom: 20px; }
		.nav a { margin-right: 15px; padding: 8px 15px; text-decoration: none; background: #f0f0f0; border-radius: 4px; }
		form { background: #f9f9f9; padding: 20px; border-radius: 5px; }
		.form-group { margin-bottom: 15px; }
		label { display: block; margin-bottom: 5px; font-weight: bold; }
		input, select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
		.btn { padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; }
		.btn-primary { background: #4CAF50; color: white; }
		.btn-secondary { background: #999; color: white; margin-left: 10px; }
		.form-actions { margin-top: 20px; }
		.species-detail { font-size: 12px; color: #666; }
		/* Hide Tom Select input when item is selected */
		.ts-control .item ~ input { display: none !important; }
	</style>
</head>
<body>
	<div class="nav">
		<a href="/">Batches</a>
		<a href="/species">Species</a>
	</div>

	<h1>Edit Germination Batch</h1>
	<form method="POST" action="/batches/update">
		<input type="hidden" name="id" value="{{.Batch.ID}}">

		<div class="form-group">
			<label for="species_id">Species *</label>
			<select id="species_id" name="species_id" required>
				{{range .Species}}
				<option value="{{.ID}}" {{if eq .ID $.Batch.SpeciesID}}selected{{end}}
					data-stratified="{{.ColdStratified}}"
					data-days="{{if .StratificationDays}}{{.StratificationDays}}{{end}}">
					{{.Name}}{{if .ColdStratified}} 🧊{{end}}{{if .StratificationDays}} ({{.StratificationDays}} days){{end}}
				</option>
				{{end}}
			</select>
		</div>

		<div class="form-group">
			<label for="num_cells">Number of Cells *</label>
			<input type="number" id="num_cells" name="num_cells" value="{{.Batch.NumCells}}" min="1" required>
		</div>

		<div class="form-group">
			<label for="seeds_per_cell">Seeds per Cell *</label>
			<input type="number" id="seeds_per_cell" name="seeds_per_cell" value="{{.Batch.SeedsPerCell}}" min="1" required>
		</div>

		<div class="form-group">
			<label for="total_seeds">Total Seeds (optional)</label>
			<input type="number" id="total_seeds" name="total_seeds" value="{{if .Batch.TotalSeeds}}{{.Batch.TotalSeeds}}{{end}}" min="1">
		</div>

		<div class="form-group">
			<label for="date_created">Date *</label>
			<input type="date" id="date_created" name="date_created" value="{{.Batch.DateCreated.Format "2006-01-02"}}" required>
		</div>

		<div class="form-actions">
			<button type="submit" class="btn btn-primary">Update Batch</button>
			<a href="/" class="btn btn-secondary">Cancel</a>
		</div>
	</form>

	<script src="https://cdn.jsdelivr.net/npm/tom-select@2.3.1/dist/js/tom-select.complete.min.js"></script>
	<script>
		new TomSelect('#species_id', {
			placeholder: 'Search for a species...',
			create: false,
			maxItems: 1,
			maxOptions: 50,
			closeAfterSelect: true
		});
	</script>
</body>
</html>
	`))

	tmpl.Execute(w, data)
}

func updateBatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	id, _ := strconv.Atoi(r.FormValue("id"))
	speciesID, _ := strconv.Atoi(r.FormValue("species_id"))
	numCells, _ := strconv.Atoi(r.FormValue("num_cells"))
	seedsPerCell, _ := strconv.Atoi(r.FormValue("seeds_per_cell"))
	dateCreated, _ := time.Parse("2006-01-02", r.FormValue("date_created"))

	var totalSeeds *int
	if ts := r.FormValue("total_seeds"); ts != "" {
		val, _ := strconv.Atoi(ts)
		totalSeeds = &val
	}

	batch := &Batch{
		ID:           id,
		SpeciesID:    speciesID,
		NumCells:     numCells,
		SeedsPerCell: seedsPerCell,
		TotalSeeds:   totalSeeds,
		DateCreated:  dateCreated,
	}

	if err := updateBatch(batch); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteBatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))

	if err := deleteBatch(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
