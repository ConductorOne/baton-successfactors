package connector

import (
	"context"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
)

type Connector struct {
	ctx         context.Context
	instanceUrl string
	client      *client.SuccessFactorsClient
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "My Baton Connector",
		Description: "The template implementation of a baton connector",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context,
	compId string,
	clientId string,
	pubkey string,
	privkey string,
	instanceURL string,
	issuerURL string,
	subjectNameId string,
	samlapikey string,
) (*Connector, error) {
	SuccessFactorsClient, err := client.New(
		ctx,
		instanceURL,
		compId,
		clientId,
		pubkey,
		privkey,
		issuerURL,
		subjectNameId,
		samlapikey,
	)
	if err != nil {
		return nil, err
	}
	connector := Connector{
		client:      SuccessFactorsClient,
		ctx:         ctx,
		instanceUrl: instanceURL,
	}
	return &connector, nil
}
