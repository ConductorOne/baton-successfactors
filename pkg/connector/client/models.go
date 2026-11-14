package client

import "encoding/json"

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

type Results struct {
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

type D struct {
	Results []Results `json:"results"`
	Next    string    `json:"__next,omitempty"`
}
type SuccessFactorsUserIdList struct {
	Ds D `json:"d"`
}

// ErrorResponse is a SuccessFactors error body. The OAuth token endpoint uses
// {"errorHttpCode":"400","errorMessage":"..."}; OData endpoints use
// {"error":{"code":"...","message":{"lang":"en-US","value":"..."}}}.
type ErrorResponse struct {
	// ErrorHTTPCode is a string in the token error body. It is a raw value, so a number does not fail decoding.
	ErrorHTTPCode json.RawMessage `json:"errorHttpCode"`
	ErrorMessage  string          `json:"errorMessage"`
	Error         *struct {
		Code    string `json:"code"`
		Message struct {
			Lang  string `json:"lang"`
			Value string `json:"value"`
		} `json:"message"`
	} `json:"error"`
}

// Message returns the error text of either shape, or "" if the body has none.
func (e *ErrorResponse) Message() string {
	if e.ErrorMessage != "" {
		return e.ErrorMessage
	}
	if e.Error != nil && e.Error.Message.Value != "" {
		if e.Error.Code != "" {
			return e.Error.Code + ": " + e.Error.Message.Value
		}
		return e.Error.Message.Value
	}
	return ""
}
