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
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { getDebConfigurationCode } from './helpers';

const oldHostname = window.location.hostname;
const postTestCleanUp = () => {
  window.location = {
    ...window.location,
    hostname: oldHostname
  };
};

describe('getDebConfigurationCode function', () => {
  let code;
  describe('configuring devices for hosted mender', () => {
    beforeEach(() => {
      code = getDebConfigurationCode({
        ipAddress: '192.168.7.41',
        isTrial: true,
        deviceType: 'raspberrypi3',
        token: 'omnomnom'
      });
    });
    afterEach(postTestCleanUp);
    it('should not contain any template string leftovers', async () => {
      expect(code).not.toMatch(/\$\{([^}]+)\}/);
    });
    it('should return a sane result', async () => {
      expect(code).toMatch(
        `wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --demo --server-ip 192.168.7.41`
      );
    });
    it('should not contain tenant information for OS calls', async () => {
      expect(code).not.toMatch(/tenant/);
      expect(code).not.toMatch(/token/);
      expect(code).not.toMatch(/TENANT/);
      expect(code).not.toMatch(/TOKEN/);
    });
  });
  describe('configuring devices for hosted mender', () => {
    beforeEach(() => {
      window.location = {
        ...window.location,
        hostname: 'hosted.mender.io'
      };
    });
    afterEach(postTestCleanUp);

    it('should contain sane information for hosted calls', async () => {
      code = getDebConfigurationCode({
        deviceType: 'raspberrypi3',
        hasMonitor: true,
        isHosted: true,
        isOnboarding: true,
        tenantToken: 'token',
        token: 'omnomnom'
      });
      expect(code).toMatch(
        `JWT_TOKEN="omnomnom"
TENANT_TOKEN="token"
wget -O- https://get.mender.io | sudo bash -s -- --demo --commercial --jwt-token $JWT_TOKEN --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-url https://hosted.mender.io --server-cert=""`
      );
    });
    it('should contain sane information for hosted calls by users without monitor access', async () => {
      code = getDebConfigurationCode({
        deviceType: 'raspberrypi3',
        isHosted: true,
        isOnboarding: true,
        tenantToken: 'token',
        token: 'omnomnom'
      });
      expect(code).toMatch(
        `JWT_TOKEN="omnomnom"
TENANT_TOKEN="token"
wget -O- https://get.mender.io | sudo bash -s -- --demo --jwt-token $JWT_TOKEN --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-url https://hosted.mender.io --server-cert=""`
      );
    });
  });
  describe('configuring devices for staging.hosted.mender', () => {
    beforeEach(() => {
      window.location = {
        ...window.location,
        hostname: 'staging.hosted.mender.io'
      };
    });
    afterEach(postTestCleanUp);

    it('should contain sane information for staging preview calls', async () => {
      code = getDebConfigurationCode({
        deviceType: 'raspberrypi3',
        hasMonitor: true,
        isHosted: true,
        isOnboarding: true,
        isPreRelease: true,
        tenantToken: 'token',
        token: 'omnomnom'
      });
      expect(code).toMatch(
        `JWT_TOKEN="omnomnom"
TENANT_TOKEN="token"
wget -O- https://get.mender.io/staging | sudo bash -s -- --demo -c experimental --commercial --jwt-token $JWT_TOKEN --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-url https://staging.hosted.mender.io --server-cert=""`
      );
    });
  });
  describe('configuring devices for fancy.enterprise.on.prem', () => {
    beforeEach(() => {
      window.location = {
        ...window.location,
        hostname: 'fancy.enterprise.on.prem'
      };
    });
    afterEach(postTestCleanUp);

    it('should contain sane information for enterprise demo on-prem calls', async () => {
      code = getDebConfigurationCode({
        ipAddress: '1.2.3.4',
        isTrial: true,
        tenantToken: 'token',
        token: 'omnomnom',
        deviceType: 'raspberrypi3'
      });
      expect(code).toMatch(
        `TENANT_TOKEN="token"
wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-ip 1.2.3.4`
      );
    });
    it('should contain sane information for enterprise production on-prem calls', async () => {
      code = getDebConfigurationCode({
        ipAddress: '1.2.3.4',
        isTrial: false,
        tenantToken: 'token',
        deviceType: 'raspberrypi3'
      });
      expect(code).toMatch(
        `TENANT_TOKEN="token"
wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --retry-poll 300 --update-poll 1800 --inventory-poll 28800 --server-url https://fancy.enterprise.on.prem --server-cert=""`
      );
    });
  });
  describe('configuring devices for fancy.opensource.on.prem', () => {
    beforeEach(() => {
      window.location = {
        ...window.location,
        hostname: 'fancy.opensource.on.prem'
      };
    });
    afterEach(postTestCleanUp);

    it('should contain sane information for OS demo on-prem calls', async () => {
      code = getDebConfigurationCode({
        ipAddress: '1.2.3.4',
        isTrial: true,
        tenantToken: 'token',
        token: 'omnomnom',
        deviceType: 'raspberrypi3'
      });
      expect(code).toMatch(
        `wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --demo --server-ip 1.2.3.4`
      );
    });
    it('should contain sane information for OS production on-prem calls', async () => {
      code = getDebConfigurationCode({
        ipAddress: '1.2.3.4',
        isTrial: false,
        tenantToken: 'token',
        token: 'omnomnom',
        deviceType: 'raspberrypi3'
      });
      expect(code).toMatch(
        `wget -O- https://get.mender.io | sudo bash -s -- --demo --force-mender-client4 -- --quiet --device-type "raspberrypi3" --tenant-token $TENANT_TOKEN --retry-poll 300 --update-poll 1800 --inventory-poll 28800 --server-url https://fancy.opensource.on.prem --server-cert=""`
      );
    });
  });
});
