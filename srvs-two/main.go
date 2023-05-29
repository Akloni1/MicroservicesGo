package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	chiprometheus "github.com/nathan-jones/chi-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Pizzas struct {
	CountAvailable int    `json:"countAvailable"`
	Name           string `json:"name"`
}

type OrderPizza struct {
	CountAvailable int `json:"countAvailable"`
}

func GetPizzasMany() ([]Pizzas, error) {
	resp, err := http.Get("http://host.docker.internal:8081/pizzas?countMore=10")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var response []Pizzas
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func ReserveAudience(name string) (*Pizzas, error) {
	order := OrderPizza{
		CountAvailable: 1,
	}

	jsonBytes, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://host.docker.internal:8081/pizzas/"+name, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var pizza Pizzas
	err = json.NewDecoder(resp.Body).Decode(&pizza)
	if err != nil {
		return nil, err
	}

	return &pizza, nil
}

func main() {

	logger := httplog.NewLogger("service-two", httplog.Options{
		JSON: true,
	})

	r := chi.NewRouter()
	r.Use(chiprometheus.NewMiddleware("service-two"))
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Handle("/metrics", promhttp.Handler())

	r.Get("/pizzas-many", func(w http.ResponseWriter, r *http.Request) {
		pizzas, err := GetPizzasMany()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(pizzas)
	})

	r.Post("/order-pizza/{name}", func(w http.ResponseWriter, r *http.Request) {

		name := chi.URLParam(r, "name")
		audience, err := ReserveAudience(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if audience == nil {
			http.NotFound(w, r)
			return
		}

		json.NewEncoder(w).Encode(audience)
	})

	http.ListenAndServe(":8082", r)
}
