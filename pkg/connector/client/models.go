package client

type SuccessFactorUser struct {
	//todo
}
type Bearer struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
type __metadata struct {
	Uri  string `json:"uri"`
	Type string `json:"type"`
}

type picklistLabelsResults struct {
	Metadata __metadata `json:"__metadata"`
	Label    string     `json:"label"`
}

type picklistLabels struct {
	Results []picklistLabelsResults `json:"results"`
}

type emplStatusNav struct {
	Metadata       __metadata     `json:"__metadata"`
	PicklistLabels picklistLabels `json:"picklistLabels"`
}

type userNav struct {
	Metadata  __metadata `json:"__metadata"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Custom07  string     `json:"custom07"`
	Mi        string     `json:"mi"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
}

type businessUnitNav struct {
	Metadata __metadata `json:"__metadata"`
	Name     string     `json:"name"`
}

type locationNav struct {
	Metadata __metadata `json:"__metadata"`
	Name     string     `json:"name"`
}

type employmentNav struct {
	Metadata  __metadata `json:"__metadata"`
	EndDate   string     `json:"endDate"`
	StartDate string     `json:"startDate"`
}

type divisionNav struct {
	Metadata __metadata `json:"__metadata"`
	Name     string     `json:"name"`
}

type positionNav struct {
	Metadata                  __metadata `json:"__metadata"`
	Code                      string     `json:"code"`
	ExternalName_defaultValue string     `json:"externalName_defaultValue"`
}

type costCenterNav struct {
	Metadata          __metadata `json:"__metadata"`
	Name_defaultValue string     `json:"name_defaultValue"`
}
type __deferred struct {
	Uri string `json:"uri"`
}
type employeeClassNav struct {
	Metadata       __metadata     `json:"__metadata"`
	PicklistLabels picklistLabels `json:"picklistLabels"`
}

type departmentNav struct {
	Metadata __metadata `json:"__metadata"`
	Name     string     `json:"name"`
}

type managerUserNav struct {
	Metadata __metadata `json:"__metadata"`
	UserId   string     `json:"userId"`
	Email    string     `json:"email"`
}
type countryNav struct {
	Metadata      __metadata `json:"__metadata"`
	TerritoryName string     `json:"territoryName"`
}
type companyNav struct {
	Metadata       __metadata `json:"__metadata"`
	Name_localized string     `json:"name_localized"`
	CountryNav     countryNav `json:"countryNav"`
}

type Results struct {
	Metadata         __metadata       `json:"__metadata"`
	UserId           string           `json:"userId"`
	JobTitle         string           `json:"jobTitle"`
	LocalJobTitle    string           `json:"localJobTitle"`
	EmplStatusNav    emplStatusNav    `json:"emplStatusNav"`
	UserNav          userNav          `json:"userNav"`
	BusinessUnitNav  businessUnitNav  `json:"businessUnitNav"`
	LocationNav      locationNav      `json:"locationNav"`
	EmploymentNav    employmentNav    `json:"employmentNav"`
	DivisionNav      divisionNav      `json:"divisionNav"`
	PositionNav      positionNav      `json:"positionNav"`
	CostCenterNav    costCenterNav    `json:"costCenterNav"`
	EmployeeClassNav employeeClassNav `json:"employeeClassNav"`
	DepartmentNav    departmentNav    `json:"departmentNav"`
	ManagerUserNav   managerUserNav   `json:"managerUserNav"`
	CompanyNav       companyNav       `json:"companyNav"`
}

type D struct {
	Results []Results `json:"results"`
	Next    string    `json:"__next,omitempty"`
}
type SuccessFactorsUserIdList struct {
	Ds D `json:"d"`
}
