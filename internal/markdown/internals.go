package markdown

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/horiagug/youtube-md-go/internal/playlist"
	"github.com/horiagug/youtube-md-go/internal/repository"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"google.golang.org/genai"
)

func (m service) getVideosFromUrl(videoUrl string) ([]string, error) {

	videoURLs := []string{}
	if strings.Contains(videoUrl, "playlist?list=") {

		p, err := playlist.NewPlaylist(videoUrl)
		if err != nil {
			return videoURLs, fmt.Errorf("Error creating playlist: %v", err)
		}

		playlist_service := playlist.NewPlaylistService(repository.NewHTMLFetcher())

		videos, err := playlist_service.VideoURLs(p.PlaylistID)
		if err != nil {
			return videoURLs, fmt.Errorf("Error getting video URLs: %v", err)
		}
		videoURLs = videos
	} else {
		videoURLs = append(videoURLs, videoUrl)
	}
	return videoURLs, nil
}

func (m service) fetchTranscriptsForVideos(videoUrls []string, language string) ([]yt_transcript_models.Transcript, error) {

	results := make(chan transcriptServiceResults, len(videoUrls))
	var wg sync.WaitGroup

	// Launch goroutine for each video URL
	for _, url := range videoUrls {
		wg.Add(1)
		go func(videoURL string) {
			defer wg.Done()
			transcript, err := m.transcriptService.GetTranscripts(videoURL, []string{language})
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
			return []yt_transcript_models.Transcript{}, fmt.Errorf("Error getting transcript for %v", result.err)
		}
		transcripts = append(transcripts, result.transcripts...)
	}
	return transcripts, nil
}

func (m *service) generateMarkdownFileFromTranscripts(ctx context.Context, transcripts []yt_transcript_models.Transcript, statusChan *chan string) error {
	var fullText strings.Builder
	for _, transcript := range transcripts {
		for _, line := range transcript.Lines {
			fullText.WriteString(line.Text)
			fullText.WriteString(" ")
		}
	}

	text := fullText.String()
	words := strings.Fields(text)

	var chunks []string
	chunkSize := 6000

	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}

	file, err := os.OpenFile("output.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Error opening file: %v", err)
	}
	defer file.Close()

	language := display.English.Tags().Name(language.MustParse(transcripts[0].LanguageCode))
	previous_response := ""
	*statusChan <- fmt.Sprintf("Progress: %v%%", 0)

	for i, chunk := range chunks {
		context_prompt := ""
		if previous_response != "" {
			context_prompt = fmt.Sprintf("The following text is a continuation... \nPrevious response: \n %s \n \n New text to process(Do Not Repeat the Previous response:): \n", previous_response)
		}
		formatted_prompt := strings.ReplaceAll(prompt, "[Language]", language)
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

		percent_done := int((float64(i+1) / float64(len(chunks))) * 100)
		*statusChan <- fmt.Sprintf("Progress: %v%%", percent_done)
	}
	return nil
}

func (m *service) generateMarkdownFromTranscripts(ctx context.Context, transcripts []yt_transcript_models.Transcript) (string, error) {
	var fullText strings.Builder
	for _, transcript := range transcripts {
		for _, line := range transcript.Lines {
			fullText.WriteString(line.Text)
			fullText.WriteString(" ")
		}
	}

	text := fullText.String()
	words := strings.Fields(text)

	var chunks []string
	chunkSize := 6000

	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}

	language := display.English.Tags().Name(language.MustParse(transcripts[0].LanguageCode))
	var responseBuilder strings.Builder
	previous_response := ""

	for _, chunk := range chunks {
		context_prompt := ""
		if previous_response != "" {
			context_prompt = fmt.Sprintf("The following text is a continuation... \nPrevious response: \n %s \n \n New text to process(Do Not Repeat the Previous response:): \n", previous_response)
		}
		formatted_prompt := strings.ReplaceAll(prompt, "[Language]", language)
		full_prompt := fmt.Sprintf("%s%s \n\n %s", context_prompt, formatted_prompt, chunk)

		parts := []*genai.Part{
			{Text: full_prompt},
		}
		content := []*genai.Content{{Parts: parts}}

		response, err := m.genaiClient.Models.GenerateContent(ctx, m.config.GeminiModel, content, nil)
		if err != nil {
			return "", fmt.Errorf("Error generating content: %v", err)
		}

		responseText := response.Candidates[0].Content.Parts[0].Text
		responseBuilder.WriteString(responseText)
		responseBuilder.WriteString("\n")
		previous_response = responseText
	}

	return responseBuilder.String(), nil
}
