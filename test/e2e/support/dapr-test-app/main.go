package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/gorilla/mux"
)

var (
	stateStoreName string
	appPort        string
	daprClient     dapr.Client
	once           sync.Once
)

const (
	HTTPReadHeaderTimeout = 3 * time.Second
)

func init() {
	appPort = os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	stateStoreName = os.Getenv("STATESTORE_NAME")
	if stateStoreName == "" {
		stateStoreName = "statestore"
	}
}

type MyValues struct {
	Values []string `json:"values"`
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("message")
	values, _ := read(r.Context(), stateStoreName, "values")

	values.Values = append(values.Values, value)

	data, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}

	err = save(r.Context(), stateStoreName, "values", data)
	if err != nil {
		panic(err)
	}

	err = respondWithJSON(w, http.StatusOK, values)
	if err != nil {
		panic(err)
	}
}

func client() dapr.Client {
	once.Do(func() {
		dc, err := dapr.NewClient()
		if err != nil {
			panic(err)
		}

		daprClient = dc
	})

	return daprClient
}
func readHandler(w http.ResponseWriter, r *http.Request) {
	values, _ := read(r.Context(), stateStoreName, "values")

	if err := respondWithJSON(w, http.StatusOK, values); err != nil {
		panic(err)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err = w.Write(response)

	return err
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/write", writeHandler).Methods("POST")
	r.HandleFunc("/read", readHandler).Methods("GET")
	r.HandleFunc("/health/readiness", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	r.HandleFunc("/health/liveness", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	http.Handle("/", r)

	server := &http.Server{
		Addr:              ":" + appPort,
		ReadHeaderTimeout: HTTPReadHeaderTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func save(ctx context.Context, storeName string, key string, data []byte) error {
	return client().SaveState(ctx, storeName, key, data, nil)
}

func read(ctx context.Context, storeName string, key string) (MyValues, error) {
	result, err := client().GetState(ctx, storeName, key, nil)
	if err != nil {
		return MyValues{}, err
	}

	myValues := MyValues{}
	if result.Value != nil {
		err = json.Unmarshal(result.Value, &myValues)
		if err != nil {
			return MyValues{}, err
		}
	}

	if myValues.Values == nil {
		myValues.Values = make([]string, 0)
	}

	return myValues, nil
}
