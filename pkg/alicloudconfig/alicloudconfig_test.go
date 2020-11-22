package alicloudconfig

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestUpdateSamlConfig(t *testing.T) {
	os.Remove(".credentials")

	logrus.SetLevel(logrus.DebugLevel)

	sharedCreds := &CredentialsProvider{".credentials", "saml"}

	exist, err := sharedCreds.CredsExists()
	assert.Nil(t, err)
	assert.True(t, exist)

	awsCreds := &AliCloudCredentials{
		AliCloudAccessKey:     "testid",
		AliCloudSecretKey:     "testsecret",
		AliCloudSessionToken:  "testtoken",
		AliCloudSecurityToken: "testtoken",
	}

	err = sharedCreds.Save(awsCreds)
	assert.Nil(t, err)

	profile, err := sharedCreds.Load()
	assert.Nil(t, err)
	assert.Equal(t, "testid", profile.AccessKeyId)
	assert.Equal(t, "testsecret", profile.AccessKeySecret)
	assert.Equal(t, "testtoken", profile.StsToken)

	os.Remove(".credentials")
}
