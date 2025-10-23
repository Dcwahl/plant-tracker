package db

import (
	"database/sql"

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
	DateCreated  string
}

var DB *sql.DB

func Init() error {
	var err error
	DB, err = sql.Open("sqlite3", "./plant_tracker.db")
	if err != nil {
		return err
	}

	// Create species table
	speciesSchema := `
	CREATE TABLE IF NOT EXISTS species (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		cold_stratified BOOLEAN NOT NULL DEFAULT 0,
		stratification_days INTEGER
	);`

	if _, err := DB.Exec(speciesSchema); err != nil {
		return err
	}

	// Check if batches table needs migration
	var columnExists bool
	err = DB.QueryRow(`
		SELECT COUNT(*)
		FROM pragma_table_info('batches')
		WHERE name='species_id'
	`).Scan(&columnExists)

	if err == nil && !columnExists {
		// Old table exists, need to migrate
		DB.Exec("DROP TABLE IF EXISTS batches")
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

	_, err = DB.Exec(batchesSchema)
	return err
}
