package sdk

type (
	GetMetadataFuncType  func(Item) (Item, error)
	GetSearchResultsFunc func(Item) ([]Item, error)
)

type Agent struct {
	Identifier           string
	Name                 string
	GetMetadataFunc      GetMetadataFuncType
	GetSearchResultsFunc GetSearchResultsFunc
}

type AgentPlugin interface {
	GetIdentifier() string
	GetName() string
	GetMovieAgent() *Agent
	GetTVAgent() *Agent
	GetAlbumAgent() *Agent
}

type MovieAgent interface {
	GetIdentifier() string
	GetName() string
	GetMovieAgent() *Agent
}

type TVAgent interface {
	GetIdentifier() string
	GetName() string
	GetTVAgent() *Agent
}

type AlbumAgent interface {
	GetIdentifier() string
	GetName() string
	GetAlbumAgent() *Agent
}
