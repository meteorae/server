package resolvers

import (
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/filesystem/analyzer"
	audioResolver "github.com/meteorae/meteorae-server/resolvers/audio"
	movieResolver "github.com/meteorae/meteorae-server/resolvers/movie"
	videoResolver "github.com/meteorae/meteorae-server/resolvers/video"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
)

func ResolveFile(mediaPart *database.MediaPart, library database.Library) {
	if library.Type == database.MovieLibrary || library.Type == database.TVLibrary ||
		library.Type == database.AnimeMovieLibrary || library.Type == database.AnimeTVLibrary {
		// For video-based libraries, we check if it's in our supported video extensions
		if videoResolver.IsValidVideoFile(mediaPart.FilePath) {
			// If it's a video file, analyze it early, so the workers run while we're resolving the file itself
			err := ants.Submit(func() {
				err := analyzer.AnalyzeVideo(*mediaPart)
				if err != nil {
					log.Err(err).Msgf("Could not analyze %s", mediaPart.FilePath)
				}
			})
			if err != nil {
				log.Err(err).Msgf("Could not schedule analyzis job for %s", mediaPart.FilePath)
			}

			switch library.Type {
			case database.MovieLibrary:
				// TODO: Handle movie extras
				err := movieResolver.Resolve(mediaPart, library)
				if err != nil {
					log.Error().Err(err).Msg("Could not resolve movie")
				}
			case database.TVLibrary:
				log.Error().Msg("TV libraries not yet supported")
			case database.AnimeMovieLibrary:
				log.Error().Msg("Anime movie libraries not yet supported")
			case database.AnimeTVLibrary:
				log.Error().Msg("Anime TV libraries not yet supported")
			case database.MusicLibrary:
				log.Error().Msg("Music libraries not yet supported")
			default:
				log.Error().Msgf("Unhandled video library type %s", library.Type)
			}
		} else if audioResolver.IsValidAudioFile(mediaPart.FilePath) {
			// If it's an audio file, analyze it early, so the workers run while we're resolving the file itself
			err := ants.Submit(func() {
				err := analyzer.AnalyzeAudio(*mediaPart)
				if err != nil {
					log.Err(err).Msgf("Could not analyze %s", mediaPart.FilePath)
				}
			})
			if err != nil {
				log.Err(err).Msgf("Could not schedule analyzis job for %s", mediaPart.FilePath)
			}
		}
	}
}
