package client

// MovieResult represents the result for a movie.
type MovieResult struct {
	Name   string `json:"name"`
	TmdbID string `json:"tmdb_id"` //nolint:tagliatelle
	ImdbID string `json:"imdb_id"` //nolint:tagliatelle

	Arts        []*ImageInfo `json:"movieart"`
	Backgrounds []*ImageInfo `json:"moviebackground"`
	Banners     []*ImageInfo `json:"moviebanner"`
	Discs       []*ImageInfo `json:"moviedisc"`
	HDClearArts []*ImageInfo `json:"hdmovieclearart"`
	HDLogos     []*ImageInfo `json:"hdmovielogo"`
	Logos       []*ImageInfo `json:"movielogo"`
	Posters     []*ImageInfo `json:"movieposter"`
	Thumbs      []*ImageInfo `json:"moviethumb"`
}

// GetMovieImages returns the images for a movie.
func (c *Client) GetMovieImages(imdbID string) (*MovieResult, error) {
	url := c.Endpoint + "/movies/" + imdbID

	var mr MovieResult
	if err := c.get(url, &mr); err != nil {
		return nil, err
	}

	return &mr, nil
}
