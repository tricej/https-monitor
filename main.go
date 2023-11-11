package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type responseData struct {
	address       string
	responseCode  int
	responseTime  time.Duration
	responseError error
}

func checkResponse(address string) responseData {
	dnsName := strings.Split(address, "www.")[1]
	if _, err := net.LookupIP(dnsName); err != nil {
		slog.Error("unable to resolve dns name", "URL", address)
		return responseData{
			address:       address,
			responseCode:  0,
			responseTime:  0,
			responseError: err,
		}
	}

	startTime := time.Now()

	// Perform an HTTP GET request to the specified URL
	response, err := http.Get(address)

	// Record the end time after the HTTP request is complete
	endTime := time.Now()

	// Calculate the response time
	responseTime := endTime.Sub(startTime)

	if err != nil {
		// If there was an error connecting to the URL, log an error message
		slog.Error("Unable to connect to endpoint", "URL", address, "Error", err)
		return responseData{
			address:       address,
			responseCode:  0,
			responseTime:  0,
			responseError: err,
		}
	}

	if response == nil {
		slog.Error("empty http response", "url", address)
		return responseData{
			address:       address,
			responseCode:  0,
			responseTime:  0,
			responseError: err,
		}
	}
	defer response.Body.Close()

	return responseData{
		address:       address,
		responseCode:  response.StatusCode,
		responseTime:  time.Duration(responseTime),
		responseError: err,
	}
}

func testLoop(testList []string, loopWaitTime time.Duration) {
	for {
		testResults := []responseData{}

		// Test each address and store the results
		for _, address := range testList {
			testResult := checkResponse(address)
			// Append the result regardless of whether there is an error or not
			testResults = append(testResults, testResult)
		}

		// Print the results for each tested address
		for _, testResult := range testResults {
			fmt.Printf("Address: %v\n", testResult.address)
			fmt.Printf("Response Code: %v\n", testResult.responseCode)
			fmt.Printf("Response Time: %vms\n", testResult.responseTime.Milliseconds())
			if testResult.responseError != nil {
				fmt.Printf("Reponse Error: %v\n", testResult.responseError.Error())
			}
		}

		fmt.Printf("Sleeping %v\n", loopWaitTime.String())
		time.Sleep(loopWaitTime)
	}
}
func main() {
	// List of addresses to test
	addressList := [5]string{
		"https://www.google.com",
		"https://www.cloudflare.com",
		"https://www.amazonfdad.com",
		"https://www.github.com",
		"https://www.stackoverflow.com",
	}

	// Wait time between test iterations
	var waitTime time.Duration = 30 * time.Second

	// Start the test loop
	testLoop(addressList[:], waitTime)
}
