package service

import (
	"encoding/csv"
	"fmt"
	"os"
)


// CredentialsChecker keeps track and validates credentials.
type CredentialsChecker struct {
	creds map[string]string
}

// NewCredentialsChecker returns a new CredentialsChecker.
func NewCredentialsChecker() CredentialsChecker {
	return CredentialsChecker{creds: map[string]string{}}
}

// Check validates credentials.
func (c *CredentialsChecker) Check(form interface{}) bool {
	m := form.(map[string]interface{})
	username := m["username"].(string)
	password := m["password"].(string)
	pass, ok := c.creds[username]
	return ok && pass == password
}

// AddCreds adds username/password pairs to credentials.
func (c *CredentialsChecker) AddCreds(creds map[string]string) {
	for user, pass := range creds {
		c.creds[user] = pass
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
	for i, row := range rows {
		if len(row) != 2 {
			return fmt.Errorf("invalid length on row %d", i+1)
		}
		creds[row[0]] = row[1]
	}
	c.AddCreds(creds)
	return nil
}
