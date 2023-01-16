package user

// --- GraphQL for user query --- //
type GraphQlUserResponse struct {
	Data   data        `json:"data"`
	Errors interface{} `json:"errors"`
}

type data struct {
	Actor actor `json:"actor"`
}

type actor struct {
	Organization organization `json:"organization"`
}

type organization struct {
	UserManagement userManagement `json:"userManagement"`
}

type userManagement struct {
	AuthenticationDomains authenticationDomains `json:"authenticationDomains"`
}

type authenticationDomains struct {
	NextCursor            *string                `json:"nextCursor"`
	AuthenticationDomains []authenticationDomain `json:"authenticationDomains"`
}

type authenticationDomain struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Users users  `json:"users"`
}

type users struct {
	NextCursor *string `json:"nextCursor"`
	Users      []user  `json:"users"`
}

type user struct {
	Id                     string   `json:"id"`
	Name                   string   `json:"name"`
	UserType               userType `json:"type"`
	Email                  string   `json:"email"`
	EmailVerificationState string   `json:"emailVerificationState"`
	LastActive             string   `json:"lastActive"`
	TimeZone               string   `json:"timeZone"`
}

type userType struct {
	Id string `json:"id"`
}
