/*

Copyright 2020 The Vouch Proxy Authors.
Use of this source code is governed by The MIT License (MIT) that
can be found in the LICENSE file. Software distributed under The
MIT License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
OR CONDITIONS OF ANY KIND, either express or implied.

*/

package nextcloud

import (
	"encoding/json"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"

	"github.com/vouch/vouch-proxy/pkg/cfg"
	"github.com/vouch/vouch-proxy/pkg/providers/common"
	"github.com/vouch/vouch-proxy/pkg/structs"
	"go.uber.org/zap"
)

// Provider provider specific functions
type Provider struct{}

var log *zap.SugaredLogger
var config *cfg.OauthConfig

// Configure see main.go configure()
func (Provider) Configure(configp *cfg.OauthConfig) {
	log = cfg.Logging.Logger
	config = configp
}

// GetUserInfo provider specific call to get userinfomation
func (Provider) GetUserInfo(r *http.Request, user *structs.User, customClaims *structs.CustomClaims, ptokens *structs.PTokens, opts ...oauth2.AuthCodeOption) (rerr error) {
	client, _, err := common.PrepareTokensAndClient(r, ptokens, true)
	if err != nil {
		return err
	}
	userinfo, err := client.Get(config.UserInfoURL)
	if err != nil {
		return err
	}
	defer func() {
		if err := userinfo.Body.Close(); err != nil {
			rerr = err
		}
	}()
	data, _ := ioutil.ReadAll(userinfo.Body)
	log.Infof("Ocs userinfo body: %s", string(data))
	if err = common.MapClaims(data, customClaims); err != nil {
		log.Error(err)
		return err
	}
	ncUser := structs.NextcloudUser{}
	if err = json.Unmarshal(data, &ncUser); err != nil {
		log.Error(err)
		return err
	}
	ncUser.PrepareUserData()
	user.Username = ncUser.Username
	user.Email = ncUser.Email
	return nil
}
