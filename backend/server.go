package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type SalaryEntry struct {
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Year   int     `json:"year"`
}

type Server struct {
	mu      sync.RWMutex
	names   []string
	entries []SalaryEntry
	csvFile string
}

func NewServer() *Server {
	s := &Server{
		csvFile: "salary_entries.csv",
	}
	s.loadData()
	return s
}

func (s *Server) loadData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.csvFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Printf("Error opening file: %v", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV: %v", err)
		return
	}

	s.entries = []SalaryEntry{}
	s.names = []string{}

	for _, record := range records {
		if len(record) != 3 {
			continue
		}

		salary, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			continue
		}

		year, err := strconv.Atoi(record[2])
		if err != nil {
			continue
		}

		entry := SalaryEntry{
			Name:   record[0],
			Salary: salary,
			Year:   year,
		}

		s.entries = append(s.entries, entry)

		// Add unique names
		found := false
		for _, name := range s.names {
			if name == record[0] {
				found = true
				break
			}
		}
		if !found {
			s.names = append(s.names, record[0])
		}
	}
}

func (s *Server) saveEntry(entry SalaryEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		entry.Name,
		fmt.Sprintf("%.2f NOK", entry.Salary),
		strconv.Itoa(entry.Year),
	}
	if err := writer.Write(record); err != nil {
		return err
	}

	s.entries = append(s.entries, entry)

	// Add to names if not exists
	found := false
	for _, name := range s.names {
		if name == entry.Name {
			found = true
			break
		}
	}
	if !found {
		s.names = append(s.names, entry.Name)
	}

	return nil
}

func (s *Server) handleCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *Server) handleData(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := struct {
		Names   []string      `json:"names"`
		Entries []SalaryEntry `json:"entries"`
	}{
		Names:   s.names,
		Entries: s.entries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleNames(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for _, name := range s.names {
		if name == req.Name {
			found = true
			break
		}
	}

	if !found {
		s.names = append(s.names, req.Name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.names)
}

func (s *Server) handleEntries(w http.ResponseWriter, r *http.Request) {
	var entry SalaryEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.saveEntry(entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.entries)
}

func main() {
	server := NewServer()

	http.HandleFunc("/data", server.handleCORS(server.handleData))
	http.HandleFunc("/names", server.handleCORS(server.handleNames))
	http.HandleFunc("/entries", server.handleCORS(server.handleEntries))

	fmt.Println("Backend server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
