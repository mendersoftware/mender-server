// Copyright 2022 Northern.tech AS
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
//@ts-nocheck
import { useCallback, useMemo } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

import {
  commonProcessor,
  formatAuditlogs,
  formatDeployments,
  formatDeviceSearch,
  formatPageState,
  formatReleases,
  formatTenants,
  generateDeploymentsPath,
  generateDevicePath,
  generateReleasesPath,
  generateTenantPath,
  parseAuditlogsQuery,
  parseDeploymentsQuery,
  parseDeviceQuery,
  parseReleasesQuery,
  parseTenantsQuery
} from './locationutils';

export const defaultProcessors = {
  auditlogs: {
    format: formatAuditlogs,
    locate: () => undefined,
    parse: parseAuditlogsQuery
  },
  common: {
    format: formatPageState,
    locate: () => undefined,
    parse: commonProcessor
  },
  deployments: {
    format: formatDeployments,
    locate: generateDeploymentsPath,
    parse: parseDeploymentsQuery
  },
  devices: {
    format: formatDeviceSearch,
    locate: generateDevicePath,
    parse: parseDeviceQuery
  },
  releases: {
    format: formatReleases,
    locate: generateReleasesPath,
    parse: parseReleasesQuery
  },
  tenants: {
    format: formatTenants,
    locate: generateTenantPath,
    parse: parseTenantsQuery
  }
};

export const useLocationParams = (key, extras, processors = defaultProcessors) => {
  let [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const location = useLocation();

  let value = useMemo(() => {
    const { pageState, params, sort } = processors.common.parse(searchParams);
    const extendedExtras = { ...extras, pageState, location };
    return {
      ...pageState,
      sort,
      ...processors[key].parse(params, extendedExtras)
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(extras), key, location.search, location.pathname, searchParams.toString()]);

  let setValue = useCallback(
    (newValue, options = {}) => {
      const pathname = processors[key].locate({ pageState: newValue.pageState, location });
      const searchQuery = [processors.common.format(newValue.pageState, extras), processors[key].format(newValue, extras)].filter(i => i).join('&');
      navigate({ pathname, replace: true, search: `?${searchQuery}`, ...options });
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(extras), key]
  );

  return [value, setValue];
};
