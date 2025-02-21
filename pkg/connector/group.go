package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type dynamicGroupResourceType struct {
	resourceType *v2.ResourceType
	client       interface{} // Replace with your SuccessFactors client
}

func (o *dynamicGroupResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

// // newDynamicGroupBuilder creates a new connector resource for a SuccessFactors Dynamic Group.
// func NewDynamicGroupBuilder(client *client.Client, l *zap.Logger) *dynamicGroupResourceType {
// 	return &dynamicGroupResourceType{
// 		resourceType: resourceTypeGroup,
// 		client:       client,
// 	}
// }

// // Create a new connector resource for a SuccessFactors Dynamic Group.
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
	l := ctxzap.Extract(ctx)

	var rv []*v2.Resource

	// In here you must fetch and list all dynamic groups

	// nextToken := ""

	l.Debug("dynamic group listing", zap.Int("entries", len(rv)), zap.String("next_token", ""))

	return rv, "", nil, nil
}

func (o *dynamicGroupResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (o *dynamicGroupResourceType) Grants(_ context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}