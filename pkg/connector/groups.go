package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	sdkEntitlement "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	sdkGrant "github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
)

type groupBuilder struct {
	resourceType *v2.ResourceType
	client       *client.SuccessFactorsClient
}

func groupResource(group client.DynamicGroupResult) (*v2.Resource, error) {
	groupID := group.GroupID.String()

	profile := map[string]interface{}{
		"groupID":          groupID,
		"groupName":        group.GroupName,
		"groupType":        group.GroupType,
		"staticGroup":      group.StaticGroup,
		"totalMemberCount": group.TotalMemberCount.String(),
		"createdBy":        group.CreatedBy,
	}

	groupTraitOptions := []resource.GroupTraitOption{
		resource.WithGroupProfile(profile),
	}

	newGroupResource, err := resource.NewGroupResource(
		group.GroupName,
		resourceTypeGroup,
		groupID,
		groupTraitOptions,
	)
	if err != nil {
		return nil, err
	}
	return newGroupResource, nil
}

func (g *groupBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return resourceTypeGroup
}

// List returns all groups from SuccessFactors as resource objects.
func (g *groupBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	groups, paginationURL, err := g.client.GetGroups(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list groups: %w", err)
	}

	var rv []*v2.Resource
	for _, group := range groups {
		gr, err := groupResource(group)
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, gr)
	}

	return rv, paginationURL, nil, nil
}

// Entitlements returns the "member" entitlement for each group.
func (g *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	entitlement := sdkEntitlement.NewAssignmentEntitlement(
		resource,
		"member",
		sdkEntitlement.WithDisplayName("Member"),
		sdkEntitlement.WithDescription(fmt.Sprintf("Member of %s group in SuccessFactors", resource.DisplayName)),
		sdkEntitlement.WithGrantableTo(resourceTypeUser),
	)

	return []*v2.Entitlement{entitlement}, "", nil, nil
}

// Grants returns the users who are members of the group by calling the expanded group API.
func (g *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	groupID := resource.Id.Resource

	expanded, err := g.client.GetExpandedGroup(ctx, groupID)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to get expanded group %s: %w", groupID, err)
	}

	var grants []*v2.Grant
	for _, pool := range expanded.DgIncludePools.Results {
		for _, user := range pool.Users.Results {
			if user.UserID == "" {
				continue
			}
			grant := sdkGrant.NewGrant(
				resource,
				"member",
				&v2.ResourceId{
					ResourceType: resourceTypeUser.Id,
					Resource:     user.UserID,
				},
			)
			grants = append(grants, grant)
		}
	}

	return grants, "", nil, nil
}

func newGroupBuilder(client *client.SuccessFactorsClient) *groupBuilder {
	return &groupBuilder{
		resourceType: resourceTypeGroup,
		client:       client,
	}
}
