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
import { List, ListItem, ListItemText, Typography, darken, getOverlayAlpha, lighten, listClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import { Link } from '@northern.tech/common-ui/Link';
import MenderTooltip from '@northern.tech/common-ui/helptips/MenderTooltip';
import storeActions from '@northern.tech/store/actions';
import { TIMEOUTS } from '@northern.tech/store/constants';
import { getFeatures, getUserCapabilities, getVersionInformation } from '@northern.tech/store/selectors';
import { isDarkMode } from '@northern.tech/store/utils';
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
  licenseLink: { fontWeight: 'inherit' },
  list: {
    backgroundColor: isDarkMode(theme.palette.mode) ? lighten(theme.palette.background.paper, 0.08) : darken(theme.palette.background.paper, 0.08),
    position: 'relative',
    [`.${listClasses.root}`]: { paddingTop: 0 },
    '&::after': {
      content: '""',
      position: 'absolute',
      top: 0,
      right: 0,
      bottom: 0,
      width: 1,
      backgroundColor: theme.palette.divider,
      zIndex: 1
    }
  },
  navLink: {
    padding: theme.spacing(3.5),
    paddingLeft: theme.spacing(5),
    color: theme.palette.text.primary,
    borderTop: '1px solid transparent',
    borderBottom: '1px solid transparent',
    [`&:hover`]: {
      backgroundColor: isDarkMode(theme.palette.mode) ? lighten(theme.palette.background.default, getOverlayAlpha(1)) : theme.palette.background.default,
      color: theme.palette.text.primary
    },
    [`&.active`]: {
      backgroundColor: theme.palette.background.default,
      borderTop: `1px solid ${theme.palette.divider}`,
      borderBottom: `1px solid ${theme.palette.divider}`,
      position: 'relative',
      zIndex: 2
    }
  },
  lowerList: { gap: theme.spacing(), paddingLeft: theme.spacing(5), paddingRight: theme.spacing(2) },
  versions: { display: 'grid', gridTemplateColumns: 'max-content max-content', columnGap: theme.spacing() }
}));

const linkables = {
  'Integration': 'integration',
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

  useEffect(
    () => () => {
      clearTimeout(timer.current);
    },
    []
  );

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
                <Link href={`https://github.com/mendersoftware/${linkables[key]}/tree/${version}`} external>
                  {key}
                </Link>
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
    <div className={`leftFixed leftNav flexbox column space-between ${classes.list}`}>
      <List>
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
              <ListItemText primary={item.title} />
            </ListItem>
          );
          return accu;
        }, [])}
      </List>
      <List className={`flexbox column padding-bottom ${classes.lowerList}`}>
        <NavLink to={`/${routeConfigs.help.path}`}>
          <Typography variant="body2">{routeConfigs.help.title}</Typography>
        </NavLink>
        <VersionInfo />
        <Typography variant="body2">
          <DocsLink
            className={classes.licenseLink}
            path={`release-information/release-notes-changelog/${getDocsLocation({ isEnterprise, isHosted })}`}
            title="Release information"
          />
        </Typography>
        <Typography variant="body2">
          <DocsLink className={classes.licenseLink} path="release-information/supported-releases#mender-client" title="Supported Mender client releases" />
        </Typography>
        <Typography variant="body2">
          <DocsLink className={classes.licenseLink} path="release-information/open-source-licenses" title="License information" />
        </Typography>
      </List>
    </div>
  );
};

export default LeftNav;
