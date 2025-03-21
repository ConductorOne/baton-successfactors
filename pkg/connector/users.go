package connector

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-successfactors/pkg/connector/client"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.SuccessFactorsClient
}

func userResource(user client.Results) (*v2.Resource, error) {
	displayName := user.UserNav.Username
	status := v2.UserTrait_Status_STATUS_DISABLED
	if user.EmploymentNav.EndDate == "" {
		status = v2.UserTrait_Status_STATUS_ENABLED
	} else {
		r, err := regexp.Compile(`-?\d+`)
		if err != nil {
			return nil, err
		}
		// Successfactors uses an outdated Date format so we have to parse it
		// "older versions of .Net framework may serialize the c# datetime object into a strange string format like /Date(1530144000000)/"
		i, err := strconv.ParseInt(r.FindString(user.EmploymentNav.EndDate), 10, 64)
		if err != nil {
			return nil, err
		}
		end := time.UnixMilli(i)
		// -1 means we are before the endDate.
		if time.Now().Compare(end) == -1 {
			status = v2.UserTrait_Status_STATUS_ENABLED
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
		resource.WithUserProfile(profile),
		resource.WithStatus(status),
		resource.WithUserLogin(user.UserNav.Username),
		resource.WithEmail(user.UserNav.Email, true),
	}
	newUserResource, err := resource.NewUserResource(
		displayName,
		resourceTypeUser,
		user.UserNav.Username,
		userTraitOptions,
	)
	if err != nil {
		return nil, err
	}
	return newUserResource, nil
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return resourceTypeUser
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	users, paginationURL, err := o.client.GetUserData(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list users: %w", err)
	}
	rv := make([]*v2.Resource, 0)
	for _, user := range users {
		ur, err := userResource(user)
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, ur)
	}

	return rv, paginationURL, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.SuccessFactorsClient) *userBuilder {
	return &userBuilder{
		resourceType: resourceTypeUser,
		client:       client,
	}
}
