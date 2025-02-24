package playlist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/horiagug/youtube-md-go/internal/repository"
)

type Playlist struct {
	InputURL    string
	PlaylistID  string
	HTML        string
	YTCfg       map[string]interface{} // Placeholder for ytcfg
	InitialData map[string]interface{} // Placeholder for initial data
}

type playlistService struct {
	fetcher *repository.HTMLFetcher
}

func NewPlaylistService(fetcher *repository.HTMLFetcher) *playlistService {
	return &playlistService{
		fetcher: fetcher,
	}
}

// NewPlaylist creates a new Playlist instance.
func NewPlaylist(url string) (*Playlist, error) {
	playlistID, err := extractPlaylistID(url)
	if err != nil {
		return nil, err
	}

	return &Playlist{
		InputURL:    url,
		PlaylistID:  playlistID,
		HTML:        "",
		YTCfg:       make(map[string]interface{}),
		InitialData: make(map[string]interface{}),
	}, nil
}

func extractPlaylistID(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	queryValues := parsedURL.Query()
	listID := queryValues.Get("list")
	if listID == "" {
		// Attempt to extract playlist ID from the URL path
		re := regexp.MustCompile(`list=([a-zA-Z0-9_-]+)`)
		match := re.FindStringSubmatch(parsedURL.String())
		if len(match) > 1 {
			listID = match[1]
		} else {
			return "", fmt.Errorf("could not extract playlist ID from URL")
		}
	}

	return listID, nil
}

func (p *playlistService) fetchPlaylist(playlistId string) (string, error) {
	playlistURL := fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlistId)

	resp, err := p.fetcher.Fetch(playlistURL, nil)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

// Pure madness

type videoData struct {
	Contents []struct {
		PlaylistVideoRenderer struct {
			VideoId string `json:"videoId"`
		} `json:"playlistVideoRenderer"`
	} `json:"contents"`
	Continuations []struct {
		NextContinuationData struct {
			Continuation string `json:"continuation"`
		} `json:"nextContinuationData"`
	} `json:"continuations"`
}

type initialData struct {
	Contents struct {
		TwoColumnBrowseResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer struct {
							Contents []struct {
								ItemSectionRenderer struct {
									Contents []struct {
										PlaylistVideoListRenderer videoData
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// extractVideos extracts video URLs and a continuation token from raw JSON
func (p *playlistService) extractVideos(rawHTML string) ([]string, string, error) {
	data := initialData{}

	// Extract initial data
	matches := regexp.MustCompile(`var ytInitialData = ({.*?});`).FindStringSubmatch(rawHTML)
	if len(matches) < 2 {
		return nil, "", fmt.Errorf("could not find ytInitialData")
	}

	if err := json.Unmarshal([]byte(matches[1]), &data); err != nil {
		return nil, "", err
	}

	// Get playlist data
	playlist := data.Contents.TwoColumnBrowseResultsRenderer.Tabs[0].TabRenderer.
		Content.SectionListRenderer.Contents[0].ItemSectionRenderer.Contents[0].
		PlaylistVideoListRenderer

	// Extract video URLs
	videos := make([]string, len(playlist.Contents))
	for i, v := range playlist.Contents {
		videos[i] = fmt.Sprintf("https://www.youtube.com/watch?v=%s",
			v.PlaylistVideoRenderer.VideoId)
	}

	// Get continuation token
	continuation := ""
	if len(playlist.Continuations) > 0 {
		continuation = playlist.Continuations[0].NextContinuationData.Continuation
	}

	return videos, continuation, nil
}

// buildContinuationURL constructs the URL for fetching the next page of videos.
func (p *playlistService) buildContinuationURL(continuation string, inner_html string) (string, map[string]string, map[string]interface{}) {
	// Extract API key from the HTML
	re := regexp.MustCompile(`"INNERTUBE_API_KEY":"([^"]+)"`)
	matches := re.FindStringSubmatch(inner_html)
	apiKey := ""
	if len(matches) > 1 {
		apiKey = matches[1]
	}

	loadMoreURL := fmt.Sprintf("https://www.youtube.com/youtubei/v1/browse?key=%s", apiKey)
	headers := map[string]string{
		"X-YouTube-Client-Name":    "1",
		"X-YouTube-Client-Version": "2.20200720.00.02",
		"Content-Type":             "application/json",
	}
	data := map[string]interface{}{
		"continuation": continuation,
		"context": map[string]interface{}{
			"client": map[string]interface{}{
				"clientName":    "WEB",
				"clientVersion": "2.20200720.00.02",
			},
		},
	}

	return loadMoreURL, headers, data
}

// paginate fetches all video URLs from the playlist.
func (p *playlistService) paginate(playlistId string, html string) ([]string, error) {
	if html == "" {
		var err error
		html, err = p.fetchPlaylist(playlistId)
		if err != nil {
			return nil, err
		}
	}

	videoURLs, continuation, err := p.extractVideos(html)
	if err != nil {
		return nil, err
	}

	allVideoURLs := make([]string, 0)
	allVideoURLs = append(allVideoURLs, videoURLs...)

	for continuation != "" {
		loadMoreURL, headers, data := p.buildContinuationURL(continuation, html)

		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", loadMoreURL, strings.NewReader(string(jsonData)))
		if err != nil {
			return nil, err
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		newVideoURLs, newContinuation, err := p.extractVideos(string(body))
		if err != nil {
			return nil, err
		}

		allVideoURLs = append(allVideoURLs, newVideoURLs...)
		continuation = newContinuation
	}

	return allVideoURLs, nil
}

// VideoURLs returns a list of all video URLs in the playlist.
func (p *playlistService) VideoURLs(playlistId string) ([]string, error) {
	urls, err := p.paginate(playlistId, "")
	if err != nil {
		return nil, err
	}
	return urls, nil
}
