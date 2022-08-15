package client

import "context"

// MovieResult represents the result for a movie.
type MovieResult struct {
	Name   string `json:"name"`
	TmdbID string `json:"tmdb_id"` //nolint:tagliatelle // TheMovieDB uses snake_case for the field names.
	ImdbID string `json:"imdb_id"` //nolint:tagliatelle // TheMovieDB uses snake_case for the field names.

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
func (c *Client) GetMovieImages(ctx context.Context, imdbID string) (*MovieResult, error) {
	url := c.Endpoint + "/movies/" + imdbID

	var mr MovieResult
	if err := c.get(ctx, url, &mr); err != nil {
		return nil, err
	}

	return &mr, nil
}
