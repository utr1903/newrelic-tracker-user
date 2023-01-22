package users

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/utr1903/newrelic-tracker-user/pkg/graphql/user"
)

const (
	dom1 = "dom1"
	dom2 = "dom2"

	dom1user1 = "dom1user1"
	dom1user2 = "dom1user2"
	dom2user1 = "dom2user1"
	dom2user2 = "dom2user2"
)

type loggerMock struct {
	msgs []string
}

func newLoggerMock() *loggerMock {
	return &loggerMock{
		msgs: make([]string, 0),
	}
}

func (l *loggerMock) LogWithFields(
	lvl logrus.Level,
	msg string,
	attributes map[string]string,
) {
	l.msgs = append(l.msgs, msg)
}

func (l *loggerMock) Flush() error {
	return nil
}

type graphqlClientMockDomains struct {
	failRequest bool
}

func (c *graphqlClientMockDomains) Execute(
	qv any,
	result any,
) error {
	if c.failRequest {
		return errors.New("error_fetch_domains")
	}

	qvParsed := parseQueryVariablesDomains(qv)

	var authDomainsResponse user.AuthenticationDomains
	if qvParsed.Cursor == "null" {
		nextCursor := "notnull"
		authDomainsResponse = user.AuthenticationDomains{
			NextCursor: &nextCursor,
			AuthenticationDomains: []user.AuthenticationDomain{
				{
					Id: dom1,
				},
			},
		}
	} else {
		authDomainsResponse = user.AuthenticationDomains{
			NextCursor: nil,
			AuthenticationDomains: []user.AuthenticationDomain{
				{
					Id: dom2,
				},
			},
		}
	}

	res := user.GraphQlUserResponse{
		Data: user.Data{
			Actor: user.Actor{
				Organization: user.Organization{
					UserManagement: user.UserManagement{
						AuthenticationDomains: authDomainsResponse,
					},
				},
			},
		},
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, result)
	if err != nil {
		panic(err)
	}

	return nil
}

type graphqlClientMockUsers struct {
	failRequest bool
}

func (c *graphqlClientMockUsers) Execute(
	qv any,
	result any,
) error {
	if c.failRequest {
		return errors.New("error_fetch_users")
	}

	authDomainsMock := createAuthDomainUsersMock()
	qvParsed := parseQueryVariablesUsers(qv)

	var authDomainsResponse user.AuthenticationDomains
	if qvParsed.Cursor == "null" && qvParsed.AuthDomainId == dom1 {
		nextCursor := "notnull"
		authDomainsResponse = user.AuthenticationDomains{
			NextCursor: nil,
			AuthenticationDomains: []user.AuthenticationDomain{
				{
					Id:   dom1,
					Name: dom1,
					Users: user.Users{
						NextCursor: &nextCursor,
						Users: []user.User{
							{
								Id:                     authDomainsMock[dom1][dom1user1].Id,
								Name:                   authDomainsMock[dom1][dom1user1].Name,
								UserType:               authDomainsMock[dom1][dom1user1].UserType,
								Email:                  authDomainsMock[dom1][dom1user1].Email,
								EmailVerificationState: authDomainsMock[dom1][dom1user1].EmailVerificationState,
								LastActive:             authDomainsMock[dom1][dom1user1].LastActive,
								TimeZone:               authDomainsMock[dom1][dom1user1].TimeZone,
							},
						},
					},
				},
			},
		}
	} else if qvParsed.Cursor == "notnull" && qvParsed.AuthDomainId == dom1 {
		authDomainsResponse = user.AuthenticationDomains{
			NextCursor: nil,
			AuthenticationDomains: []user.AuthenticationDomain{
				{
					Id:   dom1,
					Name: dom1,
					Users: user.Users{
						NextCursor: nil,
						Users: []user.User{
							{
								Id:                     authDomainsMock[dom1][dom1user2].Id,
								Name:                   authDomainsMock[dom1][dom1user2].Name,
								UserType:               authDomainsMock[dom1][dom1user2].UserType,
								Email:                  authDomainsMock[dom1][dom1user2].Email,
								EmailVerificationState: authDomainsMock[dom1][dom1user2].EmailVerificationState,
								LastActive:             authDomainsMock[dom1][dom1user2].LastActive,
								TimeZone:               authDomainsMock[dom1][dom1user2].TimeZone,
							},
						},
					},
				},
			},
		}
	} else if qvParsed.Cursor == "null" && qvParsed.AuthDomainId == dom2 {
		nextCursor := "notnull"
		authDomainsResponse = user.AuthenticationDomains{
			NextCursor: nil,
			AuthenticationDomains: []user.AuthenticationDomain{
				{
					Id:   dom2,
					Name: dom2,
					Users: user.Users{
						NextCursor: &nextCursor,
						Users: []user.User{
							{
								Id:                     authDomainsMock[dom2][dom2user1].Id,
								Name:                   authDomainsMock[dom2][dom2user1].Name,
								UserType:               authDomainsMock[dom2][dom2user1].UserType,
								Email:                  authDomainsMock[dom2][dom2user1].Email,
								EmailVerificationState: authDomainsMock[dom2][dom2user1].EmailVerificationState,
								LastActive:             authDomainsMock[dom2][dom2user1].LastActive,
								TimeZone:               authDomainsMock[dom2][dom2user1].TimeZone,
							},
						},
					},
				},
			},
		}
	} else if qvParsed.Cursor == "notnull" && qvParsed.AuthDomainId == dom2 {
		authDomainsResponse = user.AuthenticationDomains{
			NextCursor: nil,
			AuthenticationDomains: []user.AuthenticationDomain{
				{
					Id:   dom2,
					Name: dom2,
					Users: user.Users{
						NextCursor: nil,
						Users: []user.User{
							{
								Id:                     authDomainsMock[dom2][dom2user2].Id,
								Name:                   authDomainsMock[dom2][dom2user2].Name,
								UserType:               authDomainsMock[dom2][dom2user2].UserType,
								Email:                  authDomainsMock[dom2][dom2user2].Email,
								EmailVerificationState: authDomainsMock[dom2][dom2user2].EmailVerificationState,
								LastActive:             authDomainsMock[dom2][dom2user2].LastActive,
								TimeZone:               authDomainsMock[dom2][dom2user2].TimeZone,
							},
						},
					},
				},
			},
		}
	}

	res := user.GraphQlUserResponse{
		Data: user.Data{
			Actor: user.Actor{
				Organization: user.Organization{
					UserManagement: user.UserManagement{
						AuthenticationDomains: authDomainsResponse,
					},
				},
			},
		},
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, result)
	if err != nil {
		panic(err)
	}

	return nil
}

