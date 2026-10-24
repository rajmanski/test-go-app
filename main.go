package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		log.Fatal("Brak zmiennej środowiskowej DATABASE_URL")
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Błąd otwierania bazy: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Błąd połączenia z bazą: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS visits (id SERIAL PRIMARY KEY, count INT)`)
	if err != nil {
		log.Fatalf("Błąd tworzenia tabeli: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var count int
		err := db.QueryRow("SELECT COALESCE(MAX(count), 0) + 1 FROM visits").Scan(&count)
		if err == nil {
			db.Exec("INSERT INTO visits (count) VALUES ($1)", count)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h1>🚀 Cześć z Go na Proxmoxie!</h1>")
		fmt.Fprintf(w, "<p>To zapytanie numer: <b>%d</b> obsłużone przez bazę PostgreSQL (`main_db`).</p>", count)
		fmt.Fprintf(w, "<p style='color: green;'>Status: Połączenie z bazą i siecią Dockera działa perfekcyjnie!</p>")
	})

	log.Println("Serwer w Go wystartował na porcie 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
