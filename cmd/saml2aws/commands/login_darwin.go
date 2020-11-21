package commands

import (
	"github.com/daxingplay/saml2alicloud/v2/helper/credentials"
	"github.com/daxingplay/saml2alicloud/v2/helper/osxkeychain"
)

func init() {
	credentials.CurrentHelper = &osxkeychain.Osxkeychain{}
}
