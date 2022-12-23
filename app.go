// app.go

package main

import (
	"database/sql"
	"fmt"

	// tom: for Initialize

	"log"

	// tom: for route handlers
	"encoding/json"
	"net/http"
	"strconv"

	// tom: go get required
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/r3labs/sse/v2"

	"github.com/samber/lo"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// tom: initial function is empty, it's filled afterwards
// func (a *App) Initialize(user, password, dbname string) { }

// tom: added "sslmode=disable" to connection string
func (a *App) Initialize(user, password, dbname string) {
	// connectionString :=
	// 	fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	// var err error
	// a.DB, err = sql.Open("postgres", connectionString)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	a.Router = mux.NewRouter()

	// tom: this line is added after initializeRoutes is created later on
	a.initializeRoutes()
}

// tom: initial version
// func (a *App) Run(addr string) { }
// improved version
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) initializeRoutes() {

	type Score struct {
		StudentId string  `json:"studentId"`
		Exam      int     `json:"exam"`
		Score     float32 `json:"score"`
	}

	var allScores []Score

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		client := sse.NewClient("http://live-test-scores.herokuapp.com/scores")
		client.Subscribe("messages", func(msg *sse.Event) {
			// Got some data!
			var scores Score
			err := json.Unmarshal(msg.Data, &scores)
			if err != nil {
				fmt.Println("error:", err)
			}
			// fmt.Printf("%+v", scores)
			fmt.Print(scores)
			allScores = append(allScores, scores)
			fmt.Println(len(allScores))
		})
	}()

	a.Router.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, allScores)
	}).Methods("GET")

	a.Router.HandleFunc("/students/{id}", func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]
		if lo.IsEmpty(id) {
			respondWithError(w, http.StatusBadRequest, "Invalid product ID")
			return
		}

		students := lo.Filter(allScores, func(x Score, index int) bool {
			return x.StudentId == id
		})

		respondWithJSON(w, http.StatusOK, students)

	}).Methods("GET")

	a.Router.HandleFunc("/exams", func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, allScores)
	}).Methods("GET")

	a.Router.HandleFunc("/exams/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		examId, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid ID")
			return
		}

		exams := lo.Filter(allScores, func(x Score, index int) bool {
			return x.Exam == examId
		})

		respondWithJSON(w, http.StatusOK, exams)
	}).Methods("GET")
}
