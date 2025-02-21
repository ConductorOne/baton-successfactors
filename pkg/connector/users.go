package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type userResourceType struct {
	resourceType *v2.ResourceType
	client       *client.Client
	batchSize    int
}

func (o *userResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

func userBuilder(client *client.Client, l *zap.Logger) *userResourceType {
	return &userResourceType{
		resourceType: resourceTypeUser,
		client:       client,
		batchSize:    500, //  this might need tuning for the SuccessFactors API
	}
}

// Create a new connector resource for a SuccessFactors user.
func userResource(ctx context.Context, user interface{}) (*v2.Resource, error) {
	// Replace interface{} with your actual User type and access the attributes
	// TODO: Look into all those attributes that should be in the userTraits
	userId := "some generated UUID" // Replace with the actual user ID
	userName := "someName"             // Replace with the actual username
	profile := map[string]interface{}{
		"user_id":   userId,
		"user_name": userName,
	}

	userTraitOptions := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithEmail(userName, true), // Replace with actual email if available
		resource.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),      // Replace with actual status logic
	}

	ret, err := resource.NewUserResource(
		userName,
		resourceTypeUser,
		userId,
		userTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (o *userResourceType) List(ctx context.Context, parentResourceID *v2.ResourceId, token *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	page, err := parsePageToken(token.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	params := map[string]interface{}{
		"$top":  o.batchSize,
		"$skip": page.Current(),
		// Add other necessary parameters, such as $select and $filter
	}

	users, rl, err := o.client.GetUsers(ctx, params)
	if err != nil {
		return nil, "", annotations.New(&rl), fmt.Errorf("error fetching users: %w", err)
	}

	var rv []*v2.Resource
	for _, user := range users {
		ur, err := userResource(ctx, user)
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, ur)
	}

	nextToken := ""
	if len(users) == o.batchSize {
		nextToken = page.Next(o.batchSize).String()
	}

	l.Debug("user listing", zap.Int("entries", len(rv)), zap.String("next_token", nextToken))

	return rv, nextToken, nil, nil
}

func (o *userResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (o *userResourceType) Grants(_ context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.Client, l *zap.Logger) *userResourceType {
	return &userResourceType{
		resourceType: resourceTypeUser,
		client:       client,
		batchSize:    500, // You can adjust the batch size as needed
	}
}