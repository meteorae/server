package all

import (
	// Import all resolvers to trigger their init() functions and register them.
	_ "github.com/meteorae/meteorae-server/resolvers/image"
	_ "github.com/meteorae/meteorae-server/resolvers/imageAlbum"
	_ "github.com/meteorae/meteorae-server/resolvers/movie"
)
