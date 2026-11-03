package connector

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.SuccessFactorsClient
}

func userResource(user client.Results) (*v2.Resource, error) {
	displayName := user.UserNav.Username
	status := v2.Status_RESOURCE_STATUS_DISABLED
	if user.EmploymentNav.EndDate == "" {
		status = v2.Status_RESOURCE_STATUS_ENABLED
	} else {
		r := regexp.MustCompile(`-?\d+`)
		// Successfactors uses an outdated Date format so we have to parse it
		// "older versions of .Net framework may serialize the c# datetime object into a strange string format like /Date(1530144000000)/"
		i, err := strconv.ParseInt(r.FindString(user.EmploymentNav.EndDate), 10, 64)
		if err != nil {
			return nil, err
		}
		end := time.UnixMilli(i)
		// -1 means we are before the endDate.
		if time.Now().Compare(end) == -1 {
			status = v2.Status_RESOURCE_STATUS_ENABLED
		}
	}
	profile := make(map[string]interface{})
	// If any of these are nil then the db will be populated with empty strings for them.
	profile["userID"] = user.UserId
	profile["Legal First Name"] = user.UserNav.FirstName
	profile["Legal Last Name"] = user.UserNav.LastName
	profile["middle name"] = user.UserNav.Mi
	profile["username"] = user.UserNav.Username
	profile["email"] = user.UserNav.Email
	profile["FedRamp Authorized-User(US Persons)"] = user.UserNav.Custom07
	profile["Hire Date"] = user.EmploymentNav.StartDate
	profile["Termination date"] = user.EmploymentNav.EndDate
	profile["Job Title"] = user.JobTitle
	profile["Local Job Title"] = user.LocalJobTitle
	profile["Company name"] = user.CompanyNav.NameLocalized
	profile["Country/Region"] = user.CompanyNav.CountryNav.TerritoryName
	profile["Epicenter"] = user.BusinessUnitNav.Name
	profile["Function"] = user.DivisionNav.Name
	profile["Sub Function"] = user.DepartmentNav.Name
	profile["office location"] = user.LocationNav.Name
	profile["Cost Center"] = user.CostCenterNav.NameDefaultValue
	profile["Position code"] = user.PositionNav.Code
	profile["Position jobTitle"] = user.PositionNav.ExternalNameDefaultValue
	profile["Manager userid"] = user.ManagerUserNav.UserId
	profile["Manager Email"] = user.ManagerUserNav.Email

	// Results needs to be checked before indexing
	if len(user.EmployeeClassNav.PicklistLabels.Results) == 0 {
		profile["Employee Class"] = ""
	} else {
		profile["Employee Class"] = user.EmployeeClassNav.PicklistLabels.Results[0].Label
	}

	if len(user.EmplStatusNav.PicklistLabels.Results) == 0 {
		profile["Employee Status"] = ""
	} else {
		profile["Employee Status"] = user.EmplStatusNav.PicklistLabels.Results[0].Label
	}

	userTraitOptions := []resource.UserTraitOption{
		resource.WithUserLogin(user.UserNav.Username),
		resource.WithEmail(user.UserNav.Email, true),
	}
	newUserResource, err := resource.NewUserResource(
		displayName,
		resourceTypeUser,
		user.UserNav.Username,
		userTraitOptions,
		resource.WithResourceProfile(profile),
		resource.WithResourceStatus(status, ""),
	)
	if err != nil {
		return nil, err
	}
	return newUserResource, nil
}

// redactPageToken removes the page token from msg. The token is a full __next URL
// with a $skiptoken query parameter. It replaces the raw and escaped forms of the
// token, and the $skiptoken value on its own.
func redactPageToken(msg, token string) string {
	if token == "" {
		return msg
	}
	for _, t := range []string{token, url.QueryEscape(token), url.PathEscape(token)} {
		msg = strings.ReplaceAll(msg, t, "<page-token>")
	}
	if u, err := url.Parse(token); err == nil {
		if v := u.Query().Get("$skiptoken"); v != "" {
			for _, t := range []string{v, url.QueryEscape(v), url.PathEscape(v)} {
				msg = strings.ReplaceAll(msg, t, "<redacted>")
			}
		}
	}
	return msg
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return resourceTypeUser
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resource.SyncOpAttrs) ([]*v2.Resource, *resource.SyncOpResults, error) {
	users, paginationURL, err := o.client.GetUserData(ctx, opts.PageToken.Token)
	if err != nil {
		// A 404 from the user list must fail the sync. The SDK treats NotFound as a
		// warning and would end the sync with a partial list, so C1 would delete the
		// missing users. FailedPrecondition is not retried (Unavailable is, without
		// limit). The error is formatted with %v on purpose: %w would let status.Code
		// find the inner NotFound again. The inner text is kept for debugging, with the
		// page token removed: it is a URL, and the server could echo it in the response body.
		if status.Code(err) == codes.NotFound {
			return nil, nil, status.Errorf(codes.FailedPrecondition,
				"baton-successfactors: user list returned not found (page token set: %t): %s",
				opts.PageToken.Token != "", redactPageToken(err.Error(), opts.PageToken.Token))
		}
		// The error text can include a SuccessFactors error body that echoes the request URL.
		// Keep the gRPC code, so that retries and warnings behave as before, and remove the
		// page token from the text. %w is not used: it would keep the original text.
		return nil, nil, status.Errorf(status.Code(err), "baton-successfactors: failed to list users: %s",
			redactPageToken(err.Error(), opts.PageToken.Token))
	}
	rv := make([]*v2.Resource, 0)
	for _, user := range users {
		ur, err := userResource(user)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-successfactors: failed to build user resource: %w", err)
		}
		rv = append(rv, ur)
	}

	return rv, &resource.SyncOpResults{NextPageToken: paginationURL}, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ resource.SyncOpAttrs) ([]*v2.Entitlement, *resource.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(_ context.Context, _ *v2.Resource, _ resource.SyncOpAttrs) ([]*v2.Grant, *resource.SyncOpResults, error) {
	return nil, nil, nil
}

func newUserBuilder(client *client.SuccessFactorsClient) *userBuilder {
	return &userBuilder{
		resourceType: resourceTypeUser,
		client:       client,
	}
}
