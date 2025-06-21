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
	GenerateMarkdownFile(ctx context.Context, videoID string, language string, statusChan *chan string) error
	GenerateMarkdown(ctx context.Context, videoID string, language string) (string, error)
	GenerateMarkdownData(ctx context.Context, videoID string, language string) (*MarkdownData, error)
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

func (m service) GenerateMarkdown(ctx context.Context, videoUrl string, language string) (string, error) {
	videoURLs, err := m.getVideosFromUrl(videoUrl)

	if err != nil {
		return "", fmt.Errorf("Error getting video from url")
	}

	transcripts, err := m.fetchTranscriptsForVideos(videoURLs, language)
	if err != nil {
		return "", fmt.Errorf("Error fetching transcripts: %v", err)
	}

	markdown_text, err := m.generateMarkdownFromTranscripts(ctx, transcripts)
	if err != nil {
		return "", fmt.Errorf("Error processing transcripts: %v", err)
	}
	return markdown_text, nil
}

func (m service) GenerateMarkdownFile(ctx context.Context, videoUrl string, language string, statusChan *chan string) error {
	*statusChan <- fmt.Sprintf("Processing video: %v", videoUrl)
	videoURLs, err := m.getVideosFromUrl(videoUrl)

	if err != nil {
		return fmt.Errorf("Error getting video from url")
	}

	*statusChan <- fmt.Sprintf("Fetching transcripts for %v videos", len(videoURLs))

	transcripts, err := m.fetchTranscriptsForVideos(videoURLs, language)
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

type MarkdownData struct {
	Markdown   string
	VideoUrl   string
	VideoId    string
	VideoTitle string
}

func (m service) GenerateMarkdownData(ctx context.Context, videoUrl string, language string) (*MarkdownData, error) {
	videoURLs, err := m.getVideosFromUrl(videoUrl)

	if err != nil {
		return nil, fmt.Errorf("Error getting video from url")
	}

	transcripts, err := m.fetchTranscriptsForVideos(videoURLs, language)
	if err != nil {
		return nil, fmt.Errorf("Error fetching transcripts: %v", err)
	}

	markdown_text, err := m.generateMarkdownFromTranscripts(ctx, transcripts)
	if err != nil {
		return nil, fmt.Errorf("Error processing transcripts: %v", err)
	}
	return &MarkdownData{Markdown: markdown_text, VideoUrl: fmt.Sprintf("https://www.youtube.com/watch?v=%s", transcripts[0].VideoID), VideoId: transcripts[0].VideoID, VideoTitle: transcripts[0].VideoTitle}, nil
}
