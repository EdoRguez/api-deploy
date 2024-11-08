package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/EdoRguez/api-deploy/models"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	rc := routesConfig(r)

	systems := []string{"navigation", "communications", "life_support", "engines", "deflector_shield"}
	codes := []string{"NAV-01", "COM-02", "LIFE-03", "ENG-04", "SHLD-05"}
	systemIdx := 3

	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		model := models.Status{
			DamagedSystem: systems[systemIdx],
		}

		json.NewEncoder(w).Encode(&model)
	}).Methods("GET")

	r.HandleFunc("/repair-bay", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		tmpl, err := template.ParseFiles("./templates/template.html")
		if err != nil {
			log.Fatalf("Error parsing template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		p := models.Status{DamagedSystem: codes[systemIdx]}

		err = tmpl.Execute(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods("GET")

	r.HandleFunc("/teapot", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("I'm a teapot"))
	}).Methods("POST")

	r.HandleFunc("/set-system-idx/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if id < 0 || id > len(systems)-1 {
			http.Error(w, "ID doesn't exist", http.StatusInternalServerError)
			return
		}

		systemIdx = id

		w.WriteHeader(http.StatusNoContent)
	}).Methods("PUT")

	s := &http.Server{
		Addr:         ":3000",
		Handler:      rc,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		fmt.Println("Starting server on port 3000")

		err := s.ListenAndServe()
		if err != nil {
			fmt.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	// Block until a signal is received.
	sig := <-c
	fmt.Println("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}

func routesConfig(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Check if the request is for CORS preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass down the request to the next middleware (or final handler)
		next.ServeHTTP(w, r)
	})
}
