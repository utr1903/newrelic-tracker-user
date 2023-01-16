package users

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	fetch "github.com/utr1903/newrelic-tracker-internal/fetch"
	flush "github.com/utr1903/newrelic-tracker-internal/flush"
	graphql "github.com/utr1903/newrelic-tracker-internal/graphql"
	logging "github.com/utr1903/newrelic-tracker-internal/logging"
	metrics "github.com/utr1903/newrelic-tracker-internal/metrics"
	"github.com/utr1903/newrelic-tracker-user/pkg/graphql/user"
)

const (
	USERS_GRAPHQL_HAS_RETURNED_ERRORS = "graphql has returned errors"
	USERS_LOGS_COULD_NOT_BE_FORWARDED = "logs could not be forwarded"
)

const queryTemplate = `
{
	actor {
		organization {
			userManagement {
				authenticationDomains(cursor: {{ .CursorDomain }}) {
					nextCursor
					authenticationDomains {
						id
						name
						users(cursor: {{ .CursorUser }}) {
							nextCursor
							users {
								id
								name
								email
								timeZone
                emailVerificationState
                lastActive
								type {
									id
								}
							}
						}
					}
				}
			}
		}
	}
}
`

const trackedAttributeType = "users"

type queryVariables struct {
	CursorDomain string
	CursorUser   string
}

type authDomainUser struct {
	AuthDomainId           string `json:"authenticationDomainId"`
	Id                     string `json:"id"`
	Name                   string `json:"name"`
	UserType               string `json:"userType"`
	Email                  string `json:"email"`
	EmailVerificationState string `json:"emailVerificationState"`
	LastActive             string `json:"lastActive"`
	TimeZone               string `json:"timeZone"`
}

type Users struct {
	OrganizationId  string
	Logger          logging.ILogger
	Gqlc            graphql.IGraphQlClient
	MetricForwarder metrics.IMetricForwarder
}

func NewUsers(
	organizationId string,
) *Users {
	logger := logging.NewLoggerWithForwarder(
		"DEBUG",
		os.Getenv("NEWRELIC_LICENSE_KEY"),
		"https://log-api.eu.newrelic.com/log/v1",
		setCommonAttributes(organizationId),
	)
	gqlc := graphql.NewGraphQlClient(
		logger,
		"https://api.eu.newrelic.com/graphql",
		trackedAttributeType,
		queryTemplate,
	)
	mf := metrics.NewMetricForwarder(
		logger,
		os.Getenv("NEWRELIC_LICENSE_KEY"),
		"https://metric-api.eu.newrelic.com/metric/v1",
		setCommonAttributes(organizationId),
	)
	return &Users{
		OrganizationId:  organizationId,
		Logger:          logger,
		Gqlc:            gqlc,
		MetricForwarder: mf,
	}
}

func setCommonAttributes(
	organizationId string,
) map[string]string {
	return map[string]string{
		"tracker.attributeType":  trackedAttributeType,
		"tracker.organizationId": organizationId,
	}
}

func (u *Users) Run() error {

	// Fetch the unique application names per GraphQL
	authDomainUsers, err := u.fetchUsers()
	if err != nil {
		return nil
	}

	// Create & flush metrics
	err = u.flushMetrics(authDomainUsers)
	if err != nil {
		return nil
	}

	// Flush logs
	u.flushLogs()

	return nil
}

func (u *Users) fetchUsers() (
	[]authDomainUser,
	error,
) {

	var cursorDomain *string = nil
	var cursorUser *string = nil
	authDomainUsers := make([]authDomainUser, 0)

	for {

		qv := &queryVariables{
			CursorDomain: setNextCursor(cursorDomain),
			CursorUser:   setNextCursor(cursorUser),
		}

		res := &user.GraphQlUserResponse{}
		err := fetch.Fetch(
			u.Gqlc,
			qv,
			res,
		)
		if err != nil {
			return nil, err
		}

		for _, authDomain := range res.Data.Actor.Organization.UserManagement.AuthenticationDomains.AuthenticationDomains {

			// Add users
			for _, user := range authDomain.Users.Users {
				authDomainUsers = append(authDomainUsers, authDomainUser{
					AuthDomainId:           authDomain.Id,
					Id:                     user.Id,
					Name:                   user.Name,
					UserType:               user.UserType.Id,
					Email:                  user.Email,
					EmailVerificationState: user.EmailVerificationState,
					LastActive:             user.LastActive,
					TimeZone:               user.TimeZone,
				})
			}

			// Continue to fetch users until cursor is null
			nextCursorUser := authDomain.Users.NextCursor
			if nextCursorUser != nil {
				continue
			}
			cursorUser = nextCursorUser
		}

		// Continue to fetch domains until cursor is null
		nextCursorDomain := res.Data.Actor.Organization.UserManagement.AuthenticationDomains.NextCursor
		if nextCursorDomain == nil {
			break
		}
		cursorDomain = nextCursorDomain
	}
	return authDomainUsers, nil
}

func setNextCursor(
	nextCursor *string,
) string {
	if nextCursor == nil {
		return "null"
	}
	cursor := strings.Clone(*nextCursor)
	return cursor
}

func (u *Users) flushMetrics(
	authDomainUsers []authDomainUser,
) error {
	metrics := []flush.FlushMetric{}
	for _, user := range authDomainUsers {
		userType, _ := strconv.ParseFloat(user.UserType, 64)
		metrics = append(metrics, flush.FlushMetric{
			Name:  "tracker.users.type",
			Value: userType,
			Attributes: map[string]string{
				"tracker.users.authDomainId":           user.AuthDomainId,
				"tracker.users.id":                     user.Id,
				"tracker.users.name":                   user.Name,
				"tracker.users.email":                  user.Email,
				"tracker.users.emailVerificationState": user.EmailVerificationState,
				"tracker.users.lastActive":             user.LastActive,
				"tracker.users.timeZone":               user.TimeZone,
			},
		})
	}
	err := flush.Flush(u.MetricForwarder, metrics)
	if err != nil {
		return err
	}

	return nil
}

func (u *Users) flushLogs() {
	err := u.Logger.Flush()
	if err != nil {
		fmt.Println(USERS_LOGS_COULD_NOT_BE_FORWARDED, err.Error())
	}
}
