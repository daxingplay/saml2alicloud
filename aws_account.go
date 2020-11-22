package saml2aws

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// AWSAccount holds the AWS account name and roles
type AWSAccount struct {
	Name  string
	Roles []*AWSRole
}

// ParseAWSAccounts extract the aws accounts from the saml assertion
func ParseAWSAccounts(audience string, samlAssertion string) ([]*AWSAccount, error) {
	res, err := http.PostForm(audience, url.Values{"SAMLResponse": {samlAssertion}})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving AlibabaCloud login form")
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving AlibabaCloud login body")
	}

	return ExtractAWSAccounts(data)
}

// ExtractAWSAccounts extract the accounts from the AWS html page
func ExtractAWSAccounts(data []byte) ([]*AWSAccount, error) {
	accounts := []*AWSAccount{}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "failed to build document from response")
	}

	doc.Find("#samlRoleForm > div.form-group > div.col-sm-4 > label").Each(func(i int, s *goquery.Selection) {
		account := new(AWSAccount)
		name := s.Text()
		parts := strings.Split(name, ":")
		account.Name = strings.TrimSpace(parts[1])
		s.Parent().Parent().Next().Find("input[name='roleAttribute']").Each(func(i int, s *goquery.Selection) {
			role := new(AWSRole)
			label := s.Parent()
			role.Name = strings.TrimSpace(label.Text())
			value, _ := s.Attr("value")
			parts = strings.Split(value, ",")
			role.RoleARN = strings.TrimSpace(parts[0])
			role.PrincipalARN = strings.TrimSpace(parts[1])
			account.Roles = append(account.Roles, role)
		})
		accounts = append(accounts, account)
	})

	return accounts, nil
}

// AssignPrincipals assign principal from roles
func AssignPrincipals(awsRoles []*AWSRole, awsAccounts []*AWSAccount) {

	awsPrincipalARNs := make(map[string]string)
	for _, awsRole := range awsRoles {
		awsPrincipalARNs[awsRole.RoleARN] = awsRole.PrincipalARN
	}

	for _, awsAccount := range awsAccounts {
		for _, awsRole := range awsAccount.Roles {
			awsRole.PrincipalARN = awsPrincipalARNs[awsRole.RoleARN]
		}
	}

}

// LocateRole locate role by name
func LocateRole(awsRoles []*AWSRole, roleName string) (*AWSRole, error) {
	for _, awsRole := range awsRoles {
		if awsRole.RoleARN == roleName {
			return awsRole, nil
		}
	}

	return nil, fmt.Errorf("Supplied RoleArn not found in saml assertion: %s", roleName)
}
