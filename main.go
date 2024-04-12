package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"sync"
)

var mapShortenKey = sync.Map{}

type ShortURLRequest struct {
	TargetURL string
}

type ShortURLResponse struct {
	ShortURL string `json:"shortURL"`
}

func generateShortKey(targetURL string) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	//rand.Seed(time.Now().UnixNano())
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}

	strKey := string(shortKey)

	if _, ok := mapShortenKey.Load(strKey); !ok {
		mapShortenKey.Store(strKey, targetURL)
	}

	return strKey
}

func Short(writer http.ResponseWriter, request *http.Request) {
	var shortenURLRequest ShortURLRequest
	var shortURLResponse ShortURLResponse

	err := json.NewDecoder(request.Body).Decode(&shortenURLRequest)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(writer).Encode(shortURLResponse)
	}

	fmt.Println(shortenURLRequest.TargetURL)

	shortKey := generateShortKey(shortenURLRequest.TargetURL)
	if shortKey == "" {
		writer.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(writer).Encode(shortURLResponse)
	}

	shortURLResponse.ShortURL = fmt.Sprintf("www.gatherloo.com/%s", shortKey)

	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(&shortURLResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(writer).Encode(nil)
	}
}

func Redirect(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	redirectURL, ok := mapShortenKey.Load(vars["key"])
	if !ok {
		redirectURL = "/"
	}

	http.Redirect(writer, request, redirectURL.(string), http.StatusMovedPermanently)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/short", Short).Methods(http.MethodPost)
	r.HandleFunc("/short/{key}", Redirect).Methods(http.MethodGet)

	err := http.ListenAndServe(":8000", r)
	if err != nil {
		panic(err)
	}
}
