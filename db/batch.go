package db

func GetAllBatches() ([]Batch, error) {
	rows, err := DB.Query(`
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

func GetBatchByID(id int) (*Batch, error) {
	var b Batch
	err := DB.QueryRow(`
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

func CreateBatch(b *Batch) error {
	_, err := DB.Exec(`
		INSERT INTO batches (species_id, num_cells, seeds_per_cell, total_seeds, date_created)
		VALUES (?, ?, ?, ?, ?)
	`, b.SpeciesID, b.NumCells, b.SeedsPerCell, b.TotalSeeds, b.DateCreated)
	return err
}

func UpdateBatch(b *Batch) error {
	_, err := DB.Exec(`
		UPDATE batches
		SET species_id = ?, num_cells = ?, seeds_per_cell = ?, total_seeds = ?, date_created = ?
		WHERE id = ?
	`, b.SpeciesID, b.NumCells, b.SeedsPerCell, b.TotalSeeds, b.DateCreated, b.ID)
	return err
}

func DeleteBatch(id int) error {
	_, err := DB.Exec("DELETE FROM batches WHERE id = ?", id)
	return err
}
