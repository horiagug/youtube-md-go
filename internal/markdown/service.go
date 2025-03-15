package markdown

import (
	"context"
	"fmt"
	"github.com/horiagug/youtube-md-go/pkg/config"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
	"google.golang.org/genai"
)

type Service interface {
	GenerateMarkdownFile(ctx context.Context, videoID string, statusChan *chan string) error
	GenerateMarkdown(ctx context.Context, videoID string) (string, error)
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

func (m service) GenerateMarkdown(ctx context.Context, videoUrl string) (string, error) {
	videoURLs, err := m.getVideosFromUrl(videoUrl)

	if err != nil {
		return "", fmt.Errorf("Error getting video from url")
	}

	transcripts, err := m.fetchTranscriptsForVideos(videoURLs)
	if err != nil {
		return "", fmt.Errorf("Error fetching transcripts: %v", err)
	}

	markdown_text, err := m.generateMarkdownFromTranscripts(ctx, transcripts)
	if err != nil {
		return "", fmt.Errorf("Error processing transcripts: %v", err)
	}
	return markdown_text, nil
}

func (m service) GenerateMarkdownFile(ctx context.Context, videoUrl string, statusChan *chan string) error {
	*statusChan <- fmt.Sprintf("Processing video: %v", videoUrl)
	videoURLs, err := m.getVideosFromUrl(videoUrl)

	if err != nil {
		return fmt.Errorf("Error getting video from url")
	}

	*statusChan <- fmt.Sprintf("Fetching transcripts for %v videos", len(videoURLs))

	transcripts, err := m.fetchTranscriptsForVideos(videoURLs)
	if err != nil {
		return fmt.Errorf("Error fetching transcripts: %v", err)
	}

	*statusChan <- fmt.Sprintf("Generating markdown for %v transcript(s)", len(transcripts))

	err = m.generateMarkdownFileFromTranscripts(ctx, transcripts, statusChan)
	if err != nil {
		return fmt.Errorf("Error processing transcripts: %v", err)
	}
	return nil
}
