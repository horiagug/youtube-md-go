package youtube_md

// Playlist represents a YouTube playlist.
type Playlist struct {
	InputURL    string
	PlaylistID  string
	HTML        string
	YTCfg       map[string]interface{} // Placeholder for ytcfg
	InitialData map[string]interface{} // Placeholder for initial data
}
