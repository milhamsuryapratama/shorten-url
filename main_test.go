package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	vegeta "github.com/tsenart/vegeta/lib"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"testing"
	"time"
)

func TestShortenURL(t *testing.T) {
	client := &http.Client{}

	shortURLRequest := ShortURLRequest{
		TargetURL: "https://www.youtube.com",
	}

	shortURLRequestByte, _ := json.Marshal(shortURLRequest)

	request, err := http.NewRequest("POST", "http://localhost:8000/short", bytes.NewBuffer(shortURLRequestByte))
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	ko, _ := httputil.DumpResponse(response, true)
	fmt.Println(string(ko))

	defer response.Body.Close()

	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestShortenURLWithGoroutine(t *testing.T) {
	wg := &sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			client := &http.Client{}

			shortURLRequest := ShortURLRequest{
				TargetURL: fmt.Sprintf("www.google.com/%d", i),
			}

			shortURLRequestByte, _ := json.Marshal(shortURLRequest)

			request, err := http.NewRequest("POST", "http://localhost:8000/short", bytes.NewBuffer(shortURLRequestByte))
			if err != nil {
				wg.Done()
				t.Error(err)
				return
			}

			response, err := client.Do(request)
			if err != nil {
				wg.Done()
				t.Error(err)
				return
			}

			defer response.Body.Close()

			assert.Equal(t, http.StatusOK, response.StatusCode)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestShortenURLWithVegeta(t *testing.T) {
	rate := vegeta.Rate{Freq: 1000, Per: time.Second}
	duration := 1 * time.Second
	body := `{"targetURL":"http://www.google.com"}`
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "POST",
		URL:    "http://localhost:8000/short",
		Body:   []byte(body),
	})
	attacker := vegeta.NewAttacker()
	vegeta.Workers(100)

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "") {
		metrics.Add(res)
	}
	metrics.Close()

	reporter := vegeta.NewTextReporter(&metrics)
	reporter.Report(os.Stdout)

	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}
