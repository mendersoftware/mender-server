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
import type { ComponentType, RefObject } from 'react';

import { Typography, useMediaQuery } from '@mui/material';
import { useTheme } from '@mui/material/styles';
import { makeStyles } from 'tss-react/mui';

import Loader from '@northern.tech/common-ui/Loader';
import Pagination from '@northern.tech/common-ui/Pagination';
import { DEPLOYMENT_STATES } from '@northern.tech/store/constants';
import { Deployment } from '@northern.tech/store/deploymentsSlice';

import DeploymentItem, {
  DeploymentDeviceCount,
  DeploymentDeviceGroup,
  DeploymentEndTime,
  DeploymentItemProps,
  DeploymentRelease,
  DeploymentStartTime
} from './DeploymentItem';
import { DeploymentProgress } from './progress/DeploymentProgress';

export interface ColumnHeader {
  class?: string;
  props?: Record<string, unknown>;
  renderer: ComponentType<any>;
  title: string;
}

export const defaultHeaders: ColumnHeader[] = [
  { title: 'Release', renderer: DeploymentRelease },
  { title: 'Target device(s)', class: 'text-overflow', renderer: DeploymentDeviceGroup },
  { title: 'Start time', class: 'align-right', renderer: DeploymentStartTime },
  { title: `End time`, class: 'align-right', renderer: DeploymentEndTime },
  { title: '# devices', class: 'align-right', renderer: DeploymentDeviceCount },
  { title: 'Status', renderer: DeploymentProgress, props: { variant: 'list' } }
];

const defaultRowsPerPage = 20;

const deploymentTypeCommonColumns = '2fr 2fr 1.5fr 1.5fr 1fr';

const useStyles = makeStyles()(theme => ({
  row: {
    alignItems: 'center',
    borderBottom: `1px solid ${theme.palette.divider}`,
    ':hover': {
      backgroundColor: theme.palette.action.hover
    }
  },
  rowHeader: {
    fontWeight: theme.typography.fontWeightBold,
    ':hover': {
      backgroundColor: 'initial'
    }
  },
  rowState: {
    display: 'grid',
    gridColumnGap: theme.spacing(3),
    minWidth: 1300,
    [`&.${DEPLOYMENT_STATES.inprogress}-item`]: {
      gridTemplateColumns: `${deploymentTypeCommonColumns} 4fr 1.5fr ${theme.spacing(6)}`
    },
    [`&.${DEPLOYMENT_STATES.pending}-item`]: {
      gridTemplateColumns: `${deploymentTypeCommonColumns} 2fr 1.5fr ${theme.spacing(6)}`
    },
    [`&.${DEPLOYMENT_STATES.scheduled}-item`]: {
      gridTemplateColumns: `${deploymentTypeCommonColumns} 1fr 1.5fr ${theme.spacing(6)}`
    },
    [`&.${DEPLOYMENT_STATES.finished}-item`]: {
      gridTemplateColumns: `${deploymentTypeCommonColumns} 2.25fr 1fr 1.25fr`
    }
  }
}));

interface DeploymentsListProps extends DeploymentItemProps {
  count: number;
  headers?: ColumnHeader[];
  items: Deployment[];
  loading?: boolean;
  onChangePage?: (page: number) => void;
  onChangeRowsPerPage?: (perPage: number) => void;
  openReport: (type: string, id: string) => void;
  page?: number;
  pageSize?: number;
  rootRef?: RefObject<HTMLElement>;
  showPagination?: boolean;
}

export const DeploymentsList = ({
  abort,
  canDeploy,
  canConfigure,
  count,
  devices,
  headers = defaultHeaders,
  idAttribute,
  isEnterprise,
  items,
  loading,
  openReport,
  onChangePage,
  onChangeRowsPerPage,
  page,
  pageSize,
  rootRef,
  showPagination,
  type
}: DeploymentsListProps) => {
  const { classes } = useStyles();
  const theme = useTheme();
  const isCompact = useMediaQuery(theme.breakpoints.down(1600));

  if (!items.length) {
    return null;
  }

  return (
    <div className="fadeIn" ref={rootRef}>
      {!isCompact && (
        <div className={`padding-small ${classes.row} ${classes.rowState} ${classes.rowHeader} ${type}-item`}>
          {headers.map((item, index) => (
            <Typography key={`${item.title}-${index}`} className={item.class || ''}>
              {item.title}
            </Typography>
          ))}
        </div>
      )}
      {items.map(deployment => (
        <DeploymentItem
          abort={abort}
          canConfigure={canConfigure}
          canDeploy={canDeploy}
          className={isCompact ? classes.row : `${classes.row} ${classes.rowState} ${type}-item`}
          columnHeaders={headers}
          deployment={deployment}
          devices={devices}
          key={deployment.id}
          idAttribute={idAttribute}
          isCompact={isCompact}
          isEnterprise={isEnterprise}
          openReport={openReport}
          type={type}
        />
      ))}
      <div className="flexbox">
        {(count > items.length || items.length > defaultRowsPerPage || showPagination) && (
          <Pagination
            classes={{ toolbar: 'padding-left-none', spacer: 'flexbox no-basis' }}
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
};

export default DeploymentsList;
