package cauth

import "fmt"

const (
	APITenants = "/api/v2/tenants"
	APITeams   = "/api/v2/teams"
	APIRoles   = "/api/v2/roles"
	APIUser    = "/api/v2/users/%s"
)

func TenantsPath(user string) string {
	return fmt.Sprintf("%s?user=%s", APITenants, user)
}

func TeamsPath(user string) string {
	return fmt.Sprintf("%s?user=%s", APITeams, user)
}

func RolesPath(subType, subId string) string {
	return fmt.Sprintf("%s?subType=%s&subId=%s", APIRoles, subType, subId)
}

func UserPath(name string) string {
	return fmt.Sprintf(APIUser, name)
}
