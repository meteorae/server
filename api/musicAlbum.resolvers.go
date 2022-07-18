package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"github.com/meteorae/meteorae-server/models"
)

func (r *musicAlbumResolver) ID(ctx context.Context, obj *models.MusicAlbum) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil //nolint:gomnd
}

func (r *musicAlbumResolver) ReleaseDate(ctx context.Context, obj *models.MusicAlbum) (*string, error) {
	YMDReleaseDate := obj.ReleaseDate.Format("2006-01-02")

	return &YMDReleaseDate, nil
}

// MusicAlbum returns models.MusicAlbumResolver implementation.
func (r *Resolver) MusicAlbum() models.MusicAlbumResolver { return &musicAlbumResolver{r} }

type musicAlbumResolver struct{ *Resolver }
