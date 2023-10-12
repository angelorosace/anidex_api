package handlers

import (
	responses "anidex_api/http/responses"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type Stat struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")

	// Parse and handle URL query parameters
	queryParams := r.URL.Query()

	// Extract specific query parameters
	table := queryParams.Get("table")
	groupBy := queryParams.Get("groupBy")

	if table == "" || groupBy == "" {
		resp, err := responses.MissingURLParametersResponse(w)
		if err != nil {
			return
		}
		w.Write(resp)
		return
	}

	//retrieve DB from context
	db := r.Context().Value("db").(*sql.DB)

	res, err := db.Query(fmt.Sprintf("SELECT %s,COUNT(*) from %s group by %s", groupBy, table, groupBy))
	if err != nil {
		resp, err := responses.MySqlError(w, err)
		if err != nil {
			return
		}
		w.Write(resp)
		return
	}
	defer res.Close()

	// Create a slice to hold the results
	var stats []Stat

	// Iterate through the rows and scan data into the slice
	for res.Next() {
		var stat Stat
		err := res.Scan(&stat.ID, &stat.Count)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stats = append(stats, stat)
	}

	// Check for errors from iterating over rows
	if err := res.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := responses.HttpResponse{
		Data:    stats,
		Message: "Stats successfully computed",
		Status:  http.StatusOK,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}
