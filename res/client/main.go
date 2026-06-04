// Copyright (c) 2026, Anton Bespalov. Licensed under the MIT License.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080/render", "URL of the rendering server")
	outputFile := flag.String("output", "dashboard.png", "Output image file name")
	interval := flag.Duration("interval", 0, "Polling interval (e.g. 5s). If 0, runs once.")
	flag.Parse()

	if *interval > 0 {
		log.Printf("Starting polling every %v...", *interval)
		ticker := time.NewTicker(*interval)
		for {
			fetchImage(*serverURL, *outputFile)
			<-ticker.C
		}
	} else {
		fetchImage(*serverURL, *outputFile)
	}
}

func fetchImage(url, outPath string) {
	log.Printf("Fetching from %s...", url)

	// 1. Make HTTP GET request
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error requesting image: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Server returned status: %s", resp.Status)
		return
	}

	// 2. Create output file
	out, err := os.Create(outPath)
	if err != nil {
		log.Printf("Error creating output file: %v", err)
		return
	}
	defer out.Close()

	// 3. Write data to file
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("Error writing image data: %v", err)
		return
	}

	fmt.Printf("Successfully downloaded %d bytes to %s\n", written, outPath)
}
