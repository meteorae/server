package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"github.com/meteorae/meteorae-server/models"
)

func (r *photoAlbumResolver) ID(ctx context.Context, obj *models.PhotoAlbum) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil //nolint:gomnd
}

func (r *photoAlbumResolver) ReleaseDate(ctx context.Context, obj *models.PhotoAlbum) (*string, error) {
	YMDDate := obj.ReleaseDate.Format("2006-01-02")

	return &YMDDate, nil
}

// PhotoAlbum returns models.PhotoAlbumResolver implementation.
func (r *Resolver) PhotoAlbum() models.PhotoAlbumResolver { return &photoAlbumResolver{r} }

type photoAlbumResolver struct{ *Resolver }
