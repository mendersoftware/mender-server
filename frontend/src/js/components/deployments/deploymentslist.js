// Copyright 2019 Northern.tech AS
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
import React from 'react';

import Loader from '../common/loader';
import Pagination from '../common/pagination';
import DeploymentItem, {
  DeploymentDeviceCount,
  DeploymentDeviceGroup,
  DeploymentEndTime,
  DeploymentProgress,
  DeploymentRelease,
  DeploymentStartTime,
  deploymentTypeClasses
} from './deploymentitem';

export const defaultHeaders = [
  { title: 'Release', renderer: DeploymentRelease },
  { title: 'Target device(s)', renderer: DeploymentDeviceGroup },
  { title: 'Start time', renderer: DeploymentStartTime },
  { title: `End time`, renderer: DeploymentEndTime },
  { title: '# devices', class: 'align-right column-defined', renderer: DeploymentDeviceCount },
  { title: 'Status', renderer: DeploymentProgress }
];

const defaultRowsPerPage = 20;

export const DeploymentsList = ({
  abort,
  canDeploy,
  canConfigure,
  componentClass = '',
  count,
  devices,
  headers = defaultHeaders,
  idAttribute,
  isEnterprise,
  items,
  listClass = '',
  loading,
  openReport,
  onChangePage,
  onChangeRowsPerPage,
  page,
  pageSize,
  rootRef,
  showPagination,
  type
}) =>
  !!items.length && (
    <div className={`fadeIn deploy-table-contain ${componentClass}`} ref={rootRef}>
      <div className={`deployment-item deployment-header-item muted ${deploymentTypeClasses[type] || ''}`}>
        {headers.map((item, index) => (
          <div key={`${item.title}-${index}`} className={item.class || ''}>
            {item.title}
          </div>
        ))}
      </div>
      <div className={listClass}>
        {items.map(deployment => (
          <DeploymentItem
            abort={abort}
            canConfigure={canConfigure}
            canDeploy={canDeploy}
            columnHeaders={headers}
            deployment={deployment}
            devices={devices}
            key={`${type}-deployment-${deployment.created}`}
            idAttribute={idAttribute}
            isEnterprise={isEnterprise}
            openReport={openReport}
            type={type}
          />
        ))}
      </div>
      <div className="flexbox">
        {(count > items.length || items.length > defaultRowsPerPage || showPagination) && (
          <Pagination
            className="margin-top-none"
            count={count}
            rowsPerPage={pageSize}
            onChangeRowsPerPage={onChangeRowsPerPage}
            page={page}
            onChangePage={onChangePage}
          />
        )}
        <Loader show={loading} small />
      </div>
    </div>
  );

export default DeploymentsList;
