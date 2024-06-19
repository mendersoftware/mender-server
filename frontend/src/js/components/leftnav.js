// Copyright 2018 Northern.tech AS
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
import React, { useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link, NavLink } from 'react-router-dom';

// material ui
import { List, ListItem, ListItemText, Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import copy from 'copy-to-clipboard';

import { setSnackbar, setVersionInfo } from '../actions/appActions';
import { TIMEOUTS, canAccess } from '../constants/appConstants';
import { getFeatures, getUserCapabilities, getVersionInformation } from '../selectors';
import DocsLink from './common/docslink';

const listItems = [
  { route: '/', text: 'Dashboard', canAccess },
  { route: '/devices', text: 'Devices', canAccess: ({ userCapabilities: { canReadDevices } }) => canReadDevices },
  { route: '/releases', text: 'Releases', canAccess: ({ userCapabilities: { canReadReleases, canUploadReleases } }) => canReadReleases || canUploadReleases },
  { route: '/deployments', text: 'Deployments', canAccess: ({ userCapabilities: { canDeploy, canReadDeployments } }) => canReadDeployments || canDeploy },
  { route: '/auditlog', text: 'Audit log', canAccess: ({ userCapabilities: { canAuditlog } }) => canAuditlog }
];

const useStyles = makeStyles()(theme => ({
  licenseLink: { fontSize: '13px', position: 'relative', top: '6px', color: theme.palette.primary.main },
  infoList: { padding: 0, position: 'absolute', bottom: 30, left: 0, right: 0 },
  list: {
    backgroundColor: theme.palette.background.lightgrey,
    borderRight: `1px solid ${theme.palette.grey[300]}`
  },
  navLink: { padding: '22px 16px 22px 42px' },
  listItem: { padding: '16px 16px 16px 42px' },
  versions: { display: 'grid', gridTemplateColumns: 'max-content 60px', columnGap: theme.spacing(), '>a': { color: theme.palette.grey[100] } }
}));

const linkables = {
  'Integration': 'integration',
  'Mender-Client': 'mender',
  'Mender-Artifact': 'mender-artifact',
  'GUI': 'gui'
};

const VersionInfo = () => {
  const [clicks, setClicks] = useState(0);
  const timer = useRef();
  const { classes } = useStyles();

  const dispatch = useDispatch();
  const { isHosted } = useSelector(getFeatures);
  // eslint-disable-next-line no-unused-vars
  const { latestRelease, ...versionInformation } = useSelector(getVersionInformation);

  useEffect(() => {
    return () => {
      clearTimeout(timer.current);
    };
  }, []);

  const onVersionClick = () => {
    copy(JSON.stringify(versionInformation));
    dispatch(setSnackbar('Version information copied to clipboard'));
  };

  const versions = (
    <div className={classes.versions}>
      {Object.entries(versionInformation).reduce((accu, [key, version]) => {
        if (version) {
          accu.push(
            <React.Fragment key={key}>
              {linkables[key] ? (
                <a href={`https://github.com/mendersoftware/${linkables[key]}/tree/${version}`} target="_blank" rel="noopener noreferrer">
                  {key}
                </a>
              ) : (
                <div>{key}</div>
              )}
              <div className="align-right text-overflow" title={version}>
                {version}
              </div>
            </React.Fragment>
          );
        }
        return accu;
      }, [])}
    </div>
  );

  const onClick = () => {
    setClicks(clicks + 1);
    clearTimeout(timer.current);
    timer.current = setTimeout(() => {
      setClicks(0);
    }, TIMEOUTS.threeSeconds);
    if (clicks > 5) {
      dispatch(setVersionInfo({ Integration: 'next' }));
    }
    onVersionClick();
  };

  let title = versionInformation.Integration ? `Version: ${versionInformation.Integration}` : '';
  if (isHosted && versionInformation.Integration !== 'next') {
    title = 'Version: latest';
  }
  return (
    <Tooltip title={versions} placement="top">
      <div className="clickable slightly-smaller" onClick={onClick}>
        {title}
      </div>
    </Tooltip>
  );
};

export const LeftNav = () => {
  const releasesRef = useRef();
  const { classes } = useStyles();

  const userCapabilities = useSelector(getUserCapabilities);
  return (
    <div className={`leftFixed leftNav ${classes.list}`}>
      <List style={{ padding: 0 }}>
        {listItems.reduce((accu, item, index) => {
          if (!item.canAccess({ userCapabilities })) {
            return accu;
          }
          accu.push(
            <ListItem
              className={`navLink leftNav ${classes.navLink}`}
              component={NavLink}
              end={item.route === '/'}
              key={index}
              ref={item.route === '/releases' ? releasesRef : null}
              to={item.route}
            >
              <ListItemText primary={item.text} style={{ textTransform: 'uppercase' }} />
            </ListItem>
          );
          return accu;
        }, [])}
      </List>
      <List className={classes.infoList}>
        <ListItem className={`navLink leftNav ${classes.listItem}`} component={Link} to="/help">
          <ListItemText primary="Help & support" />
        </ListItem>
        <ListItem className={classes.listItem}>
          <ListItemText
            primary={<VersionInfo />}
            secondary={<DocsLink className={classes.licenseLink} path="release-information/open-source-licenses" title="License information" />}
          />
        </ListItem>
      </List>
    </div>
  );
};

export default LeftNav;
