package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type SuccessFactors struct {
	client *client.Client
}

// Metadata returns metadata about the connector.
func (s *SuccessFactors) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "SuccessFactors",
		Description: "Connector syncing users and groups from SuccessFactors to Baton.",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (s *SuccessFactors) Validate(ctx context.Context) (annotations.Annotations, error) {
	err := s.client.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("baton-successfactors: failed to authenticate. Error: %w", err)
	}

	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, client *client.Client) (*SuccessFactors, error) {
	return &SuccessFactors{
		client: client,
	}, nil
}

func (s *SuccessFactors) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	l := ctxzap.Extract(ctx)

	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(s.client, l),
		// NewDynamicGroupBuilder(s.client, l), // You'll need to create this one, and it may need two clients depending on what endpoints are used
		// NewRBPRoleBuilder(s.client, l), // You'll need to create this one
	}
}