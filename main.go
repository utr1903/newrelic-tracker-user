package main

import (
	"os"
	"strconv"
	"sync"

	"github.com/utr1903/newrelic-tracker-user/pkg/audit"
	"github.com/utr1903/newrelic-tracker-user/pkg/users"
)

func main() {
	organizationId := os.Getenv("NEWRELIC_ORGANIZATION_ID")
	accountId, _ := strconv.ParseInt(os.Getenv("NEWRELIC_ACCOUNT_ID"), 10, 64)
	wg := new(sync.WaitGroup)

	// Users
	wg.Add(1)
	go authUsers(wg, organizationId)

	// Audit
	wg.Add(1)
	go auditEvents(wg, organizationId, accountId)

	wg.Wait()
}

func authUsers(
	wg *sync.WaitGroup,
	organizationId string,
) {
	defer wg.Done()
	us := users.NewUsers(organizationId)
	us.Run()
}

func auditEvents(
	wg *sync.WaitGroup,
	organizationId string,
	accountId int64,
) {
	defer wg.Done()
	us := audit.NewAuditEvents(organizationId, accountId)
	us.Run()
}
