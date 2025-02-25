package youtube_md

import (
	"context"
	"time"

	"github.com/horiagug/youtube-md-go/config"
	"github.com/horiagug/youtube-md-go/internal/markdown"
)

type Client struct {
	markdownService markdown.Service
	config          *config.Config
	ctx             context.Context
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		c.ctx = ctx
		go func() {
			<-ctx.Done()
			cancel()
		}()
	}
}

func New(cfg *config.Config, opts ...ClientOption) (*Client, error) {
	client := &Client{
		config: cfg,
		ctx:    context.Background(),
	}

	for _, opt := range opts {
		opt(client)
	}

	// Initialize services
	markdownService, err := markdown.NewService(client.ctx, cfg)
	if err != nil {
		return nil, err
	}

	client.markdownService = markdownService
	return client, nil
}

func (c *Client) GenerateMarkdown(videoID string, statusChan chan string) error {
	return c.markdownService.Generate(c.ctx, videoID, &statusChan)
}
