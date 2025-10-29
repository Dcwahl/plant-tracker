package db

func GetAllSpecies() ([]Species, error) {
	rows, err := DB.Query(`
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

		// Load stratification steps
		steps, err := GetStratificationSteps(s.ID)
		if err != nil {
			return nil, err
		}
		s.StratificationSteps = steps

		species = append(species, s)
	}
	return species, nil
}

func GetSpeciesByID(id int) (*Species, error) {
	var s Species
	err := DB.QueryRow(`
		SELECT id, name, cold_stratified, stratification_days
		FROM species
		WHERE id = ?
	`, id).Scan(&s.ID, &s.Name, &s.ColdStratified, &s.StratificationDays)
	if err != nil {
		return nil, err
	}

	// Load stratification steps
	steps, err := GetStratificationSteps(s.ID)
	if err != nil {
		return nil, err
	}
	s.StratificationSteps = steps

	return &s, nil
}

func CreateSpecies(s *Species) error {
	result, err := DB.Exec(`
		INSERT INTO species (name, cold_stratified, stratification_days)
		VALUES (?, 0, NULL)
	`, s.Name)
	if err != nil {
		return err
	}

	// Get the ID of the newly created species
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	s.ID = int(id)

	// Save stratification steps
	return SaveStratificationSteps(s.ID, s.StratificationSteps)
}

func UpdateSpecies(s *Species) error {
	_, err := DB.Exec(`
		UPDATE species
		SET name = ?
		WHERE id = ?
	`, s.Name, s.ID)
	if err != nil {
		return err
	}

	// Update stratification steps (delete old ones and insert new ones)
	return SaveStratificationSteps(s.ID, s.StratificationSteps)
}

func DeleteSpecies(id int) error {
	_, err := DB.Exec("DELETE FROM species WHERE id = ?", id)
	return err
}

// GetStratificationSteps retrieves all stratification steps for a species
func GetStratificationSteps(speciesID int) ([]StratificationStep, error) {
	rows, err := DB.Query(`
		SELECT id, species_id, step_order, type, moist, days
		FROM stratification_steps
		WHERE species_id = ?
		ORDER BY step_order ASC
	`, speciesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []StratificationStep
	for rows.Next() {
		var step StratificationStep
		if err := rows.Scan(&step.ID, &step.SpeciesID, &step.StepOrder, &step.Type, &step.Moist, &step.Days); err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// SaveStratificationSteps deletes old steps and saves new ones for a species
func SaveStratificationSteps(speciesID int, steps []StratificationStep) error {
	// Delete existing steps
	_, err := DB.Exec("DELETE FROM stratification_steps WHERE species_id = ?", speciesID)
	if err != nil {
		return err
	}

	// Insert new steps
	for i, step := range steps {
		_, err := DB.Exec(`
			INSERT INTO stratification_steps (species_id, step_order, type, moist, days)
			VALUES (?, ?, ?, ?, ?)
		`, speciesID, i, step.Type, step.Moist, step.Days)
		if err != nil {
			return err
		}
	}

	return nil
}