type metricForwarderMock struct {
	returnError bool
}

func (mf *metricForwarderMock) AddMetric(
	metricTimestamp int64,
	metricName string,
	metricType string,
	metricValue float64,
	metricAttributes map[string]string,
) {
}

func (mf *metricForwarderMock) Run() error {

	if mf.returnError {
		return errors.New("error_flush_metrics")
	}
	return nil
}

func Test_FetchingDomainsFails(t *testing.T) {
	logger := newLoggerMock()
	gqlcDomains := &graphqlClientMockDomains{
		failRequest: true,
	}
	gqlcUsers := &graphqlClientMockUsers{
		failRequest: true,
	}
	mf := &metricForwarderMock{
		returnError: true,
	}

	us := &Users{
		OrganizationId:  "organizationId",
		Logger:          logger,
		GqlcDomains:     gqlcDomains,
		GqlcUsers:       gqlcUsers,
		MetricForwarder: mf,
	}

	err := us.Run()

	assert.NotNil(t, err)
	assert.Equal(t, "error_fetch_domains", err.Error())
}

func Test_FetchingDomainsSucceeds(t *testing.T) {
	logger := newLoggerMock()
	gqlcDomains := &graphqlClientMockDomains{
		failRequest: false,
	}
	gqlcUsers := &graphqlClientMockUsers{
		failRequest: true,
	}
	mf := &metricForwarderMock{
		returnError: true,
	}

	us := &Users{
		OrganizationId:  "organizationId",
		Logger:          logger,
		GqlcDomains:     gqlcDomains,
		GqlcUsers:       gqlcUsers,
		MetricForwarder: mf,
	}

	authDomainIds, err := us.fetchDomainIds()

	assert.Nil(t, err)
	assert.Equal(t, dom1, authDomainIds[0])
	assert.Equal(t, dom2, authDomainIds[1])
}

func Test_FetchingUsersFails(t *testing.T) {
	logger := newLoggerMock()
	gqlcDomains := &graphqlClientMockDomains{
		failRequest: false,
	}
	gqlcUsers := &graphqlClientMockUsers{
		failRequest: true,
	}
	mf := &metricForwarderMock{
		returnError: true,
	}

	us := &Users{
		OrganizationId:  "organizationId",
		Logger:          logger,
		GqlcDomains:     gqlcDomains,
		GqlcUsers:       gqlcUsers,
		MetricForwarder: mf,
	}

	err := us.Run()

	assert.NotNil(t, err)
	assert.Equal(t, "error_fetch_users", err.Error())
}

