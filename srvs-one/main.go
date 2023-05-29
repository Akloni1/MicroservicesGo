package main

import (
	"encoding/json"
	"net/http"
	"strconv"

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

var pizzas = []*Pizzas{
	{CountAvailable: 20, Name: "Pepperoni"},
	{CountAvailable: 10, Name: "Meat"},
	{CountAvailable: 5, Name: "4 cheeses"},
	{CountAvailable: 34, Name: "Margarita"},
}

func GetPizzas(w http.ResponseWriter, r *http.Request) {
	countMore, _ := strconv.Atoi(r.URL.Query().Get("countMore"))
	countLess, _ := strconv.Atoi(r.URL.Query().Get("countLess"))

	if countMore != 0 || countLess != 0 {
		filteredPizzasMore := []Pizzas{}
		filteredPizzas := []Pizzas{}
		if countMore != 0 {
			for _, pizza := range pizzas {
				if pizza.CountAvailable > countMore {
					filteredPizzasMore = append(filteredPizzasMore, *pizza)
				}
			}
		} else {
			for _, pizza := range pizzas {
				filteredPizzasMore = append(filteredPizzasMore, *pizza)
			}
		}

		if countLess != 0 {
			for _, pizza := range filteredPizzasMore {
				if pizza.CountAvailable < countLess {
					filteredPizzas = append(filteredPizzas, pizza)
				}
			}
		} else {
			filteredPizzas = filteredPizzasMore
		}
		json.NewEncoder(w).Encode(filteredPizzas)
	} else {
		json.NewEncoder(w).Encode(pizzas)
	}
}

func ReduceCountpizzas(w http.ResponseWriter, r *http.Request) {
	namePizza := chi.URLParam(r, "namePizza")

	var reductionCount struct {
		CountAvailable int `json:"countAvailable"`
	}

	err := json.NewDecoder(r.Body).Decode(&reductionCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i, pizza := range pizzas {
		if namePizza == pizza.Name {
			pizzas[i].CountAvailable = pizzas[i].CountAvailable - reductionCount.CountAvailable
			json.NewEncoder(w).Encode(pizzas[i])
			return
		}
	}
	http.NotFound(w, r)
}

func main() {

	logger := httplog.NewLogger("service-one", httplog.Options{
		JSON: true,
	})

	r := chi.NewRouter()
	r.Use(chiprometheus.NewMiddleware("service-one"))
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	r.Handle("/metrics", promhttp.Handler())

	r.Get("/pizzas", GetPizzas)
	r.Post("/pizzas/{namePizza}", ReduceCountpizzas)

	http.ListenAndServe(":8081", r)
}
