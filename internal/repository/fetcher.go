package repository

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type HTMLFetcher struct{}

func NewHTMLFetcher() *HTMLFetcher {
	return &HTMLFetcher{}
}

func (f *HTMLFetcher) Fetch(url string, cookie *http.Cookie) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept-Language", "en-US")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (f *HTMLFetcher) createConsentCookie(videoID string) (*http.Cookie, error) {
	html, err := f.Fetch(videoID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML to extract consent value: %w", err)
	}

	re := regexp.MustCompile(`name="v" value="(.*?)"`)
	match := re.FindSubmatch(html)
	if len(match) < 2 {
		return nil, fmt.Errorf("failed to find consent value in HTML")
	}
	consentValue := string(match[1])

	cookieValue := "YES+" + consentValue
	cookie := &http.Cookie{
		Name:   "CONSENT",
		Value:  cookieValue,
		Domain: ".youtube.com",
	}
	return cookie, nil
}

func consentRequired(body []byte) bool {
	consentRegex := regexp.MustCompile(`action="https://consent\.youtube\.com/s`)
	return consentRegex.Match(body)
}
