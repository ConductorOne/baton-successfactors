package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.uber.org/zap"
)

type rbproleResourceType struct {
	resourceType *v2.ResourceType
	client       interface{} //  Update the type here
	enterpriseID string
}

func (o *rbproleResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

func newRBPRoleBuilder(client interface{}, l *zap.Logger) *rbproleResourceType {
	// Update to *client.Client and import the correct package

	return &rbproleResourceType{
		resourceType: resourceTypeRBPRole,
		client:       client,
	}
}

func rbpRoleResource(
	_ context.Context,
	roleID string,
	parentResourceID *v2.ResourceId,
) (*v2.Resource, error) {
	r, err := resource.NewRoleResource(
		roleID,
		resourceTypeRBPRole,
		roleID,
		nil,
		resource.WithParentResourceID(parentResourceID))
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (o *rbproleResourceType) List(ctx context.Context, parentResourceID *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	// You would want to call something like
	// roles, _, err := o.client.ListRBPRoles(ctx, parentResourceID.Resource)
	// Where you fetch the roles specific for that workspace.
	var ret []*v2.Resource
	return ret, "", nil, nil
}

func (o *rbproleResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement
	return rv, "", nil, nil
}

// Grants would normally return the grants for each role resource. Due to how
// the Slack API works, it is more efficient to emit these roles while listing
// grants for each individual user. Instead of having to list users for each
// role we can divine which roles a user should be granted when calculating
// their grants.
func (o *rbproleResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	// TODO: (marcos) remove this function completely

	return nil, "", nil, nil
}