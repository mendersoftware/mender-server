// Copyright 2025 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

interface DebConfigurationProps {
  deviceType?: string;
  hasMonitor?: boolean;
  ipAddress?: string;
  isHosted?: boolean;
  isPreRelease?: boolean;
  isTrial?: boolean;
  tenantToken?: string;
  token: string;
}

const getInstallScriptArgs = ({ isHosted, isPreRelease, hasMonitor }: Partial<DebConfigurationProps>) => {
  let installScriptArgs = '--demo';
  installScriptArgs = isPreRelease ? `${installScriptArgs} -c experimental` : installScriptArgs;
  installScriptArgs = isHosted && hasMonitor ? `${installScriptArgs} --commercial` : installScriptArgs;
  installScriptArgs = isHosted ? `${installScriptArgs} --jwt-token $JWT_TOKEN` : installScriptArgs;
  return installScriptArgs;
};

const getSetupArgs = ({ deviceType = 'generic-armv6', ipAddress, tenantToken, isTrial }: Partial<DebConfigurationProps>) => {
  let menderSetupArgs = `--quiet --device-type "${deviceType}"`;
  menderSetupArgs = tenantToken ? `${menderSetupArgs} --tenant-token $TENANT_TOKEN` : menderSetupArgs;
  // in production we use polling intervals from the client examples: https://github.com/mendersoftware/mender/blob/master/examples/mender.conf.production
  menderSetupArgs = isTrial ? `${menderSetupArgs} --demo` : `${menderSetupArgs} --retry-poll 300 --update-poll 1800 --inventory-poll 28800`;
  menderSetupArgs = ipAddress
    ? `${menderSetupArgs} --server-ip ${ipAddress}` // we still need to forward the ipAddress when showing the snippets for a gateway setup
    : `${menderSetupArgs} --server-url https://${window.location.hostname} --server-cert=""`;
  return menderSetupArgs;
};

const installComponents = '--force-mender-client4';

export const getDebConfigurationCode = (props: DebConfigurationProps) => {
  const { tenantToken, token, isPreRelease } = props;
  const envVars = tenantToken ? `JWT_TOKEN="${token}"\nTENANT_TOKEN="${tenantToken}"\n` : '';
  const installScriptArgs = getInstallScriptArgs(props);
  const scriptUrl = isPreRelease ? 'https://get.mender.io/staging' : 'https://get.mender.io';
  const menderSetupArgs = getSetupArgs(props);
  return `${envVars}wget -O- ${scriptUrl} | sudo bash -s -- ${installScriptArgs} ${installComponents} -- ${menderSetupArgs}`;
};
