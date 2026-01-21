// Copyright 2015 Northern.tech AS
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
import { useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router-dom';

import { Alert } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Loader from '@northern.tech/common-ui/Loader';
import { DEPLOYMENT_ROUTES } from '@northern.tech/store/constants';
import { getCurrentUser, getShowSecurityAlert } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { saveUserSettings } from '@northern.tech/store/thunks';
import dayjs from 'dayjs';

import Deployments from './Deployments';
import Devices from './Devices';
import SoftwareDistribution from './SoftwareDistribution';

const useStyles = makeStyles()(theme => ({
  board: {
    columnGap: theme.spacing(6),
    display: 'flex',
    flexWrap: 'wrap',
    marginBottom: theme.spacing(6)
  },
  left: {
    flexGrow: 1,
    flexBasis: 0,
    minWidth: '60vw',
    display: 'flex',
    position: 'relative',
    rowGap: theme.spacing(6),
    flexDirection: 'column',
    [theme.breakpoints.up('xl')]: { minWidth: '50vw' }
  },
  right: {
    flexGrow: 1,
    minWidth: 400,
    border: 'none',
    paddingLeft: 0,
    paddingTop: 0,
    '.deployments .dashboard > h4': { marginTop: theme.spacing(6) },
    '.deployments .dashboard > h4.margin-top-none': { marginTop: 0 },
    [theme.breakpoints.up('xl')]: {
      borderLeft: `1px solid ${theme.palette.grey[500]}`,
      marginTop: theme.spacing(-2),
      paddingLeft: theme.spacing(6),
      paddingTop: theme.spacing(2),
      '.deployments .dashboard > h4': { marginTop: 0 }
    }
  },
  row: { flexWrap: 'wrap', maxWidth: '85vw' }
}));

export const Dashboard = () => {
  const { classes } = useStyles();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const { id: currentUser } = useSelector(getCurrentUser);
  const showSecurityAlert = useSelector(getShowSecurityAlert);
  const handleClick = params => {
    let redirect = params.route;
    if (params.route === 'deployments') {
      let query = params.open ? ['open=true'] : [];
      query = params.id ? [...query, `id=${params.id}`] : query;
      redirect = `/deployments/${params.tab || DEPLOYMENT_ROUTES.active.key}?${query.join('&')}`;
    }
    navigate(redirect, { state: { internal: true } });
  };
  const dismissAlert = () => {
    dispatch(saveUserSettings({ securityAlertDismissedTimestamp: dayjs().format() }));
  };

  return (
    <>
      {showSecurityAlert && (
        <Alert className="margin-bottom-small margin-right" onClose={dismissAlert} severity="warning">
          <>
            Your Mender account is not using recommended security settings. Improve your account security in{' '}
            <Link to="/settings/my-profile" color="inherit">
              My profile
            </Link>
            .
          </>
        </Alert>
      )}
      <h4 className="margin-left-small">Dashboard</h4>
      {currentUser ? (
        <div className={classes.board}>
          <div className={classes.left}>
            <Devices clickHandle={handleClick} />
            <SoftwareDistribution />
          </div>
          <Deployments className={classes.right} clickHandle={handleClick} />
        </div>
      ) : (
        <div className="flexbox centered" style={{ height: '75%' }}>
          <Loader show={true} />
        </div>
      )}
    </>
  );
};

export default Dashboard;
