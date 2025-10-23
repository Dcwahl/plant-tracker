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
	return &s, nil
}

func CreateSpecies(s *Species) error {
	_, err := DB.Exec(`
		INSERT INTO species (name, cold_stratified, stratification_days)
		VALUES (?, ?, ?)
	`, s.Name, s.ColdStratified, s.StratificationDays)
	return err
}

func UpdateSpecies(s *Species) error {
	_, err := DB.Exec(`
		UPDATE species
		SET name = ?, cold_stratified = ?, stratification_days = ?
		WHERE id = ?
	`, s.Name, s.ColdStratified, s.StratificationDays, s.ID)
	return err
}

func DeleteSpecies(id int) error {
	_, err := DB.Exec("DELETE FROM species WHERE id = ?", id)
	return err
}
