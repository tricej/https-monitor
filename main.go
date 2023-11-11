package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	responseCodeCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_code",
			Help: "HTTP response codes",
		},
		[]string{"address", "code"},
	)
	responseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time",
			Help:    "HTTP response time in milliseconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"address"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(responseCodeCounter)
	prometheus.MustRegister(responseTimeHistogram)
}

type responseData struct {
	address      string
	responseCode int
	responseTime time.Duration
}

func checkResponse(address string) responseData {
	startTime := time.Now()

	// Perform an HTTP GET request to the specified URL
	response, err := http.Get(address)

	// Record the end time after the HTTP request is complete
	endTime := time.Now()

	if err != nil {
		// If there was an error connecting to the URL, log an error message
		slog.Error("Unable to connect to endpoint", "URL", address, "Error", err)
		return responseData{
			address:      address,
			responseCode: 0,
			responseTime: 0,
		}
	}

	// Check if the HTTP response is not valid (e.g., status code other than 2xx)
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		slog.Error("Unexpected HTTP response status", "URL", address, "StatusCode", response.StatusCode)
	} else if response == nil {
		slog.Error("Error: Empty HTTP response", "URL", address)
		return responseData{
			address:      address,
			responseCode: 0,
			responseTime: 0,
		}
	}

	// Calculate the response time
	responseTime := endTime.Sub(startTime)

	defer response.Body.Close()

	// Increment Prometheus counters
	responseCodeCounter.WithLabelValues(address, fmt.Sprint(response.StatusCode)).Inc()
	responseTimeHistogram.WithLabelValues(address).Observe(float64(responseTime.Milliseconds()))

	return responseData{
		address:      address,
		responseCode: response.StatusCode,
		responseTime: time.Duration(responseTime),
	}
}

func testLoop(testList []string, loopWaitTime time.Duration) {
	for {
		testResults := []responseData{}
		fmt.Printf("Sleeping %v\n", loopWaitTime.String())
		time.Sleep(loopWaitTime)

		// Test each address and store the results
		for _, address := range testList {
			testResult := checkResponse(address)
			testResults = append(testResults, testResult)
		}

		// Print the results for each tested address
		for _, testResult := range testResults {
			fmt.Printf("Address: %v\n", testResult.address)
			fmt.Printf("Response Code: %v\n", testResult.responseCode)
			fmt.Printf("Response Time: %vms\n", testResult.responseTime.Milliseconds())
		}

		// Push metrics to Prometheus Pushgateway
		err := push.New("http://your-pushgateway-url", "https-monitor-v2").
			Collector(responseCodeCounter).
			Collector(responseTimeHistogram).
			Grouping("instance", "example").
			Push()
		if err != nil {
			slog.Error("Error pushing metrics to Pushgateway", "Error", err)
		}
	}
}

func main() {
	// List of addresses to test
	addressList := [5]string{
		"https://www.google.com",
		"https://go.dev",
		"https://www.amazon.com",
		"https://www.github123.com",
		"https://www.stackoverflow.com",
	}

	// Wait time between test iterations
	var waitTime time.Duration = 30 * time.Second

	// Start the test loop
	testLoop(addressList[:], waitTime)
}
