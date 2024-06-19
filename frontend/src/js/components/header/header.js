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
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { AccountCircle as AccountCircleIcon, ExitToApp as ExitIcon, ExpandMore } from '@mui/icons-material';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Button,
  Divider,
  IconButton,
  ListItemSecondaryAction,
  ListItemText,
  Menu,
  MenuItem,
  Toolbar,
  Typography,
  accordionClasses,
  accordionSummaryClasses,
  listItemTextClasses,
  menuItemClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import moment from 'moment';
import Cookies from 'universal-cookie';

import enterpriseLogo from '../../../assets/img/headerlogo-enterprise.png';
import logo from '../../../assets/img/headerlogo.png';
import whiteEnterpriseLogo from '../../../assets/img/whiteheaderlogo-enterprise.png';
import whiteLogo from '../../../assets/img/whiteheaderlogo.png';
import { setFirstLoginAfterSignup, setSearchState } from '../../actions/appActions';
import { getAllDeviceCounts } from '../../actions/deviceActions';
import { initializeSelf, logoutUser, setAllTooltipsReadState, setHideAnnouncement, switchUserOrganization } from '../../actions/userActions';
import { TIMEOUTS } from '../../constants/appConstants';
import { READ_STATES } from '../../constants/userConstants';
import { isDarkMode, toggle } from '../../helpers';
import {
  getAcceptedDevices,
  getCurrentSession,
  getCurrentUser,
  getDeviceCountsByStatus,
  getDeviceLimit,
  getFeatures,
  getIsEnterprise,
  getOrganization,
  getShowHelptips,
  getUserSettings
} from '../../selectors';
import Tracking from '../../tracking';
import { useDebounce } from '../../utils/debouncehook';
import Search from '../common/search';
import Announcement from './announcement';
import DemoNotification from './demonotification';
import DeploymentNotifications from './deploymentnotifications';
import DeviceNotifications from './devicenotifications';
import OfferHeader from './offerheader';
import TrialNotification from './trialnotification';

// Change this when a new feature/offer is introduced
const currentOffer = {
  name: 'add-ons',
  expires: '2021-12-30',
  trial: true,
  os: true,
  professional: true,
  enterprise: true
};

const cookies = new Cookies();

const useStyles = makeStyles()(theme => ({
  accordion: {
    ul: { paddingInlineStart: 0 },
    [`&.${accordionClasses.disabled}, &.${accordionClasses.expanded}`]: {
      backgroundColor: theme.palette.background.paper
    },
    [`.${accordionSummaryClasses.root}:hover`]: {
      backgroundColor: theme.palette.grey[400],
      color: theme.palette.text.link
    },
    [`.${menuItemClasses.root}:hover`]: {
      color: theme.palette.text.link
    }
  },
  header: {
    minHeight: 'unset',
    paddingLeft: theme.spacing(4),
    paddingRight: theme.spacing(5),
    width: '100%',
    borderBottom: `1px solid ${theme.palette.grey[100]}`,
    display: 'grid'
  },
  banner: { gridTemplateRows: `1fr ${theme.mixins.toolbar.minHeight}px` },
  buttonColor: { color: theme.palette.grey[600] },
  dropDown: {
    height: '100%',
    textTransform: 'none',
    [`.${menuItemClasses.root}:hover, .${listItemTextClasses.root}:hover`]: {
      color: theme.palette.text.link
    }
  },
  exitIcon: { color: theme.palette.grey[600], fill: theme.palette.grey[600] },
  demoTrialAnnouncement: {
    fontSize: 14,
    height: 'auto'
  },
  demoAnnouncementIcon: {
    height: 16,
    color: theme.palette.primary.main,
    '&.MuiButton-textPrimary': {
      color: theme.palette.primary.main,
      height: 'inherit'
    }
  },
  organization: { marginBottom: theme.spacing() },
  redAnnouncementIcon: {
    color: theme.palette.error.dark
  }
}));

