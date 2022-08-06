package client

// ShowResult represents the result for a show.
type ShowResult struct {
	Name   string `json:"name"`
	TvdbID string `json:"thetvdb_id"` //nolint:tagliatelle

	Backgrounds   []*ImageInfo `json:"showbackground"`
	Banners       []*ImageInfo `json:"tvbanner"`
	CharacterArts []*ImageInfo `json:"characterart"`
	ClearArts     []*ImageInfo `json:"clearart"`
	ClearLogos    []*ImageInfo `json:"clearlogo"`
	HDClearArts   []*ImageInfo `json:"hdclearart"`
	HDTVLogos     []*ImageInfo `json:"hdtvlogo"`
	Posters       []*ImageInfo `json:"tvposter"`
	SeasonBanners []*ImageInfo `json:"seasonbanner"`
	SeasonPosters []*ImageInfo `json:"seasonposter"`
	SeasonThumbs  []*ImageInfo `json:"seasonthumb"`
	Thumbs        []*ImageInfo `json:"tvthumb"`
}

// GetShowImages returns the images for a show.
func (c *Client) GetShowImages(tvdbID string) (*ShowResult, error) {
	url := c.Endpoint + "/tv/" + tvdbID

	var sr ShowResult
	if err := c.get(url, &sr); err != nil {
		return nil, err
	}

	return &sr, nil
}
