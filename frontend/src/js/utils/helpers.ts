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
  const installScriptArgs = ['--demo'];
  if (isPreRelease) installScriptArgs.push('-c experimental');
  if (isHosted && hasMonitor) installScriptArgs.push('--commercial');
  if (isHosted) installScriptArgs.push('--jwt-token $JWT_TOKEN');
  return installScriptArgs.join(' ');
};

const getSetupArgs = ({ deviceType = 'generic-armv6', ipAddress, tenantToken, isTrial }: Partial<DebConfigurationProps>) => {
  const menderSetupArgs = ['--quiet', `--device-type "${deviceType}"`];
  if (tenantToken) menderSetupArgs.push('--tenant-token $TENANT_TOKEN');
  // in production we use polling intervals from the client examples: https://github.com/mendersoftware/mender/blob/master/examples/mender.conf.production
  menderSetupArgs.push(isTrial ? '--demo' : '--retry-poll 300 --update-poll 1800 --inventory-poll 28800');
  // we still need to forward the ipAddress when showing the snippets for a gateway setup
  menderSetupArgs.push(ipAddress ? `--server-ip ${ipAddress}` : `--server-url https://${window.location.hostname} --server-cert=""`);
  return menderSetupArgs.join(' ');
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
