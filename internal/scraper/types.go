package scraper

type Episode struct {
	ID            int    `json:"id"`
	EpisodeNumber int    `json:"episodeNumber"`
	SeasonNumber  int    `json:"seasonNumber"`
	Name          string `json:"name"`
	Overview      string `json:"overview,omitempty"`
	StillPath     string `json:"stillPath,omitempty"`
	AirDate       string `json:"airDate,omitempty"`
	Runtime       int    `json:"runtime,omitempty"`
	VoteAverage   float64 `json:"voteAverage,omitempty"`
}

type MediaDetails struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Poster      string    `json:"poster,omitempty"`
	Cover       string    `json:"cover,omitempty"`
	Type        string    `json:"type"` // "movie" or "tv"
	ReleaseYear int       `json:"releaseYear,omitempty"`
	Status      string    `json:"status,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
	Rating      float64   `json:"rating,omitempty"`
	Episodes    []Episode `json:"episodes,omitempty"`
	Sources     []StreamSource `json:"sources,omitempty"`
}

type Subtitle struct {
	Lang   string `json:"lang"`
	URL    string `json:"url"`
	Format string `json:"format"` // usually "vtt"
}

type StreamSource struct {
	Quality    string            `json:"quality"`
	URL        string            `json:"url"`
	IsM3U8     bool              `json:"isM3U8"`
	IsMP4      bool              `json:"isMP4"`
	IsEmbed    bool              `json:"isEmbed"`
	ServerName string            `json:"serverName"`
	AudioType  string            `json:"audioType,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Subtitles  []Subtitle        `json:"subtitles,omitempty"`
}

type SearchResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Poster      string `json:"poster,omitempty"`
	ReleaseYear int    `json:"releaseYear,omitempty"`
}
