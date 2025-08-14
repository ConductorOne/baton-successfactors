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
	var (
		groupResources []*v2.Resource
		nextPageLink   string
	)
	outAnnotations := annotations.Annotations{}

	dynamicGroups, nextPageLink, err := b.client.GetDynamicGroups(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list dynamicGroups: %w", err)
	}

	for _, dGroup := range dynamicGroups {
		groupResource, err := parseIntoGroupResource(dGroup)
		if err != nil {
			return nil, "", nil, err
		}

		groupResources = append(groupResources, groupResource)
	}

	return groupResources, nextPageLink, outAnnotations, nil
}

func (b *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var outAnnotations annotations.Annotations

	displayName := fmt.Sprintf("Member of %s", resource.DisplayName)
	description := fmt.Sprintf("Member of the %s dynamic group.", resource.DisplayName)

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(resourceTypeUser),
		entitlement.WithDisplayName(displayName),
		entitlement.WithDescription(description),
	}

	return []*v2.Entitlement{entitlement.NewAssignmentEntitlement(resource, groupPermissionName, assigmentOptions...)}, "", outAnnotations, nil
}

func (b *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grantResources []*v2.Grant
	outAnnotations := annotations.Annotations{}

	groupID := resource.Id.Resource

	groupMembers, err := b.client.GetDGroupMembers(ctx, groupID)
	if err != nil {
		return nil, "", outAnnotations, err
	}

	for _, groupMember := range groupMembers {
		userResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: resourceTypeUser.Id,
				Resource:     groupMember.UserId,
			},
		}

		grantResources = append(grantResources, grant.NewGrant(resource, groupPermissionName, userResource))
	}

	return grantResources, "", outAnnotations, nil
}

func parseIntoGroupResource(dynamicGroup client.DynamicGroup) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"name":                    dynamicGroup.GroupName,
		"active_membership_count": dynamicGroup.ActiveMembershipCount,
	}

	if dynamicGroup.LastModifiedDate != "" {
		profile["last_modified_date"] = extractDate(dynamicGroup.LastModifiedDate).String()
	}

	groupTraits := []rs.GroupTraitOption{
		rs.WithGroupProfile(profile),
	}

	ret, err := rs.NewGroupResource(
		dynamicGroup.GroupName,
		groupResourceType,
		dynamicGroup.GroupID,
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
