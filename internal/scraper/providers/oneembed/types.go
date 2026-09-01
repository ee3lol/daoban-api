package oneembed

type TokenResponse struct {
	Token       string `json:"token"`
	LegacyToken string `json:"legacyToken"`
	S           string `json:"s"`
	ExpiresIn   int    `json:"expires_in"`
}

type AudioTrack struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Language string `json:"language"`
}

type StreamsData struct {
	RawM3U8       string `json:"raw_m3u8"`
	ProxyM3U8     string `json:"proxy_m3u8"`
	VPSProxyM3U8  string `json:"vps_proxy_m3u8"`
	LocalProxyM3U8 string `json:"local_proxy_m3u8"`
	WorkerProxyM3U8 string `json:"worker_proxy_m3u8"`
	M3U8          string `json:"m3u8"`
	Format        string `json:"format"`
}

type StreamTitle struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PosterURL   string `json:"poster_url"`
	BackdropURL string `json:"backdrop_url"`
	TmdbID      string `json:"tmdb_id"`
	ImdbID      string `json:"imdb_id"`
	Overview    string `json:"overview"`
	ReleaseDate string `json:"release_date"`
	Type        string `json:"type"`
}

type OneEmbedSubtitle struct {
	Lang  string `json:"lang"`
	Label string `json:"label"`
	File  string `json:"file"`
	URL   string `json:"url"`
}

type StreamResponse struct {
	Success        bool               `json:"success"`
	Provider       string             `json:"provider"`
	SelectedSource string             `json:"selectedSource"`
	IsIframe       bool               `json:"isIframe"`
	StreamURL      string             `json:"streamUrl"`
	SourceTitle    string             `json:"sourceTitle"`
	Error          string             `json:"error"`
	Title          StreamTitle        `json:"title"`
	AudioTracks    []AudioTrack       `json:"audioTracks"`
	Streams        StreamsData        `json:"streams"`
	Subtitles      []OneEmbedSubtitle `json:"subtitles"`
}

type TmdbGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TmdbDetailsResponse struct {
	Success         bool        `json:"success"`
	ID              int         `json:"id"`
	Title           string      `json:"title"`
	Name            string      `json:"name"`
	Overview        string      `json:"overview"`
	PosterPath      string      `json:"poster_path"`
	BackdropPath    string      `json:"backdrop_path"`
	ReleaseDate     string      `json:"release_date"`
	FirstAirDate    string      `json:"first_air_date"`
	VoteAverage     float64     `json:"vote_average"`
	Genres          []TmdbGenre `json:"genres"`
	ImdbID          string      `json:"imdb_id"`
	NumberOfSeasons int         `json:"number_of_seasons"`
	Status          string      `json:"status"`
}

type TvEpisode struct {
	ID            int     `json:"id"`
	EpisodeNumber int     `json:"episode_number"`
	SeasonNumber  int     `json:"season_number"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	StillPath     string  `json:"still_path"`
	AirDate       string  `json:"air_date"`
	Runtime       int     `json:"runtime"`
	VoteAverage   float64 `json:"vote_average"`
}

type TvSeason struct {
	SeasonNumber int    `json:"season_number"`
	Name         string `json:"name"`
	EpisodeCount int    `json:"episode_count"`
}

type TvEpisodesResponse struct {
	Success      bool        `json:"success"`
	TmdbID       string      `json:"tmdb_id"`
	Season       int         `json:"season"`
	TotalSeasons int         `json:"total_seasons"`
	ShowName     string      `json:"show_name"`
	Seasons      []TvSeason  `json:"seasons"`
	Episodes     []TvEpisode `json:"episodes"`
}

type SubtitlesResponse struct {
	Success   bool               `json:"success"`
	Subtitles []OneEmbedSubtitle `json:"subtitles"`
	Count     int                `json:"count"`
}

type ServerConfig struct {
	ID       string
	Name     string
	Icon     string
	Desc     string
	Endpoint string
}

var SERVERS = []ServerConfig{
	{ID: "NORE", Name: "Nore", Icon: "", Desc: "", Endpoint: "/api/sources/2"},
	{ID: "GORE", Name: "Gore", Icon: "", Desc: "", Endpoint: "/api/sources/3"},
	{ID: "ZORE", Name: "Zore [HD]", Icon: "", Desc: "", Endpoint: "/api/sources/1"},
	{ID: "BORE", Name: "Bore [4K]", Icon: "", Desc: "", Endpoint: "/api/sources/4"},
}
