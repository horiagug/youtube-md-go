package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/horiagug/youtube-md-go/config"
	"github.com/horiagug/youtube-md-go/pkg/youtube_md"
)

func getGeminiAPIKey() (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}
	return apiKey, nil
}

func main() {
	var (
		languages      = flag.String("languages", "en", "Comma-separated list of language codes")
		geminiAPIKey   = flag.String("gemini-api-key", "", "Gemini API Key")
		geminiAPIModel = flag.String("gemini-api-model", "gemini-2.0-flash", "Gemini API Key")
		timeout        = flag.Duration("timeout", 60*time.Second, "Operation timeout")
	)
	flag.Parse()

	if *geminiAPIKey == "" {
		var err error
		*geminiAPIKey, err = getGeminiAPIKey()
		if err != nil || *geminiAPIKey == "" {
			fmt.Println("Failed to fetch gemini api key, please provide a Gemini API Key")
			os.Exit(1)
		}
	}

	cfg := config.New(
		config.WithGeminiAPIKey(*geminiAPIKey),
		config.WithLanguages(strings.Split(*languages, ",")),
		config.WithGeminiModel(*geminiAPIModel),
	)

	client, err := youtube_md.New(cfg,
		youtube_md.WithTimeout(*timeout),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	statusChan := make(chan string)
	doneChan := make(chan bool)

	// Progress indicator routine
	go func() {
		for {
			select {
			case status := <-statusChan:
				fmt.Printf("\r\033[K%s...", status)
			case <-doneChan:
				fmt.Println("\r\033[KDone!")
				return
			}
		}
	}()

	err = client.GenerateMarkdown(flag.Arg(0), statusChan)
	if err != nil {
		log.Fatalf("Failed to generate markdown: %v", err)
		doneChan <- true
	}
	doneChan <- true
	os.Exit(0)
}
