// Package credentials defines a Checker to hold and validate credentials.
package credentials

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Checker keeps track and validates credentials.
type Checker struct {
	creds  map[string]string
	groups map[string][]string
}

// NewChecker returns a new Checker.
func NewChecker() Checker {
	return Checker{
		creds:  map[string]string{},
		groups: map[string][]string{},
	}
}

// Check validates credentials.
func (c *Checker) Check(form interface{}) bool {
	m := form.(map[string]interface{})
	username := m["username"].(string)
	password := m["password"].(string)
	pass, ok := c.creds[username]
	return ok && pass == password
}

// UserInGroups checks if a user belongs to any of the specified groups.
func (c *Checker) UserInGroups(user string, groups []string) bool {
	userGroups, ok := c.groups[user]
	if !ok {
		return false
	}

	for _, group := range groups {
		pos := sort.SearchStrings(userGroups, group)
		if pos == len(userGroups) || userGroups[pos] == group {
			return true
		}
	}
	return false
}

// AddCreds adds username/password pairs to credentials.
func (c *Checker) AddCreds(creds map[string]string) {
	for user, pass := range creds {
		c.creds[user] = pass
	}
}

// AddGroups adds username/groups mapping.
func (c *Checker) AddGroups(groups map[string][]string) {
	for user, userGroups := range groups {
		c.groups[user] = userGroups
	}
}

// LoadCreds loads credentials from a CSV file.
func (c *Checker) LoadCreds(csvFile string) error {
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