const AccountMenu = () => {
  const [anchorEl, setAnchorEl] = useState(null);
  const [tenantSwitcherShowing, setTenantSwitcherShowing] = useState(false);
  const showHelptips = useSelector(getShowHelptips);
  const { email, tenants = [] } = useSelector(getCurrentUser);
  const { name } = useSelector(getOrganization);
  const isEnterprise = useSelector(getIsEnterprise);
  const { hasMultitenancy, isHosted } = useSelector(getFeatures);
  const multitenancy = hasMultitenancy || isEnterprise || isHosted;
  const dispatch = useDispatch();

  const { classes } = useStyles();

  const handleClose = () => {
    setAnchorEl(null);
    setTenantSwitcherShowing(false);
  };

  const handleSwitchTenant = id => dispatch(switchUserOrganization(id));

  const onLogoutClick = () => {
    setAnchorEl(null);
    dispatch(logoutUser()).then(() => window.location.replace('/ui/'));
  };

  const onToggleTooltips = () => dispatch(setAllTooltipsReadState(showHelptips ? READ_STATES.read : READ_STATES.unread));

  return (
    <>
      <Button className={classes.dropDown} onClick={e => setAnchorEl(e.currentTarget)} startIcon={<AccountCircleIcon className={classes.buttonColor} />}>
        {email}
      </Button>
      <Menu
        anchorEl={anchorEl}
        className={classes.dropDown}
        onClose={handleClose}
        open={Boolean(anchorEl)}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
      >
        <MenuItem component={Link} to="/settings/my-profile" onClick={handleClose}>
          My profile
        </MenuItem>
        <Divider />
        {!!(multitenancy && name) && (
          <MenuItem component={Link} dense to="/settings/organization-and-billing" onClick={handleClose} className={classes.organization}>
            <div>
              <Typography variant="caption" className="muted">
                My organization
              </Typography>
              <Typography variant="subtitle2">{name}</Typography>
            </div>
          </MenuItem>
        )}
        {tenants.length > 1 && (
          <div>
            <Divider style={{ marginBottom: 0 }} />
            <Accordion className={classes.accordion} square expanded={tenantSwitcherShowing} onChange={() => setTenantSwitcherShowing(toggle)}>
              <AccordionSummary expandIcon={<ExpandMore />}>Switch organization</AccordionSummary>
              <AccordionDetails className="padding-left-none padding-right-none">
                {tenants.map(({ id, name }) => (
                  <MenuItem className="padding-left padding-right" key={id} onClick={() => handleSwitchTenant(id)}>
                    {name}
                  </MenuItem>
                ))}
              </AccordionDetails>
            </Accordion>
          </div>
        )}
        <Divider />
        <MenuItem component={Link} to="/settings/global-settings" onClick={handleClose}>
          Settings
        </MenuItem>
        <MenuItem onClick={onToggleTooltips}>{`Mark help tips as ${showHelptips ? '' : 'un'}read`}</MenuItem>
        <MenuItem component={Link} to="/help/get-started" onClick={handleClose}>
          Help & support
        </MenuItem>
        <MenuItem onClick={onLogoutClick}>
          <ListItemText primary="Log out" />
          <ListItemSecondaryAction>
            <IconButton>
              <ExitIcon className={classes.exitIcon} />
            </IconButton>
          </ListItemSecondaryAction>
        </MenuItem>
      </Menu>
    </>
  );
};

