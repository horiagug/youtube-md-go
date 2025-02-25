package markdown

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/horiagug/youtube-md-go/config"
	"github.com/horiagug/youtube-md-go/internal/playlist"
	"github.com/horiagug/youtube-md-go/internal/repository"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
	"google.golang.org/genai"
)

type Service interface {
	Generate(ctx context.Context, videoID string, statusChan *chan string) error
}

type service struct {
	transcriptService *yt_transcript.YtTranscriptClient
	genaiClient       *genai.Client
	config            *config.Config
}

type transcriptServiceResults struct {
	transcripts []yt_transcript_models.Transcript
	err         error
}

func NewService(ctx context.Context, cfg *config.Config) (service, error) {
	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return service{}, err
	}

	transcriptService := yt_transcript.NewClient()

	return service{
		transcriptService: transcriptService,
		genaiClient:       genaiClient,
		config:            cfg,
	}, nil
}

func (m service) Generate(ctx context.Context, videoURL string, statusChan *chan string) error {
	videoURLs := []string{}

	*statusChan <- fmt.Sprintf("Processing video: %v", videoURL)

	if strings.Contains(videoURL, "playlist?list=") {

		*statusChan <- fmt.Sprintf("Processing playlist: %v", videoURL)
		p, err := playlist.NewPlaylist(videoURL)
		if err != nil {
			return fmt.Errorf("Error creating playlist: %v", err)
		}

		playlist_service := playlist.NewPlaylistService(repository.NewHTMLFetcher())

		videos, err := playlist_service.VideoURLs(p.PlaylistID)
		if err != nil {
			return fmt.Errorf("Error getting video URLs: %v", err)
		}
		videoURLs = videos
	} else {
		videoURLs = append(videoURLs, videoURL)
	}

	*statusChan <- fmt.Sprintf("Fetching transcripts for %v videos", len(videoURLs))
	// Create channel for results
	results := make(chan transcriptServiceResults, len(videoURLs))
	var wg sync.WaitGroup

	// Launch goroutine for each video URL
	for _, url := range videoURLs {
		wg.Add(1)
		go func(videoURL string) {
			defer wg.Done()
			transcript, err := m.transcriptService.GetTranscripts(videoURL, []string{"en"})
			results <- transcriptServiceResults{
				transcripts: transcript,
				err:         err,
			}
		}(url)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var transcripts []yt_transcript_models.Transcript
	for result := range results {
		if result.err != nil {
			return fmt.Errorf("Error getting transcript for %v", result.err)
		}
		transcripts = append(transcripts, result.transcripts...)
	}

	*statusChan <- fmt.Sprintf("Generating markdown for %v transcript(s)", len(transcripts))

	err := m.processTranscripts(ctx, transcripts, statusChan)
	if err != nil {
		return fmt.Errorf("Error processing transcripts: %v", err)
	}

	return nil
}

func (m *service) processTranscripts(ctx context.Context, transcripts []yt_transcript_models.Transcript, statusChan *chan string) error {
	var chunks []string
	currentChunk := make([]string, 0, 3000)
	wordCount := 0

	for _, transcript := range transcripts {
		for _, line := range transcript.Lines {
			words := strings.Fields(line.Text)

			for _, word := range words {
				currentChunk = append(currentChunk, word)
				wordCount++

				if wordCount >= 3000 {
					// Join current chunk and add to chunks slice
					chunks = append(chunks, strings.Join(currentChunk, " "))
					// Reset for next chunk
					currentChunk = make([]string, 0, 3000)
					wordCount = 0
				}
			}
		}
	}

	if wordCount > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	file, err := os.OpenFile("output.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Error opening file: %v", err)
	}

	percent_done := 0
	defer file.Close()
	previous_response := ""
	*statusChan <- fmt.Sprintf("Progress: %v%%", percent_done)

	for i, chunk := range chunks {
		context_prompt := ""
		if previous_response != "" {
			context_prompt = fmt.Sprintf("The following text is a continuation... \nPrevious response: \n %s \n \n New text to process(Do Not Repeat the Previous response:): \n", previous_response)
		}
		formatted_prompt := strings.ReplaceAll(prompt, "[Language]", "English")

		full_prompt := fmt.Sprintf("%s%s \n\n %s", context_prompt, formatted_prompt, chunk)

		parts := []*genai.Part{
			{Text: full_prompt},
		}
		content := []*genai.Content{{Parts: parts}}

		response, err := m.genaiClient.Models.GenerateContent(ctx, m.config.GeminiModel, content, nil)
		if err != nil {
			return fmt.Errorf("Error generating content: %v", err)
		}

		responseText := response.Candidates[0].Content.Parts[0].Text

		if _, err := file.WriteString(responseText + "\n"); err != nil {
			return fmt.Errorf("Error writing to file: %v", err)
		}
		previous_response = responseText

		percent_done = int((float64(i+1) / float64(len(chunks))) * 100)
		*statusChan <- fmt.Sprintf("Progress: %v%%", percent_done)
	}
	return nil
}
