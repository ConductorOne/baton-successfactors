package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
)

type userBuilder struct {
	client *client.SuccessFactorsClient
}

func (b *userBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return userResourceType
}

func (b *userBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var userResources []*v2.Resource
	outAnnotations := annotations.Annotations{}

	users, nextPageIndex, rateLimitData, err := b.client.GetUsers(ctx, pToken.Token)
	if err != nil {
		if rateLimitData != nil {
			outAnnotations.WithRateLimiting(rateLimitData)
		}

		return nil, "", outAnnotations, fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range users {
		userResource, err := parseIntoUserResource(user)
		if err != nil {
			return nil, "", outAnnotations, err
		}

		userResources = append(userResources, userResource)
	}

	return userResources, nextPageIndex, outAnnotations, nil
}

func (b *userBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (b *userBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func parseIntoUserResource(user *client.User) (*v2.Resource, error) {
	var userTraitOptions []resource.UserTraitOption
	var primaryEmail string

	userStatus := v2.UserTrait_Status_STATUS_DISABLED
	if user.Active {
		userStatus = v2.UserTrait_Status_STATUS_ENABLED
	}

	for _, email := range user.Emails {
		if email.Primary {
			primaryEmail = email.Value
			break
		}
	}

	displayName := user.DisplayName
	profile := map[string]interface{}{
		"first_name": user.Name.GivenName,
		"last_name":  user.Name.FamilyName,
		"title":      user.Title,
		"user_type":  user.UserType,
	}

	if user.Meta != nil {
		profile["created_on"] = user.Meta.Created.String()
		profile["updated_on"] = user.Meta.LastModified.String()
	}

	userTraitOptions = append(userTraitOptions,
		resource.WithUserProfile(profile),
		resource.WithStatus(userStatus),
	)

	if primaryEmail != "" {
		userTraitOptions = append(userTraitOptions,
			resource.WithUserLogin(primaryEmail),
			resource.WithEmail(primaryEmail, true),
		)
	} else if len(user.Emails) != 0 {
		userTraitOptions = append(userTraitOptions,
			resource.WithEmail(primaryEmail, false),
		)
	}

	newUserResource, err := resource.NewUserResource(
		displayName,
		userResourceType,
		user.Id,
		userTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return newUserResource, nil
}

func newUserBuilder(client *client.SuccessFactorsClient) *userBuilder {
	return &userBuilder{
		client: client,
	}
}