export const Header = ({ mode }) => {
  const { classes } = useStyles();
  const [gettingUser, setGettingUser] = useState(false);
  const [hasOfferCookie, setHasOfferCookie] = useState(false);

  const organization = useSelector(getOrganization);
  const { total: acceptedDevices = 0 } = useSelector(getAcceptedDevices);
  const announcement = useSelector(state => state.app.hostedAnnouncement);
  const deviceLimit = useSelector(getDeviceLimit);
  const firstLoginAfterSignup = useSelector(state => state.app.firstLoginAfterSignup);
  const { trackingConsentGiven: hasTrackingEnabled } = useSelector(getUserSettings);
  const inProgress = useSelector(state => state.deployments.byStatus.inprogress.total);
  const isEnterprise = useSelector(getIsEnterprise);
  const { isDemoMode: demo, isHosted } = useSelector(getFeatures);
  const { isSearching, searchTerm, refreshTrigger } = useSelector(state => state.app.searchState);
  const { pending: pendingDevices } = useSelector(getDeviceCountsByStatus);
  const userSettingInitialized = useSelector(state => state.users.settingsInitialized);
  const user = useSelector(getCurrentUser);
  const { token } = useSelector(getCurrentSession);
  const userId = useDebounce(user.id, TIMEOUTS.debounceDefault);

  const dispatch = useDispatch();
  const deviceTimer = useRef();

  useEffect(() => {
    if ((!userId || !user.email?.length || !userSettingInitialized) && !gettingUser && token) {
      setGettingUser(true);
      dispatch(initializeSelf());
      return;
    }
    Tracking.setTrackingEnabled(hasTrackingEnabled);
    if (hasTrackingEnabled && user.id && organization.id) {
      Tracking.setOrganizationUser(organization, user);
      if (firstLoginAfterSignup) {
        Tracking.pageview('/signup/complete');
        dispatch(setFirstLoginAfterSignup(false));
      }
    }
  }, [dispatch, firstLoginAfterSignup, gettingUser, hasTrackingEnabled, organization, token, user, user.email, userId, userSettingInitialized]);

  useEffect(() => {
    const showOfferCookie = cookies.get('offer') === currentOffer.name;
    setHasOfferCookie(showOfferCookie);
    clearInterval(deviceTimer.current);
    deviceTimer.current = setInterval(() => dispatch(getAllDeviceCounts()), TIMEOUTS.refreshDefault);
    return () => {
      clearInterval(deviceTimer.current);
    };
  }, [dispatch]);

  const onSearch = useCallback((searchTerm, refreshTrigger) => dispatch(setSearchState({ refreshTrigger, searchTerm, page: 1 })), [dispatch]);

  const setHideOffer = () => {
    cookies.set('offer', currentOffer.name, { path: '/', maxAge: 2629746 });
    setHasOfferCookie(true);
  };

  const showOffer =
    isHosted && moment().isBefore(currentOffer.expires) && (organization.trial ? currentOffer.trial : currentOffer[organization.plan]) && !hasOfferCookie;

  const headerLogo = isDarkMode(mode) ? (isEnterprise ? whiteEnterpriseLogo : whiteLogo) : isEnterprise ? enterpriseLogo : logo;

  return (
    <Toolbar id="fixedHeader" className={showOffer ? `${classes.header} ${classes.banner}` : classes.header}>
      {!!announcement && (
        <Announcement
          announcement={announcement}
          errorIconClassName={classes.redAnnouncementIcon}
          iconClassName={classes.demoAnnouncementIcon}
          sectionClassName={classes.demoTrialAnnouncement}
          onHide={() => dispatch(setHideAnnouncement(true))}
        />
      )}
      {showOffer && <OfferHeader onHide={setHideOffer} />}
      <div className="flexbox space-between">
        <div className="flexbox center-aligned">
          <Link to="/">
            <img id="logo" src={headerLogo} />
          </Link>
          {demo && <DemoNotification iconClassName={classes.demoAnnouncementIcon} sectionClassName={classes.demoTrialAnnouncement} />}
          {organization.trial && (
            <TrialNotification
              expiration={organization.trial_expiration}
              iconClassName={classes.demoAnnouncementIcon}
              sectionClassName={classes.demoTrialAnnouncement}
            />
          )}
        </div>
        <Search isSearching={isSearching} searchTerm={searchTerm} onSearch={onSearch} trigger={refreshTrigger} />
        <div className="flexbox center-aligned">
          <DeviceNotifications pending={pendingDevices} total={acceptedDevices} limit={deviceLimit} />
          <DeploymentNotifications inprogress={inProgress} />
          <AccountMenu />
        </div>
      </div>
    </Toolbar>
  );
};

export default Header;
