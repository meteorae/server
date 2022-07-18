package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"github.com/meteorae/meteorae-server/models"
)

func (r *personResolver) ID(ctx context.Context, obj *models.Person) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil //nolint:gomnd
}

func (r *personResolver) Type(ctx context.Context, obj *models.Person) (string, error) {
	return obj.Type.String(), nil
}

func (r *personResolver) BirthDate(ctx context.Context, obj *models.Person) (*string, error) {
	YMDReleaseDate := obj.BirthDate.Format("2006-01-02")

	return &YMDReleaseDate, nil
}

func (r *personResolver) DeathDate(ctx context.Context, obj *models.Person) (*string, error) {
	YMDReleaseDate := obj.DeathDate.Format("2006-01-02")

	return &YMDReleaseDate, nil
}

// Person returns models.PersonResolver implementation.
func (r *Resolver) Person() models.PersonResolver { return &personResolver{r} }

type personResolver struct{ *Resolver }
