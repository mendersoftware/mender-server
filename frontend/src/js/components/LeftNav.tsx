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
import { NavLink } from 'react-router-dom';

// material ui
import { List, ListItem, ListItemText } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import MenderTooltip from '@northern.tech/common-ui/MenderTooltip';
import storeActions from '@northern.tech/store/actions';
import { TIMEOUTS } from '@northern.tech/store/constants';
import { getFeatures, getUserCapabilities, getVersionInformation } from '@northern.tech/store/selectors';
import copy from 'copy-to-clipboard';

import { routeConfigs } from '../config/routes';

const { setSnackbar, setVersionInformation } = storeActions;

const listItems = [
  { ...routeConfigs.dashboard, canAccess: ({ userCapabilities: { SPTenant } }) => !SPTenant },
  { ...routeConfigs.devices, canAccess: ({ userCapabilities: { canReadDevices, SPTenant } }) => canReadDevices && !SPTenant },
  {
    ...routeConfigs.releases,
    canAccess: ({ userCapabilities: { canReadReleases, canUploadReleases, SPTenant } }) => (canReadReleases || canUploadReleases) && !SPTenant
  },
  {
    ...routeConfigs.deployments,
    canAccess: ({ userCapabilities: { canDeploy, canReadDeployments, SPTenant } }) => (canReadDeployments || canDeploy) && !SPTenant
  },
  { ...routeConfigs.tenants, canAccess: ({ userCapabilities: { SPTenant } }) => SPTenant },
  { ...routeConfigs.auditlog, canAccess: ({ userCapabilities: { canAuditlog } }) => canAuditlog }
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
  versions: { display: 'grid', gridTemplateColumns: 'max-content max-content', columnGap: theme.spacing() }
}));

const linkables = {
  'Integration': 'integration',
  'Mender-Client': 'mender',
  'Mender-Artifact': 'mender-artifact',
  'Server': 'mender-server'
};

const VersionInfo = () => {
  const [clicks, setClicks] = useState(0);
  const timer = useRef();
  const { classes } = useStyles();

  const dispatch = useDispatch();
  const { isHosted } = useSelector(getFeatures);
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
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
    timer.current = setTimeout(() => setClicks(0), TIMEOUTS.threeSeconds);
    if (clicks > 5) {
      dispatch(setVersionInformation({ Integration: 'next' }));
    }
    onVersionClick();
  };

  let title = versionInformation.Integration ? `Version: ${versionInformation.Integration}` : '';
  if (isHosted && versionInformation.Integration !== 'next') {
    title = 'Version: latest';
  }
  return (
    <MenderTooltip arrow title={versions} placement="top">
      <div className="clickable slightly-smaller" onClick={onClick}>
        {title}
      </div>
    </MenderTooltip>
  );
};

const getDocsLocation = ({ isHosted, isEnterprise }) => {
  if (isHosted) {
    return 'hosted-mender';
  } else if (isEnterprise) {
    return 'mender-server-enterprise';
  }
  return 'mender-server';
};

export const LeftNav = () => {
  const releasesRef = useRef();
  const { classes } = useStyles();

  const { isEnterprise, isHosted } = useSelector(getFeatures); // here we have to only rely on the enterprise flag, not the tenant setting, to also point hosted enterprise users to the right location
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
              end={item.path === ''}
              key={index}
              ref={item.path === routeConfigs.releases.path ? releasesRef : null}
              to={`/${item.path}`}
            >
              <ListItemText primary={item.title} style={{ textTransform: 'uppercase' }} />
            </ListItem>
          );
          return accu;
        }, [])}
      </List>
      <List className={classes.infoList}>
        <ListItem className={`navLink leftNav ${classes.listItem}`} component={NavLink} to={`/${routeConfigs.help.path}`}>
          <ListItemText primary={routeConfigs.help.title} />
        </ListItem>
        <ListItem className={classes.listItem}>
          <ListItemText
            primary={<VersionInfo />}
            secondary={
              <>
                <DocsLink
                  className={classes.licenseLink}
                  path={`release-information/release-notes-changelog/${getDocsLocation({ isEnterprise, isHosted })}`}
                  title="Release information"
                />
                <br />
                <DocsLink className={classes.licenseLink} path="release-information/open-source-licenses" title="License information" />
              </>
            }
            slotProps={{ secondary: { lineHeight: '2em' } }}
          />
        </ListItem>
      </List>
    </div>
  );
};

export default LeftNav;
