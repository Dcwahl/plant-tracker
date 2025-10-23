package main

import (
	"fmt"
	"log"
	"net/http"
	"plant-tracker/db"
	"plant-tracker/handlers"
)

func main() {
	// Initialize database
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	// Batch routes
	http.HandleFunc("/", handlers.ListBatches)
	http.HandleFunc("/batches/new", handlers.NewBatch)
	http.HandleFunc("/batches/create", handlers.CreateBatch)
	http.HandleFunc("/batches/edit", handlers.EditBatch)
	http.HandleFunc("/batches/update", handlers.UpdateBatch)
	http.HandleFunc("/batches/delete", handlers.DeleteBatch)

	// Species routes
	http.HandleFunc("/species", handlers.ListSpecies)
	http.HandleFunc("/species/new", handlers.NewSpecies)
	http.HandleFunc("/species/create", handlers.CreateSpecies)
	http.HandleFunc("/species/edit", handlers.EditSpecies)
	http.HandleFunc("/species/update", handlers.UpdateSpecies)
	http.HandleFunc("/species/delete", handlers.DeleteSpecies)

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
