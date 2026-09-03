package anidb

type Episode struct {
	ID      int  `json:"id"`
	Number  int  `json:"number"`
	Filler  bool `json:"filler"`
}

type EpisodesResponse struct {
	Episodes []Episode `json:"episodes"`
}

type Language struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	EmbedURL string `json:"embed_url"`
}

type LanguagesResponse struct {
	Languages []Language `json:"languages"`
}
