// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package config

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	//translate to env vars: CREATE_ARTIFACT_<CAPITALIZED>
	CfgSkipVerify     = "skipverify"
	CfgVerbose        = "verbose"
	CfgWorkDir        = "workdir"
	CfgDeploymentsUrl = "deployments_url"
)

func Init() {
	viper.SetEnvPrefix("CREATE_ARTIFACT")
	viper.AutomaticEnv()

	viper.SetDefault(CfgSkipVerify, false)
	viper.SetDefault(CfgVerbose, false)
	viper.SetDefault(CfgWorkDir, "/var")
	viper.SetDefault(CfgDeploymentsUrl, "http://mender-deployments:8080")
}

func ValidUrl(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	if u.Host == "" || u.Scheme == "" {
		return errors.New("url needs scheme and host at a minimum")
	}

	return nil
}

func ValidAbsPath(s string) error {
	if !filepath.IsAbs(s) {
		return errors.New("need an absolute path")
	}

	return nil
}

func Dump() string {
	return dump(CfgSkipVerify) +
		dump(CfgVerbose) +
		dump(CfgWorkDir) +
		dump(CfgDeploymentsUrl) +
		dump(CfgDeploymentsUrl)
}

func dump(n string) string {
	return fmt.Sprintf("%s: %v\n", n, viper.Get(n))
}
