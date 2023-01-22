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

const queryTemplateDomains = `
{
	actor {
		organization {
			userManagement {
				authenticationDomains(cursor: {{ .Cursor }}) {
					nextCursor
					authenticationDomains {
						id
					}
				}
			}
		}
	}
}
`

const queryTemplateUsers = `
{
	actor {
		organization {
			userManagement {
				authenticationDomains(id: "{{ .AuthDomainId }}") {
					nextCursor
					authenticationDomains {
						id
						name
						users(cursor: {{ .Cursor }}) {
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

type queryVariablesDomains struct {
	Cursor string
}

type queryVariablesUsers struct {
	AuthDomainId string
	Cursor       string
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
	GqlcDomains     graphql.IGraphQlClient
	GqlcUsers       graphql.IGraphQlClient
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
	gqlcDomains := graphql.NewGraphQlClient(
		logger,
		"https://api.eu.newrelic.com/graphql",
		trackedAttributeType,
		queryTemplateDomains,
	)
	gqlcUsers := graphql.NewGraphQlClient(
		logger,
		"https://api.eu.newrelic.com/graphql",
		trackedAttributeType,
		queryTemplateUsers,
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
		GqlcDomains:     gqlcDomains,
		GqlcUsers:       gqlcUsers,
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

	// Fetch the domain IDs per GraphQL
	authDomainIds, err := u.fetchDomainIds()
	if err != nil {
		return nil
	}

	// Fetch the users per GraphQL
	authDomainUsers, err := u.fetchUsers(authDomainIds)
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

func (u *Users) fetchDomainIds() (
	[]string,
	error,
) {
	var cursor *string = nil
	authDomainIds := make([]string, 0)

	// Loop until fetching all domains in the organization
	for {

		qv := &queryVariablesDomains{
			Cursor: setNextCursor(cursor),
		}

		res := &user.GraphQlUserResponse{}
		err := fetch.Fetch(
			u.GqlcDomains,
			qv,
			res,
		)
		if err != nil {
			return nil, err
		}

		// Get the auth domain
		authDomain := res.GetAuthDomains()

		// Add domain Ids
		for _, domain := range authDomain.AuthenticationDomains {
			authDomainIds = append(authDomainIds, domain.Id)
		}

		// Continue to fetch domains until cursor is null
		cursor := authDomain.NextCursor
		if cursor == nil {
			break
		}
	}
	return authDomainIds, nil
}

func (u *Users) fetchUsers(
	authDomainIds []string,
) (
	[]authDomainUser,
	error,
) {

	var cursorUser *string = nil
	authDomainUsers := make([]authDomainUser, 0)

	// Loop over all auth domains
	for _, authDomainId := range authDomainIds {

		// Loop until fetching all users in a domain
		for {

			qv := &queryVariablesUsers{
				AuthDomainId: authDomainId,
				Cursor:       setNextCursor(cursorUser),
			}

			res := &user.GraphQlUserResponse{}
			err := fetch.Fetch(
				u.GqlcUsers,
				qv,
				res,
			)
			if err != nil {
				return nil, err
			}

			// Get the auth domain
			authDomain := res.GetAuthDomains().AuthenticationDomains[0]

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
			cursorUser := authDomain.Users.NextCursor
			if cursorUser == nil {
				break
			}
		}
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
