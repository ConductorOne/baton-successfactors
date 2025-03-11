package connector

import (
	"context"
	"fmt"
	"reflect"

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
	if user.EmploymentNav.EndDate != "" {
		status = v2.UserTrait_Status_STATUS_ENABLED
	}
	profile := make(map[string]interface{})
	profile["userID"] = user.UserId
	if reflect.ValueOf(user.UserNav).IsZero() {
		profile["Legal First Name"] = nil
		profile["Legal Last Name"] = nil
		profile["middle name"] = nil
		profile["username"] = nil
		profile["email"] = nil
		profile["FedRamp Authorized-User(US Persons)"] = nil
	} else {
		profile["Legal First Name"] = user.UserNav.FirstName
		profile["Legal Last Name"] = user.UserNav.LastName
		profile["middle name"] = user.UserNav.Mi
		profile["username"] = user.UserNav.Username
		profile["email"] = user.UserNav.Email
		profile["FedRamp Authorized-User(US Persons)"] = user.UserNav.Custom07
	}

	if reflect.ValueOf(user.EmploymentNav).IsZero() {
		profile["Hire Date"] = nil
		profile["Termination date"] = nil
	} else {
		profile["Hire Date"] = user.EmploymentNav.StartDate
		profile["Termination date"] = user.EmploymentNav.EndDate
	}
	profile["Job Title"] = user.JobTitle
	profile["Local Job Title"] = user.LocalJobTitle
	if reflect.ValueOf(user.CompanyNav).IsZero() {
		profile["Company name"] = nil
		profile["Country/Region"] = nil
	} else {
		profile["Company name"] = user.CompanyNav.NameLocalized
		if reflect.ValueOf(user.CompanyNav.CountryNav).IsZero() {
			profile["Country/Region"] = nil
		} else {
			profile["Country/Region"] = user.CompanyNav.CountryNav.TerritoryName
		}
	}
	if reflect.ValueOf(user.BusinessUnitNav).IsZero() {
		profile["Epicenter"] = nil
	} else {
		profile["Epicenter"] = user.BusinessUnitNav.Name
	}
	if reflect.ValueOf(user.DivisionNav).IsZero() {
		profile["Function"] = nil
	} else {
		profile["Function"] = user.DivisionNav.Name
	}
	if reflect.ValueOf(user.DepartmentNav).IsZero() {
		profile["Sub Function"] = nil
	} else {
		profile["Sub Function"] = user.DepartmentNav.Name
	}
	if reflect.ValueOf(user.LocationNav).IsZero() {
		profile["office location"] = nil
	} else {
		profile["office location"] = user.LocationNav.Name
	}
	if reflect.ValueOf(user.CostCenterNav).IsZero() {
		profile["Cost Center"] = nil
	} else {
		profile["Cost Center"] = user.CostCenterNav.NameDefaultValue
	}
	if reflect.ValueOf(user.PositionNav).IsZero() {
		profile["Position code"] = nil
		profile["Position jobTitle"] = nil
	} else {
		profile["Position code"] = user.PositionNav.Code
		profile["Position jobTitle"] = user.PositionNav.ExternalNameDefaultValue
	}
	if reflect.ValueOf(user.EmployeeClassNav).IsZero() {
		profile["Employee Class"] = nil
	} else {
		profile["Employee Class"] = user.EmployeeClassNav.PicklistLabels.Results[0].Label
	}
	if reflect.ValueOf(user.EmplStatusNav).IsZero() {
		profile["Employee Status"] = nil
	} else {
		profile["Employee Status"] = user.EmplStatusNav.PicklistLabels.Results[0].Label
	}
	if reflect.ValueOf(user.EmplStatusNav).IsZero() {
		profile["Employee Status"] = nil
	} else {
		profile["Employee Status"] = user.EmplStatusNav.PicklistLabels.Results[0].Label
	}
	if reflect.ValueOf(user.ManagerUserNav).IsZero() {
		profile["Manager userid"] = nil
		profile["Manager Email"] = nil
	} else {
		profile["Manager userid"] = user.ManagerUserNav.UserId
		profile["Manager Email"] = user.ManagerUserNav.Email
	}
	userTraitOptions := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithStatus(status),
		resource.WithUserLogin(user.UserNav.Username),
	}
	if user.UserNav.Email != "" {
		userTraitOptions = append(userTraitOptions, resource.WithEmail(user.UserNav.Email, true))
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
	users, err := o.client.GetUserData(ctx)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list users: %w", err)
	}
	i := 1
	rv := make([]*v2.Resource, 0)
	for _, user := range users {
		i++
		ur, err := userResource(user)
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, ur)
	}

	return rv, "", nil, nil
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
