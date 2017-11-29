package service

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
)

// CredentialsChecker keeps track and validates credentials.
type CredentialsChecker struct {
	creds  map[string]string
	groups map[string][]string
}

// NewCredentialsChecker returns a new CredentialsChecker.
func NewCredentialsChecker() CredentialsChecker {
	return CredentialsChecker{
		creds:  map[string]string{},
		groups: map[string][]string{},
	}
}

// Check validates credentials.
func (c *CredentialsChecker) Check(form interface{}) bool {
	m := form.(map[string]interface{})
	username := m["username"].(string)
	password := m["password"].(string)
	pass, ok := c.creds[username]
	return ok && pass == password
}

// UserInGroup checks if a user belongs to the specified group.
func (c *CredentialsChecker) UserInGroup(user, group string) bool {
	userGroups, ok := c.groups[user]
	if !ok {
		return false
	}

	pos := sort.SearchStrings(userGroups, group)
	return pos == len(userGroups) || userGroups[pos] == group
}

// AddCreds adds username/password pairs to credentials.
func (c *CredentialsChecker) AddCreds(creds map[string]string) {
	for user, pass := range creds {
		c.creds[user] = pass
	}
}

// AddGroups adds username/groups mapping.
func (c *CredentialsChecker) AddGroups(groups map[string][]string) {
	for user, userGroups := range groups {
		c.groups[user] = userGroups
	}
}

// LoadCreds loads credentials from a CSV file.
func (c *CredentialsChecker) LoadCreds(csvFile string) error {
	f, err := os.Open(csvFile)
	if err != nil {
		return err
	}

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}

	creds := map[string]string{}
	groups := map[string][]string{}
	for i, row := range rows {
		if len(row) < 2 {
			return fmt.Errorf("invalid length on row %d", i+1)
		}
		creds[row[0]] = row[1]
		if len(row) > 2 && row[2] != "" {
			groupList := strings.Split(row[2], " ")
			sort.Strings(groupList)
			groups[row[0]] = groupList
		}
	}
	c.AddCreds(creds)
	c.AddGroups(groups)
	return nil
}
