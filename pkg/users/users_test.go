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
		return errors.New("error")
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
		return errors.New("error")
	}

	res := user.GraphQlUserResponse{
		Data: user.Data{
			Actor: user.Actor{
				Organization: user.Organization{
					UserManagement: user.UserManagement{
						AuthenticationDomains: user.AuthenticationDomains{},
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
		return errors.New("error")
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

	assert.Nil(t, err)
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

	_, err := us.fetchDomainIds()

	assert.Nil(t, err)
}

func createAuthDomainUsersMock() map[string](map[string]user.User) {

	return map[string](map[string]user.User){
		dom1: map[string]user.User{
			dom1user1: user.User{
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
			dom1user2: user.User{
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
			dom2user1: user.User{
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
			dom2user2: user.User{
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
