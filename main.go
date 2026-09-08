package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	client, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	service := NewWorkspaceService(client)
	http.HandleFunc("/admin/onboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in struct{ Tenant, AccountID, Name string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.Onboard(in.Tenant, in.AccountID, in.Name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	http.HandleFunc("/online", func(w http.ResponseWriter, r *http.Request) {
		members, err := service.Online(r.URL.Query().Get("tenant"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"members": members})
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
