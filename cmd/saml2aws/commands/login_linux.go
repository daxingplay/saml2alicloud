package commands

import (
	"github.com/daxingplay/saml2alicloud/v2/helper/credentials"
	"github.com/daxingplay/saml2alicloud/v2/helper/linuxkeyring"
)

func init() {
	if keyringHelper, err := linuxkeyring.NewKeyringHelper(); err == nil {
		credentials.CurrentHelper = keyringHelper
	}
}
