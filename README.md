# YouTube Markdown Generator

A Go-based tool that converts YouTube video transcripts into well-structured markdown content using Google's Gemini AI. Supports both single videos and playlists.

## Features

- 🎥 Process single YouTube videos or entire playlists
- 🌐 Multi-language support for transcripts
- 🤖 AI-powered content structuring using Google's Gemini
- ✨ Generates well-formatted markdown with:
  - Logical sections and subheadings
  - Bullet points and numbered lists
  - Bold text for emphasis
  - Clear topic separators
- 📝 Preserves all original content details and context
- ⚡ Concurrent processing for playlist videos

## Prerequisites

- Go 1.23 or higher
- Google Gemini API key
- Internet connection for accessing YouTube and Gemini API

## Installation

```bash
go install github.com/horiagug/youtube-md-go
```

## Configuration

Set your Gemini API key using one of these methods:

1. Environment variable:

```bash
export GEMINI_API_KEY=your_api_key_here
```

2. Command-line flag:

```bash
--gemini-api-key=your_api_key_here
```

## Usage

### Basic Usage

```bash
youtube-md-go [flags] <youtube-url>
```

### Command Line Flags

- `--languages`: Comma-separated list of language codes (default: "en")
- `--gemini-api-key`: Gemini API Key (optional if set via environment)
- `--gemini-api-model`: Gemini model to use (default: "gemini-2.0-flash")
- `--timeout`: Operation timeout duration (default: 60s)

### Examples

Process a single video:

```bash
youtube-md-go "https://www.youtube.com/watch?v=video_id"
```

Process a playlist:

```bash
youtube-md-go "https://www.youtube.com/playlist?list=playlist_id"
```

With custom settings:

```bash
youtube-md-go --languages=en,es "https://www.youtube.com/watch?v=video_id"
```

## Output

The tool generates a markdown file (`output.md`) containing the structured content. The output includes:

- Organized sections with headings
- Formatted lists and bullet points
- Emphasized key terms
- Clear topic transitions

## Supported Gemini Models

- gemini-1.5-flash
- gemini-1.5-pro
- gemini-2.0-flash
- gemini-2.0-flash-thinking-exp-01-21
- gemini-2.5-flash
- gemini-2.5-pro

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
