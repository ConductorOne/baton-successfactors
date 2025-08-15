package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
)

const groupPermissionName = "member"

type groupBuilder struct {
	client *client.SuccessFactorsClient
}

func (b *groupBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return groupResourceType
}

func (b *groupBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var groupResources []*v2.Resource
	outAnnotations := annotations.Annotations{}

	groups, nextPageIndex, rateLimitData, err := b.client.GetGroups(ctx, pToken.Token)
	if err != nil {
		if rateLimitData != nil {
			outAnnotations.WithRateLimiting(rateLimitData)
		}

		return nil, "", outAnnotations, fmt.Errorf("failed to list groups: %w", err)
	}

	for _, group := range groups {
		groupResource, err := parseIntoGroupResource(group)
		if err != nil {
			return nil, "", outAnnotations, err
		}

		groupResources = append(groupResources, groupResource)
	}

	return groupResources, nextPageIndex, outAnnotations, nil
}

func (b *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var outAnnotations annotations.Annotations

	displayName := fmt.Sprintf("Member of %s", resource.DisplayName)
	description := fmt.Sprintf("Member of the %s group.", resource.DisplayName)

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDisplayName(displayName),
		entitlement.WithDescription(description),
	}

	return []*v2.Entitlement{entitlement.NewAssignmentEntitlement(resource, groupPermissionName, assigmentOptions...)}, "", outAnnotations, nil
}

func (b *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grantResources []*v2.Grant
	outAnnotations := annotations.Annotations{}

	groupID := resource.Id.Resource

	group, rateLimitData, err := b.client.GetGroup(ctx, groupID)
	if err != nil {
		if rateLimitData != nil {
			outAnnotations.WithRateLimiting(rateLimitData)
		}
		return nil, "", outAnnotations, err
	}

	for _, groupMember := range group.Members {
		userResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     groupMember.Value,
			},
		}

		grantResources = append(grantResources, grant.NewGrant(resource, groupPermissionName, userResource))
	}

	return grantResources, "", outAnnotations, nil
}

func parseIntoGroupResource(group *client.Group) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"name": group.DisplayName,
	}

	if group.Meta != nil {
		profile["modified_on"] = group.Meta.LastModified.String()
		profile["created_on"] = group.Meta.Created.String()
		profile["membership_count"] = group.Meta.MembersCnt
	}

	groupTraits := []rs.GroupTraitOption{
		rs.WithGroupProfile(profile),
	}

	ret, err := rs.NewGroupResource(
		group.DisplayName,
		groupResourceType,
		group.Id,
		groupTraits,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func newGroupBuilder(c *client.SuccessFactorsClient) *groupBuilder {
	return &groupBuilder{
		client: c,
	}
}
