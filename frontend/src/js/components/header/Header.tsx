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
import { useCallback, useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router';

import { AccountCircle as AccountCircleIcon, LogoutOutlined as ExitIcon, ExpandMore } from '@mui/icons-material';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Button,
  Chip,
  Divider,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Toolbar,
  Typography,
  accordionSummaryClasses,
  listItemIconClasses,
  menuClasses,
  menuItemClasses,
  textFieldClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { Link } from '@northern.tech/common-ui/Link';
import Search from '@northern.tech/common-ui/Search';
import storeActions from '@northern.tech/store/actions';
import { READ_STATES, TIMEOUTS } from '@northern.tech/store/constants';
import {
  getCurrentSession,
  getCurrentUser,
  getDeploymentsByStatus,
  getDeviceCountsByStatus,
  getFeatures,
  getFeedbackProbability,
  getHostedAnnouncement,
  getIsEnterprise,
  getIsFirstLogin,
  getIsServiceProvider,
  getOrganization,
  getReadAllHelptips,
  getSearchState,
  getTooltipsById,
  getUserRoles,
  getUserSettings,
  getUserSettingsInitialized
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
import durationDayJs from 'dayjs/plugin/duration.js';
import { jwtDecode } from 'jwt-decode';
import Cookies from 'universal-cookie';

import enterpriseLogo from '../../../assets/img/headerlogo-enterprise.png';
import logo from '../../../assets/img/headerlogo.png';
import whiteEnterpriseLogo from '../../../assets/img/whiteheaderlogo-enterprise.png';
import whiteLogo from '../../../assets/img/whiteheaderlogo.png';
import Tracking from '../../tracking';
import Announcement from './Announcement';
import DeploymentNotifications from './DeploymentNotifications';
import DeviceNotifications from './DeviceNotifications';
import OfferHeader from './OfferHeader';
import TrialNotification from './TrialNotification';

dayjs.extend(durationDayJs);

const { setShowFeedbackDialog } = storeActions;

// Change this when a new feature/offer is introduced
const currentOffer = {
  name: 'ai-feature',
  expires: '2026-03-31',
  trial: false,
  os: true,
  professional: true,
  enterprise: true
};

const cookies = new Cookies();

const useStyles = makeStyles()(theme => ({
  accordion: {
    border: 'none',
    [`& .${accordionSummaryClasses.root}`]: {
      minHeight: 'unset',
      padding: theme.spacing(0.75, 2),
      '&:hover': { backgroundColor: theme.palette.action.hover }
    }
  },
  accountButton: { maxWidth: 250 },
  demoAnnouncementIcon: {
    height: 16
  },
  demoTrialAnnouncement: {
    fontSize: 14
  },
  exitIcon: { color: theme.palette.grey[600] },
  exitIconWrapper: {
    [`&.${listItemIconClasses.root}`]: { minWidth: 'auto' }
  },
  header: {
    borderBottom: `1px solid ${theme.palette.divider}`,
    color: theme.palette.text.secondary,
    display: 'grid',
    gridTemplateColumns: 'max-content 1fr max-content'
  },
  logo: {
    minWidth: 142,
    height: theme.spacing(6)
  },
  divider: {
    height: theme.spacing(3)
  },
  menuPaper: {
    marginTop: theme.spacing(),
    [`&.${menuClasses.paper}`]: { maxWidth: 220 },
    [`& .${menuItemClasses.root}`]: { fontSize: '1rem', padding: theme.spacing(0.75, 2) }
  },
  redAnnouncementIcon: {
    color: theme.palette.error.dark
  },
  search: {
    justifyContent: 'center',
    display: 'flex',
    [`& .${textFieldClasses.root}`]: {
      width: 220,
      maxWidth: 640,
      transition: 'width 0.1s ease',
      '&:focus-within': {
        width: '100%'
      }
    }
  },
  serviceProvider: {
    gridTemplateColumns: '1fr max-content'
  }
}));

const AccountMenu = () => {
  const [anchorEl, setAnchorEl] = useState(null);
  const [tenantSwitcherShowing, setTenantSwitcherShowing] = useState(false);
  const hasReadHelptips = useSelector(getReadAllHelptips);
  const { email, tenants = [] } = useSelector(getCurrentUser);
  const tooltips = useSelector(getTooltipsById);
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

  const onToggleTooltips = () =>
    dispatch(setAllTooltipsReadState({ readState: hasReadHelptips ? READ_STATES.unread : READ_STATES.read, tooltipIds: Object.keys(tooltips) }));

  return (
    <>
      <Button
        className={`flexbox align-items-center ${classes.accountButton}`}
        onClick={e => setAnchorEl(e.currentTarget)}
        startIcon={<AccountCircleIcon />}
        color="inherit"
        variant="text"
      >
        <Typography variant="subtitle2" className="text-overflow">
          {email}
        </Typography>
      </Button>
      <Menu
        anchorEl={anchorEl}
        onClose={handleClose}
        open={Boolean(anchorEl)}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        slotProps={{ paper: { className: classes.menuPaper } }}
      >
        <MenuItem component={RouterLink} to="/settings/my-profile" onClick={handleClose} divider>
          My profile
        </MenuItem>
        {!!(multitenancy && name) && (
          <MenuItem component={RouterLink} dense divider to="/settings/organization" onClick={handleClose}>
            <div className="text-overflow">
              <Typography variant="body2" color="textSecondary">
                My organization
              </Typography>
              <Typography className="text-overflow">{name}</Typography>
            </div>
          </MenuItem>
        )}
        {tenants.length > 1 && (
          <div>
            <Accordion
              className={`padding-none ${classes.accordion}`}
              disableGutters
              square
              expanded={tenantSwitcherShowing}
              onChange={() => setTenantSwitcherShowing(toggle)}
            >
              <AccordionSummary expandIcon={<ExpandMore />} slotProps={{ content: { className: 'margin-none' } }}>
                <Typography className="margin-right-small">Switch organization</Typography>
              </AccordionSummary>
              <AccordionDetails className="padding-none padding-top-x-small">
                {tenants.map(({ id, name }) => (
                  <MenuItem key={id} onClick={() => handleSwitchTenant(id)}>
                    {name}
                  </MenuItem>
                ))}
              </AccordionDetails>
            </Accordion>
            <Divider />
          </div>
        )}
        <MenuItem component={RouterLink} to="/settings/global-settings" onClick={handleClose}>
          Settings
        </MenuItem>
        <MenuItem onClick={onToggleTooltips}>{`Mark help tips as ${hasReadHelptips ? 'un' : ''}read`}</MenuItem>
        <MenuItem component={RouterLink} to="/help/get-started" onClick={handleClose}>
          Help & support
        </MenuItem>
        <MenuItem onClick={onLogoutClick}>
          <ListItemText primary="Log out" slotProps={{ primary: { variant: 'body1' } }} />
          <ListItemIcon className={classes.exitIconWrapper}>
            <ExitIcon className={classes.exitIcon} />
          </ListItemIcon>
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
  const announcement = useSelector(getHostedAnnouncement);
  const feedbackProbability = useSelector(getFeedbackProbability);
  const firstLoginAfterSignup = useSelector(getIsFirstLogin);
  const { feedbackCollectedAt, trackingConsentGiven: hasTrackingEnabled, firstLoginTimestamp } = useSelector(getUserSettings);
  const { isAdmin } = useSelector(getUserRoles);
  const { inprogress: inprogressDeployments } = useSelector(getDeploymentsByStatus);
  const { total: inProgress } = inprogressDeployments;
  const isEnterprise = useSelector(getIsEnterprise);
  const { hasAiEnabled, hasFeedbackEnabled, isHosted } = useSelector(getFeatures);
  const { searchTerm, refreshTrigger } = useSelector(getSearchState);
  const { accepted: acceptedDevices, pending: pendingDevices } = useSelector(getDeviceCountsByStatus);
  const userSettingInitialized = useSelector(getUserSettingsInitialized);
  const user = useSelector(getCurrentUser);
  const { token } = useSelector(getCurrentSession);
  const userId = useDebounce(user.id, TIMEOUTS.debounceDefault);
  const isSp = useSelector(getIsServiceProvider);
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
    if (!isSp) {
      deviceTimer.current = setInterval(() => dispatch(getAllDeviceCounts()), TIMEOUTS.refreshDefault);
    }
    return () => {
      clearInterval(deviceTimer.current);
      clearTimeout(feedbackTimer.current);
    };
  }, [dispatch, isSp]);

  useEffect(() => {
    const today = dayjs();
    const lastFeedbackCollectedMonthsAgo = today.diff(feedbackCollectedAt, 'months');
    const firstLoginDaysAgo = today.diff(firstLoginTimestamp, 'days');
    const isFeedbackEligible = lastFeedbackCollectedMonthsAgo > 6 && firstLoginDaysAgo > 14;
    if (!hasFeedbackEnabled || !userSettingInitialized || !token || (feedbackCollectedAt && !isFeedbackEligible)) {
      return;
    }
    const { jti } = jwtDecode(token);
    pickAUser({ jti, probability: feedbackProbability }).then(isSelected => {
      feedbackTimer.current = setTimeout(() => dispatch(setShowFeedbackDialog(isSelected)), TIMEOUTS.threeSeconds);
    });
  }, [dispatch, feedbackCollectedAt, feedbackProbability, hasFeedbackEnabled, isAdmin, userSettingInitialized, token, firstLoginTimestamp]);

  const onSearch = useCallback((searchTerm, refreshTrigger) => dispatch(setSearchState({ refreshTrigger, searchTerm, page: 1 })), [dispatch]);

  const setHideOffer = () => {
    cookies.set('offer', currentOffer.name, { path: '/', maxAge: 2629746 });
    setHasOfferCookie(true);
  };

  const showOffer =
    hasAiEnabled &&
    isHosted &&
    dayjs().isBefore(currentOffer.expires) &&
    (organization.trial ? currentOffer.trial : currentOffer[organization.plan]) &&
    !hasOfferCookie;

  const headerLogo = isDarkMode ? (isEnterprise ? whiteEnterpriseLogo : whiteLogo) : isEnterprise ? enterpriseLogo : logo;

  return (
    <div id="fixedHeader">
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
      <Toolbar className={`${classes.header} ${isSp ? classes.serviceProvider : ''}`}>
        <div className="flexbox align-items-center">
          <Link to="/">
            <img className={`${classes.logo} margin-right-medium`} src={headerLogo} />
          </Link>
          {organization.trial && <TrialNotification expiration={organization.trial_expiration} sectionClassName={classes.demoTrialAnnouncement} />}
        </div>
        {isSp ? (
          <div className="flexbox align-items-center">
            <Chip label="Service Provider" />
            <Divider className={`margin-left-x-small margin-right-x-small ${classes.divider}`} orientation="vertical" />
            <AccountMenu />
          </div>
        ) : (
          <>
            <Search className={classes.search} searchTerm={searchTerm} onSearch={onSearch} trigger={refreshTrigger} />
            <div className="flexbox align-items-center">
              <DeviceNotifications pending={pendingDevices} total={acceptedDevices} />
              <Divider className={`margin-left-x-small margin-right-x-small ${classes.divider}`} orientation="vertical" />
              <DeploymentNotifications inprogress={inProgress} />
              <Divider className={`margin-left-x-small margin-right-x-small ${classes.divider}`} orientation="vertical" />
              <AccountMenu />
            </div>
          </>
        )}
      </Toolbar>
    </div>
  );
};

export default Header;
