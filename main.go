package main

import (
	"os"
	"sync"

	"github.com/utr1903/newrelic-tracker-user/pkg/users"
)

func main() {
	organizationId := os.Getenv("NEWRELIC_ORGANIZATION_ID")
	wg := new(sync.WaitGroup)

	// Users
	wg.Add(1)
	go authUsers(wg, organizationId)

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
