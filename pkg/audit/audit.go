package audit

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	fetch "github.com/utr1903/newrelic-tracker-internal/fetch"
	flush "github.com/utr1903/newrelic-tracker-internal/flush"
	graphql "github.com/utr1903/newrelic-tracker-internal/graphql"
	logging "github.com/utr1903/newrelic-tracker-internal/logging"
	metrics "github.com/utr1903/newrelic-tracker-internal/metrics"
	nrql "github.com/utr1903/newrelic-tracker-user/pkg/graphql/nrql"
)

const (
	AUDIT_EVENTS_GRAPHQL_HAS_RETURNED_ERRORS = "graphql has returned errors"
	AUDIT_EVENTS_LOGS_COULD_NOT_BE_FORWARDED = "logs could not be forwarded"
)

const queryTemplate = `
{
  actor {
    nrql(
			accounts: {{ .AccountId }},
			query: "{{ .NrqlQuery }}"
		) {
      results
    }
  }
}
`

const trackedAttributeType = "auditEvent"

type queryVariables struct {
	AccountId int64
	NrqlQuery string
}

type auditEvent struct {
	ActionIdentifier string `json:"actionIdentifier"`
	ActorEmail       string `json:"actorEmail"`
	ActorId          string `json:"actorId"`
	ActorType        string `json:"actorType"`
	Description      string `json:"description"`
	Id               string `json:"id"`
	ScopeId          string `json:"scopeId"`
	ScopeType        string `json:"scopeType"`
	TargetId         string `json:"targetId"`
	TargetType       string `json:"targetType"`
	Timestamp        int64  `json:"timestamp"`
}

type AuditEvent struct {
	AccountId       int64
	Logger          logging.ILogger
	Gqlc            graphql.IGraphQlClient
	MetricForwarder metrics.IMetricForwarder
}

func NewAuditEvents(
	organizationId string,
	accountId int64,
) *AuditEvent {
	logger := logging.NewLoggerWithForwarder(
		"DEBUG",
		os.Getenv("NEWRELIC_LICENSE_KEY"),
		"https://log-api.eu.newrelic.com/log/v1",
		setCommonAttributes(organizationId, accountId),
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
		setCommonAttributes(organizationId, accountId),
	)
	return &AuditEvent{
		AccountId:       accountId,
		Logger:          logger,
		Gqlc:            gqlc,
		MetricForwarder: mf,
	}
}

func setCommonAttributes(
	organizationId string,
	accountId int64,
) map[string]string {
	return map[string]string{
		"tracker.attributeType":  trackedAttributeType,
		"tracker.organizationId": organizationId,
		"tracker.accountId":      strconv.FormatInt(accountId, 10),
	}
}

func (a *AuditEvent) Run() error {

	// Fetch audit events per GraphQL
	auditEvents, err := a.fetchAuditEvents()
	if err != nil {
		return err
	}

	// Create & flush metrics
	err = a.flushMetrics(auditEvents)
	if err != nil {
		return err
	}

	// Flush logs
	a.flushLogs()

	return nil
}

func (a *AuditEvent) fetchAuditEvents() (
	[]auditEvent,
	error,
) {
	qv := &queryVariables{
		AccountId: a.AccountId,
		NrqlQuery: "FROM NrAuditEvent SELECT * SINCE 1 day ago LIMIT MAX",
	}

	res := &nrql.GraphQlNrqlResponse[auditEvent]{}
	err := fetch.Fetch(
		a.Gqlc,
		qv,
		res,
	)
	if err != nil {
		return nil, err
	}
	if res.Errors != nil {
		a.Logger.LogWithFields(logrus.DebugLevel, AUDIT_EVENTS_GRAPHQL_HAS_RETURNED_ERRORS,
			map[string]string{
				"tracker.package": "pkg.audit",
				"tracker.file":    "audit.go",
				"tracker.error":   fmt.Sprintf("%v", res.Errors),
			})
		return nil, errors.New(AUDIT_EVENTS_GRAPHQL_HAS_RETURNED_ERRORS)
	}

	return res.Data.Actor.Nrql.Results, nil
}

func (a *AuditEvent) flushMetrics(
	auditEvents []auditEvent,
) error {
	metrics := []flush.FlushMetric{}
	for _, auditEvent := range auditEvents {
		metrics = append(metrics, flush.FlushMetric{
			Name:      "tracker.users.audit.value",
			Value:     1.0,
			Timestamp: auditEvent.Timestamp,
			Attributes: map[string]string{
				"tracker.users.audit.actionIdentifier": auditEvent.ActionIdentifier,
				"tracker.users.audit.actorEmail":       auditEvent.ActorEmail,
				"tracker.users.audit.actorId":          auditEvent.ActorId,
				"tracker.users.audit.actorType":        auditEvent.ActorType,
				"tracker.users.audit.description":      auditEvent.Description,
				"tracker.users.audit.id":               auditEvent.Id,
				"tracker.users.audit.scopeId":          auditEvent.ScopeId,
				"tracker.users.audit.scopeType":        auditEvent.ScopeType,
				"tracker.users.audit.targetId":         auditEvent.TargetId,
				"tracker.users.audit.targetType":       auditEvent.TargetType,
			},
		})
	}
	err := flush.Flush(a.MetricForwarder, metrics)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuditEvent) flushLogs() {
	err := a.Logger.Flush()
	if err != nil {
		fmt.Println(AUDIT_EVENTS_LOGS_COULD_NOT_BE_FORWARDED, err.Error())
	}
}
