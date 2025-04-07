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
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { AccountCircle as AccountCircleIcon, ExitToApp as ExitIcon, ExpandMore } from '@mui/icons-material';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Button,
  Chip,
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

import Search from '@northern.tech/common-ui/Search';
import storeActions from '@northern.tech/store/actions';
import { READ_STATES, TIMEOUTS } from '@northern.tech/store/constants';
import {
  getAcceptedDevices,
  getCurrentSession,
  getCurrentUser,
  getDeviceCountsByStatus,
  getDeviceLimit,
  getFeatures,
  getFeedbackProbability,
  getHostedAnnouncement,
  getIsEnterprise,
  getIsFirstLogin,
  getIsServiceProvider,
  getOrganization,
  getSearchState,
  getShowHelptips,
  getUserRoles,
  getUserSettings
} from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { useAppInit } from '@northern.tech/store/storehooks';
import {
  getAllDeviceCounts,
  getUserOrganization,
  initializeSelf,
  logoutUser,
  setAllTooltipsReadState,
  setFirstLoginAfterSignup,
  setHideAnnouncement,
  setSearchState,
  switchUserOrganization
} from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { toggle } from '@northern.tech/utils/helpers';
import dayjs from 'dayjs';
import { jwtDecode } from 'jwt-decode';
import Cookies from 'universal-cookie';

import enterpriseLogo from '../../../assets/img/headerlogo-enterprise.png';
import logo from '../../../assets/img/headerlogo.png';
import whiteEnterpriseLogo from '../../../assets/img/whiteheaderlogo-enterprise.png';
import whiteLogo from '../../../assets/img/whiteheaderlogo.png';
import Tracking from '../../tracking';
import Announcement from './Announcement';
import DemoNotification from './DemoNotification';
import DeploymentNotifications from './DeploymentNotifications';
import DeviceNotifications from './DeviceNotifications';
import OfferHeader from './OfferHeader';
import TrialNotification from './TrialNotification';
import { DeviceCount } from './devicecount';

const { setShowFeedbackDialog } = storeActions;

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
  banner: { gridTemplateRows: `1fr ${theme.mixins.toolbar.minHeight}px` },
  buttonColor: { color: theme.palette.grey[600] },
  demoAnnouncementIcon: {
    height: 16,
    color: theme.palette.primary.main,
    '&.MuiButton-textPrimary': {
      color: theme.palette.primary.main,
      height: 'inherit'
    }
  },
  demoTrialAnnouncement: {
    fontSize: 14,
    height: 'auto'
  },
  dropDown: {
    height: '100%',
    textTransform: 'none',
    [`.${menuItemClasses.root}:hover, .${listItemTextClasses.root}:hover`]: {
      color: theme.palette.text.link
    }
  },
  exitIcon: { color: theme.palette.grey[600], fill: theme.palette.grey[600] },
  header: {
    minHeight: 'unset',
    paddingLeft: theme.spacing(4),
    paddingRight: theme.spacing(5),
    width: '100%',
    borderBottom: `1px solid ${theme.palette.grey[100]}`,
    display: 'grid',
    '#logo': {
      minWidth: 142,
      height: theme.spacing(6),
      marginRight: 25
    }
  },
  headerSection: {
    height: 24,
    fontSize: '14px',
    color: theme.palette.grey[600],
    margin: '14px 0',
    paddingLeft: theme.spacing(4.5),
    paddingRight: theme.spacing(4.5),
    borderRight: `1px solid ${theme.palette.grey[300]}`,
    display: 'flex',
    alignItems: 'center',
    lineHeight: 'initial',
    '&:hover': {
      color: theme.palette.grey[700]
    }
  },
  organization: { marginBottom: theme.spacing() },
  redAnnouncementIcon: {
    color: theme.palette.error.dark
  },
  search: { alignSelf: 'center' }
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
  const dispatch = useAppDispatch();

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

