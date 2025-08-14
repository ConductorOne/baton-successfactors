package client

import (
	"encoding/json"
)

type Bearer struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
type Metadata struct {
	URIString string `json:"uri"`
	Type      string `json:"type"`
}

type PicklistLabelsResults struct {
	Metadata Metadata `json:"__metadata"`
	Label    string   `json:"label"`
}

type PicklistLabels struct {
	Results []PicklistLabelsResults `json:"results"`
}

type EmplStatusNav struct {
	Metadata       Metadata       `json:"__metadata"`
	PicklistLabels PicklistLabels `json:"picklistLabels"`
}

type UserNav struct {
	Metadata  Metadata `json:"__metadata"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Custom07  string   `json:"custom07"`
	Mi        string   `json:"mi"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
}

type BusinessUnitNav struct {
	Metadata Metadata `json:"__metadata"`
	Name     string   `json:"name"`
}

type LocationNav struct {
	Metadata Metadata `json:"__metadata"`
	Name     string   `json:"name"`
}

type EmploymentNav struct {
	Metadata  Metadata `json:"__metadata"`
	EndDate   string   `json:"endDate"`
	StartDate string   `json:"startDate"`
}

type DivisionNav struct {
	Metadata Metadata `json:"__metadata"`
	Name     string   `json:"name"`
}

type PositionNav struct {
	Metadata                 Metadata `json:"__metadata"`
	Code                     string   `json:"code"`
	ExternalNameDefaultValue string   `json:"externalName_defaultValue"`
}

type CostCenterNav struct {
	Metadata         Metadata `json:"__metadata"`
	NameDefaultValue string   `json:"name_defaultValue"`
}
type EmployeeClassNav struct {
	Metadata       Metadata       `json:"__metadata"`
	PicklistLabels PicklistLabels `json:"picklistLabels"`
}

type DepartmentNav struct {
	Metadata Metadata `json:"__metadata"`
	Name     string   `json:"name"`
}

type ManagerUserNav struct {
	Metadata Metadata `json:"__metadata"`
	UserId   string   `json:"userId"`
	Email    string   `json:"email"`
}
type CountryNav struct {
	Metadata      Metadata `json:"__metadata"`
	TerritoryName string   `json:"territoryName"`
}
type CompanyNav struct {
	Metadata      Metadata   `json:"__metadata"`
	NameLocalized string     `json:"name_localized"`
	CountryNav    CountryNav `json:"countryNav"`
}

type User struct {
	Metadata         Metadata         `json:"__metadata"`
	UserId           string           `json:"userId"`
	JobTitle         string           `json:"jobTitle"`
	LocalJobTitle    string           `json:"localJobTitle"`
	EmplStatusNav    EmplStatusNav    `json:"emplStatusNav"`
	UserNav          UserNav          `json:"userNav"`
	BusinessUnitNav  BusinessUnitNav  `json:"businessUnitNav"`
	LocationNav      LocationNav      `json:"locationNav"`
	EmploymentNav    EmploymentNav    `json:"employmentNav"`
	DivisionNav      DivisionNav      `json:"divisionNav"`
	PositionNav      PositionNav      `json:"positionNav"`
	CostCenterNav    CostCenterNav    `json:"costCenterNav"`
	EmployeeClassNav EmployeeClassNav `json:"employeeClassNav"`
	DepartmentNav    DepartmentNav    `json:"departmentNav"`
	ManagerUserNav   ManagerUserNav   `json:"managerUserNav"`
	CompanyNav       CompanyNav       `json:"companyNav"`
}

type Response struct {
	Ds struct {
		Results json.RawMessage `json:"results"`
		Next    string          `json:"__next,omitempty"`
	} `json:"d"`
}

type DynamicGroup struct {
	Metadata              Metadata `json:"__metadata"`
	GroupID               string   `json:"groupID"`
	GroupName             string   `json:"groupName"`
	GroupType             string   `json:"groupType"`
	ActiveMembershipCount int      `json:"activeMembershipCount"`
	CreatedBy             string   `json:"createdBy"`
	LastModifiedDate      string   `json:"lastModifiedDate"`
	UserType              string   `json:"userType,omitempty"`
	TotalMemberCount      int      `json:"totalMemberCount"`
	DgExcludePools        struct {
		Deferred struct {
			Uri string `json:"uri"`
		} `json:"__deferred"`
	} `json:"dgExcludePools"`
	DgIncludePools struct {
		Deferred struct {
			Uri string `json:"uri"`
		} `json:"__deferred"`
	} `json:"dgIncludePools"`
}

type DGroupMemberResponse struct {
	Ds []DGroupMember `json:"d"`
}

type DGroupMember struct {
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	MiddleName string `json:"middleName,omitempty"`
	PersonGUID string `json:"personGUID,omitempty"`
	UserId     string `json:"userId,omitempty"`
	UserName   string `json:"userName,omitempty"`
}

type ErrorResponse struct {
	Schemas  []string `json:"schemas"`
	ScimType string   `json:"scimType"`
	Detail   string   `json:"detail"`
	Status   int      `json:"status"`
}

func (er *ErrorResponse) Message() string {
	if er.Detail == "" {
		return "Error response empty"
	}
	return er.Detail
}
