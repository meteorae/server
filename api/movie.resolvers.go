package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"github.com/meteorae/meteorae-server/models"
)

func (r *movieResolver) ID(ctx context.Context, obj *models.Movie) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil //nolint:gomnd
}

func (r *movieResolver) ReleaseDate(ctx context.Context, obj *models.Movie) (*string, error) {
	YMDReleaseDate := obj.ReleaseDate.Format("2006-01-02")

	return &YMDReleaseDate, nil
}

// Movie returns models.MovieResolver implementation.
func (r *Resolver) Movie() models.MovieResolver { return &movieResolver{r} }

type movieResolver struct{ *Resolver }
