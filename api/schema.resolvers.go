package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/meteorae/meteorae-server/models"
)

func (r *mutationResolver) CompleteOnboarding(ctx context.Context) (*models.ServerInfo, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ServerInfo(ctx context.Context) (*models.ServerInfo, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *subscriptionResolver) NewUpdateAvailable(ctx context.Context) (<-chan *models.UpdateInfo, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns models.MutationResolver implementation.
func (r *Resolver) Mutation() models.MutationResolver { return &mutationResolver{r} }

// Query returns models.QueryResolver implementation.
func (r *Resolver) Query() models.QueryResolver { return &queryResolver{r} }

// Subscription returns models.SubscriptionResolver implementation.
func (r *Resolver) Subscription() models.SubscriptionResolver { return &subscriptionResolver{r} }

type (
	mutationResolver     struct{ *Resolver }
	queryResolver        struct{ *Resolver }
	subscriptionResolver struct{ *Resolver }
)
