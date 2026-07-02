package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

var cars = map[string]string{
	"id1": "Renault Logan",
	"id2": "Renault Duster",
	"id3": "BMW X6",
	"id4": "BMW M5",
	"id5": "VW Passat",
	"id6": "VW Jetta",
	"id7": "Audi A4",
	"id8": "Audi Q7",
}

func modelHandle(rw http.ResponseWriter, r *http.Request) {
	car := strings.ToLower(chi.URLParam(r, "brand") + ` ` +
		chi.URLParam(r, "model"))
	for _, c := range cars {
		if strings.ToLower(c) == car {
			io.WriteString(rw, c)
			return
		}
	}
	http.Error(rw, "unknown model: "+car, http.StatusNotFound)
}

func CarRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/cars/{brand}/{model}", modelHandle) // GET /cars/renault/duster
	return r
}

func main() {
	http.ListenAndServe(":5050", CarRouter())
}
