package commands

import (
	"github.com/daxingplay/saml2alicloud/v2/helper/credentials"
	"github.com/daxingplay/saml2alicloud/v2/helper/wincred"
)

func init() {
	credentials.CurrentHelper = &wincred.Wincred{}
}
