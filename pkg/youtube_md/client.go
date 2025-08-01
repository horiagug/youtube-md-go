package youtube_md

import (
	"context"
	"time"

	"github.com/horiagug/youtube-md-go/internal/markdown"
	"github.com/horiagug/youtube-md-go/pkg/config"
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
		_ = cancel
	}
}

func NewClient(cfg *config.Config, opts ...ClientOption) (*Client, error) {
	client := &Client{
		config: cfg,
		ctx:    context.Background(),
	}

	for _, opt := range opts {
		opt(client)
	}

	markdownService, err := markdown.NewService(client.ctx, cfg)
	if err != nil {
		return nil, err
	}

	client.markdownService = markdownService
	return client, nil
}

func (c *Client) GenerateMarkdownFile(videoId string, language string, statusChan *chan string) error {
	return c.markdownService.GenerateMarkdownFile(c.ctx, videoId, language, statusChan)
}

func (c *Client) GenerateMarkdown(videoId string, language string) (string, error) {
	res, err := c.markdownService.GenerateMarkdown(c.ctx, videoId, language)
	return res, err
}

func (c *Client) GenerateMarkdownData(videoId string, language string) (*markdown.MarkdownData, error) {
	res, err := c.markdownService.GenerateMarkdownData(c.ctx, videoId, language)
	return res, err
}
