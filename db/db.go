package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Species struct {
	ID                  int
	Name                string
	StratificationSteps []StratificationStep
	// Deprecated fields - kept for migration only
	ColdStratified     bool
	StratificationDays *int
}

type StratificationStep struct {
	ID        int
	SpeciesID int
	StepOrder int
	Type      string // "Cold", "Warm", or "Scarification"
	Moist     bool
	Days      int
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

	if _, err = DB.Exec(batchesSchema); err != nil {
		return err
	}

	// Create stratification_steps table
	stratificationSchema := `
	CREATE TABLE IF NOT EXISTS stratification_steps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		species_id INTEGER NOT NULL,
		step_order INTEGER NOT NULL,
		type TEXT NOT NULL,
		moist BOOLEAN NOT NULL DEFAULT 0,
		days INTEGER NOT NULL,
		FOREIGN KEY (species_id) REFERENCES species(id) ON DELETE CASCADE
	);`

	if _, err = DB.Exec(stratificationSchema); err != nil {
		return err
	}

	// Run migration to convert old stratification data
	if err = migrateStratificationData(); err != nil {
		return err
	}

	return nil
}

// migrateStratificationData converts old ColdStratified/StratificationDays to new stratification_steps
func migrateStratificationData() error {
	// Check if there's any old data to migrate
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM species
		WHERE cold_stratified = 1
		AND id NOT IN (SELECT DISTINCT species_id FROM stratification_steps)
	`).Scan(&count)

	if err != nil || count == 0 {
		return err // Nothing to migrate or error checking
	}

	// Get species with old stratification data that haven't been migrated
	rows, err := DB.Query(`
		SELECT id, stratification_days
		FROM species
		WHERE cold_stratified = 1
		AND id NOT IN (SELECT DISTINCT species_id FROM stratification_steps)
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Migrate each species
	for rows.Next() {
		var speciesID int
		var days *int
		if err := rows.Scan(&speciesID, &days); err != nil {
			return err
		}

		// Create a Cold, Moist stratification step with the specified days
		daysValue := 30 // default
		if days != nil && *days > 0 {
			daysValue = *days
		}

		_, err = DB.Exec(`
			INSERT INTO stratification_steps (species_id, step_order, type, moist, days)
			VALUES (?, 0, 'Cold', 1, ?)
		`, speciesID, daysValue)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}
