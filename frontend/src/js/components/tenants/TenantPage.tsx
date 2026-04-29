// Copyright 2024 Northern.tech AS
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
import { useCallback, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';

import { Button, Paper, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { Link } from '@northern.tech/common-ui/Link';
import { getSpLimits, getTenantListWithLimits } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getTenants } from '@northern.tech/store/thunks';
import { isEmpty, toggle } from '@northern.tech/utils/helpers';

import { DeviceLimit } from '../header/DeviceNotifications';
import { TenantCreateForm } from './TenantCreateForm';
import { TenantList } from './TenantList';

interface TenantsEmptyStateProps {
  openModal: () => void;
}
const TenantsEmptyState = (props: TenantsEmptyStateProps) => {
  const { openModal } = props;
  const dispatch = useAppDispatch();
  useEffect(() => {
    dispatch(getTenants());
  }, [dispatch]);
  return (
    <div className="dashboard-placeholder">
      <p>You are not currently managing any tenants. </p>
      <p>
        <Link onClick={openModal}>Add a tenant</Link> to get started.
      </p>
    </div>
  );
};

const useStyles = makeStyles()(theme => ({
  limit: {
    maxWidth: '726px'
  },
  link: {
    color: theme.palette.secondary.main
  }
}));

export const TenantPage = () => {
  const [showCreate, setShowCreate] = useState<boolean>(false);
  const { classes } = useStyles();

  const { tenants } = useSelector(getTenantListWithLimits);
  const spLimits = useSelector(getSpLimits);

  const onToggleCreation = useCallback(() => setShowCreate(toggle), []);
  return (
    <div className="padding-right">
      <Typography variant="h5">Tenant management</Typography>
      <Typography variant="subtitle1" className="margin-top-small">
        Device allocation
      </Typography>
      <div className="full-width flexbox">
        {Object.values(spLimits).map(limit => (
          <div key={limit.id} className={`full-width margin-right ${classes.limit}`}>
            {limit.limit === -1 ? (
              <Paper variant="outlined" className="padding-small">
                <div className="flexbox space-between">
                  <Typography variant="subtitle2" className="capitalized-start">
                    {limit.name}
                  </Typography>
                  <Typography variant="body2">{limit.current}</Typography>
                </div>
              </Paper>
            ) : (
              <DeviceLimit serviceProvider total={limit.current} limit={limit.limit} type={limit.name} padded />
            )}
          </div>
        ))}
      </div>
      <Button className="margin-top-small margin-bottom-medium" color="secondary" component={RouterLink} to="/subscription" variant="text">
        Request changes to device limits
      </Button>

      <div className="flexbox full-width space-between">
        <Typography variant="subtitle1" className="margin-top-small">
          Tenants
        </Typography>
        <Button className="margin-top-small" variant="contained" onClick={onToggleCreation} disabled={isEmpty(spLimits)}>
          Create a tenant
        </Button>
      </div>
      {tenants.length ? <TenantList /> : <TenantsEmptyState openModal={onToggleCreation} />}
      {showCreate && <TenantCreateForm open={showCreate} onCloseClick={onToggleCreation} />}
    </div>
  );
};
