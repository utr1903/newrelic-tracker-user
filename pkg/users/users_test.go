package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/utr1903/newrelic-tracker-user/pkg/graphql/user"
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

type graphqlClientMock struct {
	failRequest bool
}

func (c *graphqlClientMock) Execute(
	qv any,
	result any,
) error {
	if c.failRequest {
		return errors.New("error")
	}

	authDomains := getAuthDomains(qv)
	res := user.GraphQlUserResponse{
		Data: user.Data{
			Actor: user.Actor{
				Organization: user.Organization{
					UserManagement: user.UserManagement{
						AuthenticationDomains: *authDomains,
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

func Test_FetchingFails(t *testing.T) {
	logger := newLoggerMock()
	gqlc := &graphqlClientMock{
		failRequest: true,
	}
	mf := &metricForwarderMock{
		returnError: true,
	}

	uas := &Users{
		OrganizationId:  "organizationId",
		Logger:          logger,
		Gqlc:            gqlc,
		MetricForwarder: mf,
	}

	err := uas.Run()

	assert.Nil(t, err)
}

func Test_FetchingSucceeds(t *testing.T) {
	logger := newLoggerMock()
	gqlc := &graphqlClientMock{
		failRequest: false,
	}
	mf := &metricForwarderMock{
		returnError: true,
	}

	users := &Users{
		OrganizationId:  "organizationId",
		Logger:          logger,
		Gqlc:            gqlc,
		MetricForwarder: mf,
	}

	_, err := users.fetchUsers()
	if err != nil {
		panic(err)
	}

	// authDomainUsers, err := users.fetchUsers()

	// for _, authDomainUser := range authDomainUsers {
	// 	fmt.Println(authDomainUser)
	// }
	// assert.Nil(t, err)
	// assert.NotNil(t, authDomainUsers)

	// for i, appName := range appNames {
	// 	assert.Equal(t, apps[i], appName)
	// }
}

// func Test_FlushingFails(t *testing.T) {
// 	logger := newLoggerMock()
// 	gqlc := &graphqlClientMock{
// 		failRequest: false,
// 	}
// 	mf := &metricForwarderMock{
// 		returnError: true,
// 	}

// 	uas := &Users{
// 		OrganizationId:  "organizationId",
// 		Logger:          logger,
// 		Gqlc:            gqlc,
// 		MetricForwarder: mf,
// 	}

// 	err := uas.Run()

// 	assert.Nil(t, err)
// }

// func Test_FlushingSucceeds(t *testing.T) {
// 	logger := newLoggerMock()
// 	gqlc := &graphqlClientMock{
// 		failRequest: false,
// 	}
// 	mf := &metricForwarderMock{
// 		returnError: false,
// 	}

// 	uas := &Users{
// 		OrganizationId:  "organizationId",
// 		Logger:          logger,
// 		Gqlc:            gqlc,
// 		MetricForwarder: mf,
// 	}

// 	err := uas.Run()

// 	assert.Nil(t, err)
// }

func parseQueryVariables(
	qv any,
) *queryVariables {
	bytes, err := json.Marshal(qv)
	if err != nil {
		panic(err)
	}

	qvParsed := &queryVariables{}
	err = json.Unmarshal(bytes, qvParsed)
	if err != nil {
		panic(err)
	}
	return qvParsed
}

func getAuthDomains(
	qv any,
) *user.AuthenticationDomains {

	qvParsed := parseQueryVariables(qv)

	// fmt.Println("DOMAIN: " + qvParsed.CursorDomain + " || USER: " + qvParsed.CursorDomain)

	// First call
	if qvParsed.CursorDomain == "null" && qvParsed.CursorUser == "null" {
		fmt.Println("--- 1 ---")
		nextCursorUser := "notnull"
		return &user.AuthenticationDomains{
			NextCursor: nil,
			AuthenticationDomains: []user.AuthenticationDomain{
				getAuthDomain1(&nextCursorUser),
			},
		}
	}
	// Second call
	if qvParsed.CursorDomain == "null" && qvParsed.CursorUser == "notnull" {
		fmt.Println("--- 2 ---")
		nextCursorDomain := "notnull"
		return &user.AuthenticationDomains{
			NextCursor: &nextCursorDomain,
			AuthenticationDomains: []user.AuthenticationDomain{
				getAuthDomain1(nil),
			},
		}
	}
	// Third call
	if qvParsed.CursorDomain == "notnull" && qvParsed.CursorUser == "null" {
		fmt.Println("--- 3 ---")
		nextCursorUser := "notnull"
		return &user.AuthenticationDomains{
			NextCursor: nil,
			AuthenticationDomains: []user.AuthenticationDomain{
				getAuthDomain2(&nextCursorUser),
			},
		}
	}
	// Fourth call
	if qvParsed.CursorDomain == "null" && qvParsed.CursorUser == "nill" {
		fmt.Println("--- 4 ---")
		return &user.AuthenticationDomains{
			NextCursor: nil,
			AuthenticationDomains: []user.AuthenticationDomain{
				getAuthDomain2(nil),
			},
		}
	}
	fmt.Println("--- NONE ---")
	return nil
}

func getAuthDomain1(
	nextCursorUser *string,
) user.AuthenticationDomain {
	dom := "authDomain1"

	users := []user.User{}
	if nextCursorUser != nil {
		users = append(users, getAuthDomain1User1(dom))
	} else {
		users = append(users, getAuthDomain1User2(dom))
	}

	authDomain1Users := user.Users{
		NextCursor: nextCursorUser,
		Users:      users,
	}
	return user.AuthenticationDomain{
		Id:    dom,
		Name:  dom,
		Users: authDomain1Users,
	}
}

func getAuthDomain2(
	nextCursorUser *string,
) user.AuthenticationDomain {
	dom := "authDomain2"

	users := []user.User{}
	if nextCursorUser != nil {
		users = append(users, getAuthDomain2User1(dom))
	} else {
		users = append(users, getAuthDomain2User2(dom))
	}

	authDomain2Users := user.Users{
		NextCursor: nextCursorUser,
		Users:      users,
	}
	return user.AuthenticationDomain{
		Id:    dom,
		Name:  dom,
		Users: authDomain2Users,
	}
}

func getAuthDomain1User1(
	authDomain1Id string,
) user.User {
	return user.User{
		Id:   authDomain1Id + "user1",
		Name: authDomain1Id + "user1",
		UserType: user.UserType{
			Id: "0",
		},
		Email:                  authDomain1Id + "email1",
		EmailVerificationState: "Verified",
		LastActive:             "2022-10-11T10:10:05Z",
		TimeZone:               "Etc/UTC",
	}
}

func getAuthDomain1User2(
	authDomain1Id string,
) user.User {
	return user.User{
		Id:   authDomain1Id + "user2",
		Name: authDomain1Id + "user2",
		UserType: user.UserType{
			Id: "1",
		},
		Email:                  authDomain1Id + "email2",
		EmailVerificationState: "Verified",
		LastActive:             "2022-10-11T10:10:05Z",
		TimeZone:               "Etc/UTC",
	}
}

func getAuthDomain2User1(
	authDomain2Id string,
) user.User {
	return user.User{
		Id:   authDomain2Id + "user1",
		Name: authDomain2Id + "user1",
		UserType: user.UserType{
			Id: "0",
		},
		Email:                  authDomain2Id + "email1",
		EmailVerificationState: "Verified",
		LastActive:             "2022-10-11T10:10:05Z",
		TimeZone:               "Etc/UTC",
	}
}

func getAuthDomain2User2(
	authDomain2Id string,
) user.User {
	return user.User{
		Id:   authDomain2Id + "user2",
		Name: authDomain2Id + "user2",
		UserType: user.UserType{
			Id: "0",
		},
		Email:                  authDomain2Id + "email2",
		EmailVerificationState: "Verified",
		LastActive:             "2022-10-11T10:10:05Z",
		TimeZone:               "Etc/UTC",
	}
}
