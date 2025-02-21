package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.uber.org/zap"
)

type dynamicGroupResourceType struct {
	resourceType *v2.ResourceType
	client       interface{} //  Update the type here
	batchSize    int
}

func (o *dynamicGroupResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// newDynamicGroupBuilder creates a new connector resource for a SuccessFactors Dynamic Group.
func NewDynamicGroupBuilder(client interface{}, l *zap.Logger) *dynamicGroupResourceType {
	// Update to *client.Client and import the correct package

	return &dynamicGroupResourceType{
		resourceType: resourceTypeGroup,
		client:       client,
		batchSize:    500, //  This may need tuning for the SuccessFactors API
	}
}

// Create a new connector resource for a SuccessFactors Dynamic Group.
func dynamicGroupResource(ctx context.Context, group interface{}) (*v2.Resource, error) {
	// Replace interface{} with your actual Dynamic Group type and access the attributes
	groupID := "some generated UUID" // Replace with the actual group ID
	groupName := "someName"             // Replace with the actual group name

	profile := map[string]interface{}{
		"group_id":   groupID,
		"group_name": groupName,
	}

	groupTraitOptions := []resource.GroupTraitOption{
		resource.WithGroupProfile(profile),
	}

	ret, err := resource.NewGroupResource(
		groupName,
		resourceTypeGroup,
		groupID,
		groupTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (o *dynamicGroupResourceType) List(ctx context.Context, parentResourceID *v2.ResourceId, token *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	l := zap.L()

	page, err := parsePageToken(token.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	// In here you must fetch and list all dynamic groups
	// call something like o.client.ListDynamicGroups(ctx, params)
	// params := map[string]interface{}{
	// 	"$top":  o.batchSize,
	// 	"$skip": page.Current(),
	// 	// Add other necessary parameters, such as $select and $filter
	// }

	rv := make([]*v2.Resource, 0, o.batchSize)
	// for _, sfGroup := range groupsFromSuccessFactors {
	// 	ur, err := dynamicGroupResource(ctx, sfGroup)
	// 	if err != nil {
	// 		return nil, "", nil, err
	// 	}
	// 	rv = append(rv, ur)
	// }

	nextToken := ""
	// if len(groupsFromSuccessFactors) == o.batchSize {
	// 	nextToken = page.Next(o.batchSize).String()
	// }

	l.Debug("dynamic group listing", zap.Int("entries", len(rv)), zap.String("next_token", nextToken))

	return rv, nextToken, nil, nil
}

func (o *dynamicGroupResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	rv := []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			resource,
			"member", // const here if you are not using other values for this
			entitlement.WithGrantableTo(resourceTypeUser),
		),
	}

	return rv, "", nil, nil
}

func (o *dynamicGroupResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	// list all grants and call all users
	var rv []*v2.Grant
	// add member to the user if it's not already a member
	// ur, err := userResource(ctx, user)
	// if err != nil {
	// 	return nil, "", nil, fmt.Errorf("error creating user resource: %w", err)
	// }

	// rv = append(rv, grant.NewGrant(resource, RoleAssignmentEntitlement, ur.Id))
	return rv, "", nil, nil
}