package handlers

import (
	"database/sql"
	"net/http"
)

func PingHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "нет настроект базы данных", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

}
