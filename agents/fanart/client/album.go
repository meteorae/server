package client

type Album struct {
	CDArt      []*ImageInfo `json:"cdart"`
	AlbumCover []*ImageInfo `json:"albumcover"`
}

// AlbumResult represents the result for an album.
type AlbumResult struct {
	Name string `json:"name"`
	MbID string `json:"mbid_id"` //nolint:tagliatelle

	Albums map[string]Album `json:"albums"`
}

// GetAlbumImages returns the images for a show.
func (c *Client) GetAlbumImages(mbID string) (*AlbumResult, error) {
	url := c.Endpoint + "/music/albums/" + mbID

	var ar AlbumResult
	if err := c.get(url, &ar); err != nil {
		return nil, err
	}

	return &ar, nil
}