const HEX_BASE = 16;
const date = dayjs().toISOString().split('T')[0];
const pickAUser = ({ jti, probability }) => {
  const daySessionUniqueId = `${jti}-${date}`; // jti can be unique for multiple user sessions, combined with a check at most once per day should be enough
  const hashBuffer = new TextEncoder().encode(daySessionUniqueId);
  return crypto.subtle.digest('SHA-256', hashBuffer).then(hashArrayBuffer => {
    // convert the hash buffer to a hex string for easier processing towards a number
    const hashHex = Array.from(new Uint8Array(hashArrayBuffer))
      .map(byte => byte.toString(HEX_BASE).padStart(2, '0'))
      .join('');
    const hashInt = parseInt(hashHex.slice(0, 8), HEX_BASE); // convert the hex string to an integer, use first 8 chars for simplicity
    const normalizedValue = hashInt / Math.pow(2, 32); // normalize the integer to a value between 0 and 1, within the 32bit range browsers default to
    // select the user if the normalized value is below the probability threshold
    return normalizedValue < probability;
  });
};
export const Header = ({ isDarkMode }) => {
  const { classes } = useStyles();
  const [gettingUser, setGettingUser] = useState(false);
  const [hasOfferCookie, setHasOfferCookie] = useState(false);

  const organization = useSelector(getOrganization);
  const { total: acceptedDevices = 0 } = useSelector(getAcceptedDevices);
  const announcement = useSelector(getHostedAnnouncement);
  const deviceLimit = useSelector(getDeviceLimit);
  const feedbackProbability = useSelector(getFeedbackProbability);
  const firstLoginAfterSignup = useSelector(getIsFirstLogin);
  const { feedbackCollectedAt, trackingConsentGiven: hasTrackingEnabled } = useSelector(getUserSettings);
  const { isAdmin } = useSelector(getUserRoles);
  const inProgress = useSelector(state => state.deployments.byStatus.inprogress.total);
  const isEnterprise = useSelector(getIsEnterprise);
  const { hasFeedbackEnabled, isDemoMode: demo, isHosted } = useSelector(getFeatures);
  const { isSearching, searchTerm, refreshTrigger } = useSelector(getSearchState);
  const { pending: pendingDevices } = useSelector(getDeviceCountsByStatus);
  const userSettingInitialized = useSelector(state => state.users.settingsInitialized);
  const user = useSelector(getCurrentUser);
  const { token } = useSelector(getCurrentSession);
  const userId = useDebounce(user.id, TIMEOUTS.debounceDefault);
  const isSp = useSelector(getIsServiceProvider);
  const { device_count: spDeviceUtilization, device_limit: tenantDeviceLimit, service_provider } = useSelector(getOrganization);
  const dispatch = useAppDispatch();
  const deviceTimer = useRef();
  const feedbackTimer = useRef();

  useAppInit(userId);

  useEffect(() => {
    if ((!userId || !user.email?.length || !userSettingInitialized) && !gettingUser && token) {
      setGettingUser(true);
      dispatch(getUserOrganization());
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
    if (!service_provider) {
      deviceTimer.current = setInterval(() => dispatch(getAllDeviceCounts()), TIMEOUTS.refreshDefault);
    }
    return () => {
      clearInterval(deviceTimer.current);
      clearTimeout(feedbackTimer.current);
    };
  }, [dispatch, service_provider]);

  useEffect(() => {
    const today = dayjs();
    const diff = dayjs.duration(dayjs(feedbackCollectedAt).diff(today));
    const isFeedbackEligible = diff.asMonths() > 3;
    if (!hasFeedbackEnabled || !userSettingInitialized || !token || (feedbackCollectedAt && !isFeedbackEligible)) {
      return;
    }
    const { jti } = jwtDecode(token);
    pickAUser({ jti, probability: feedbackProbability }).then(isSelected => {
      feedbackTimer.current = setTimeout(() => dispatch(setShowFeedbackDialog(isSelected)), TIMEOUTS.threeSeconds);
    });
  }, [dispatch, feedbackCollectedAt, feedbackProbability, hasFeedbackEnabled, isAdmin, userSettingInitialized, token]);

  const onSearch = useCallback((searchTerm, refreshTrigger) => dispatch(setSearchState({ refreshTrigger, searchTerm, page: 1 })), [dispatch]);

  const setHideOffer = () => {
    cookies.set('offer', currentOffer.name, { path: '/', maxAge: 2629746 });
    setHasOfferCookie(true);
  };

  const showOffer =
    isHosted && dayjs().isBefore(currentOffer.expires) && (organization.trial ? currentOffer.trial : currentOffer[organization.plan]) && !hasOfferCookie;

  const headerLogo = isDarkMode ? (isEnterprise ? whiteEnterpriseLogo : whiteLogo) : isEnterprise ? enterpriseLogo : logo;

  return (
    <Toolbar id="fixedHeader" className={showOffer ? `${classes.header} ${classes.banner}` : classes.header}>
      {!!announcement && (
        <Announcement
          announcement={announcement}
          errorIconClassName={classes.redAnnouncementIcon}
          iconClassName={classes.demoAnnouncementIcon}
          sectionClassName={classes.demoTrialAnnouncement}
          onHide={() => dispatch(setHideAnnouncement({ shouldHide: true }))}
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
        {isSp ? (
          <>
            {tenantDeviceLimit > 0 && <DeviceCount current={spDeviceUtilization} max={tenantDeviceLimit} variant="common" />}
            <div className="flexbox center-aligned">
              <div className={classes.headerSection}>
                <Chip className="bold muted uppercased" label="Service Provider" />
              </div>
              <AccountMenu />
            </div>
          </>
        ) : (
          <>
            <Search className={classes.search} isSearching={isSearching} searchTerm={searchTerm} onSearch={onSearch} trigger={refreshTrigger} />
            <div className="flexbox center-aligned">
              <DeviceNotifications className={classes.headerSection} pending={pendingDevices} total={acceptedDevices} limit={deviceLimit} />
              <DeploymentNotifications className={classes.headerSection} inprogress={inProgress} />
              <AccountMenu />
            </div>
          </>
        )}
      </div>
    </Toolbar>
  );
};

export default Header;