func Test_FetchingUsersSucceeds(t *testing.T) {
	logger := newLoggerMock()
	gqlcDomains := &graphqlClientMockDomains{
		failRequest: false,
	}
	gqlcUsers := &graphqlClientMockUsers{
		failRequest: false,
	}
	mf := &metricForwarderMock{
		returnError: true,
	}

	us := &Users{
		OrganizationId:  "organizationId",
		Logger:          logger,
		GqlcDomains:     gqlcDomains,
		GqlcUsers:       gqlcUsers,
		MetricForwarder: mf,
	}

	authDomainIds, _ := us.fetchDomainIds()
	authDomainUsers, err := us.fetchUsers(authDomainIds)

	assert.Nil(t, err)

	// dom1user1
	assert.Equal(t, dom1, authDomainUsers[0].AuthDomainId)
	assert.Equal(t, dom1user1, authDomainUsers[0].Id)

	// dom1user2
	assert.Equal(t, dom1, authDomainUsers[1].AuthDomainId)
	assert.Equal(t, dom1user2, authDomainUsers[1].Id)

	// dom2user1
	assert.Equal(t, dom2, authDomainUsers[2].AuthDomainId)
	assert.Equal(t, dom2user1, authDomainUsers[2].Id)

	// dom2user2
	assert.Equal(t, dom2, authDomainUsers[3].AuthDomainId)
	assert.Equal(t, dom2user2, authDomainUsers[3].Id)
}

func Test_FlushingFails(t *testing.T) {
	logger := newLoggerMock()
	gqlcDomains := &graphqlClientMockDomains{
		failRequest: false,
	}
	gqlcUsers := &graphqlClientMockUsers{
		failRequest: false,
	}
	mf := &metricForwarderMock{
		returnError: true,
	}

	us := &Users{
		OrganizationId:  "organizationId",
		Logger:          logger,
		GqlcDomains:     gqlcDomains,
		GqlcUsers:       gqlcUsers,
		MetricForwarder: mf,
	}

	err := us.Run()

	assert.NotNil(t, err)
	assert.Equal(t, "error_flush_metrics", err.Error())
}

func Test_FlushingSucceeds(t *testing.T) {
	logger := newLoggerMock()
	gqlcDomains := &graphqlClientMockDomains{
		failRequest: false,
	}
	gqlcUsers := &graphqlClientMockUsers{
		failRequest: false,
	}
	mf := &metricForwarderMock{
		returnError: false,
	}

	us := &Users{
		OrganizationId:  "organizationId",
		Logger:          logger,
		GqlcDomains:     gqlcDomains,
		GqlcUsers:       gqlcUsers,
		MetricForwarder: mf,
	}

	err := us.Run()

	assert.Nil(t, err)
}

func createAuthDomainUsersMock() map[string](map[string]user.User) {

	return map[string](map[string]user.User){
		dom1: map[string]user.User{
			dom1user1: {
				Id:   dom1user1,
				Name: dom1user1,
				UserType: user.UserType{
					Id: "1",
				},
				Email:                  dom1user1,
				EmailVerificationState: "Verified",
				LastActive:             "2022-10-11T10:10:05Z",
				TimeZone:               "Etc/UTC",
			},
			dom1user2: {
				Id:   dom1user2,
				Name: dom1user2,
				UserType: user.UserType{
					Id: "0",
				},
				Email:                  dom1user2,
				EmailVerificationState: "Verified",
				LastActive:             "2022-10-11T10:10:05Z",
				TimeZone:               "Etc/UTC",
			},
		},
		dom2: map[string]user.User{
			dom2user1: {
				Id:   dom2user1,
				Name: dom2user1,
				UserType: user.UserType{
					Id: "0",
				},
				Email:                  dom2user1,
				EmailVerificationState: "Verified",
				LastActive:             "2022-10-11T10:10:05Z",
				TimeZone:               "Etc/UTC",
			},
			dom2user2: {
				Id:   dom2user2,
				Name: dom2user2,
				UserType: user.UserType{
					Id: "0",
				},
				Email:                  dom2user2,
				EmailVerificationState: "Verified",
				LastActive:             "2022-10-11T10:10:05Z",
				TimeZone:               "Etc/UTC",
			},
		},
	}
}

func parseQueryVariablesDomains(
	qv any,
) *queryVariablesDomains {
	bytes, err := json.Marshal(qv)
	if err != nil {
		panic(err)
	}

	qvParsed := &queryVariablesDomains{}
	err = json.Unmarshal(bytes, qvParsed)
	if err != nil {
		panic(err)
	}
	return qvParsed
}

func parseQueryVariablesUsers(
	qv any,
) *queryVariablesUsers {
	bytes, err := json.Marshal(qv)
	if err != nil {
		panic(err)
	}

	qvParsed := &queryVariablesUsers{}
	err = json.Unmarshal(bytes, qvParsed)
	if err != nil {
		panic(err)
	}
	return qvParsed
}
