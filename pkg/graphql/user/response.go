package user

// --- GraphQL for user query --- //
type GraphQlUserResponse struct {
	Data   Data        `json:"data"`
	Errors interface{} `json:"errors"`
}

type Data struct {
	Actor Actor `json:"actor"`
}

type Actor struct {
	Organization Organization `json:"organization"`
}

type Organization struct {
	UserManagement UserManagement `json:"userManagement"`
}

type UserManagement struct {
	AuthenticationDomains AuthenticationDomains `json:"authenticationDomains"`
}

type AuthenticationDomains struct {
	NextCursor            *string                `json:"nextCursor"`
	AuthenticationDomains []AuthenticationDomain `json:"authenticationDomains"`
}

type AuthenticationDomain struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Users Users  `json:"users"`
}

type Users struct {
	NextCursor *string `json:"nextCursor"`
	Users      []User  `json:"users"`
}

type User struct {
	Id                     string   `json:"id"`
	Name                   string   `json:"name"`
	UserType               UserType `json:"type"`
	Email                  string   `json:"email"`
	EmailVerificationState string   `json:"emailVerificationState"`
	LastActive             string   `json:"lastActive"`
	TimeZone               string   `json:"timeZone"`
}

type UserType struct {
	Id string `json:"id"`
}

func (r *GraphQlUserResponse) GetAuthDomains() AuthenticationDomains {
	return r.Data.Actor.Organization.UserManagement.AuthenticationDomains
}
