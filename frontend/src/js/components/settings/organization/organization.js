// Copyright 2017 Northern.tech AS
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
import React, { useCallback, useEffect, useState } from 'react';
import CopyToClipboard from 'react-copy-to-clipboard';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { FileCopy as CopyPasteIcon } from '@mui/icons-material';
import { Button, Checkbox, Collapse, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, List, MenuItem, Select } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import copy from 'copy-to-clipboard';
import dayjs from 'dayjs';

import { setSnackbar } from '../../../actions/appActions';
import {
  changeSsoConfig,
  deleteSsoConfig,
  downloadLicenseReport,
  getSsoConfigs,
  getUserOrganization,
  storeSsoConfig
} from '../../../actions/organizationActions';
import { TIMEOUTS, yes } from '../../../constants/appConstants';
import { SSO_TYPES } from '../../../constants/organizationConstants.js';
import { createFileDownload, toggle } from '../../../helpers';
import { getCurrentSession, getFeatures, getIsEnterprise, getIsPreview, getOrganization, getSsoConfig, getUserRoles } from '../../../selectors';
import ExpandableAttribute from '../../common/expandable-attribute';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';
import Billing from './billing';
import OrganizationSettingsItem, { maxWidth } from './organizationsettingsitem';
import { SSOConfig } from './ssoconfig';

const useStyles = makeStyles()(theme => ({
  copyNotification: { height: 15 },
  deviceLimitBar: { backgroundColor: theme.palette.grey[500], margin: '15px 0' },
  tenantInfo: { marginTop: 11, paddingBottom: 3, 'span': { marginLeft: theme.spacing(0.5), color: theme.palette.text.disabled } },
  tenantToken: { width: `calc(${maxWidth}px - ${theme.spacing(4)})` },
  tokenTitle: { paddingRight: 10 },
  tokenExplanation: { margin: '1em 0' },
  ssoSelect: { minWidth: 265 }
}));

export const OrgHeader = () => {
  const { classes } = useStyles();
  return (
    <div className="flexbox center-aligned">
      <div className={classes.tokenTitle}>Organization token</div>
      <MenderHelpTooltip id={HELPTOOLTIPS.tenantToken.id} disableHoverListener={false} placement="top" />
    </div>
  );
};

export const CopyTextToClipboard = ({ onCopy = yes, token }) => {
  const [copied, setCopied] = useState(false);
  const { classes } = useStyles();

  const onCopied = () => {
    setCopied(true);
    onCopy();
    setTimeout(() => setCopied(false), TIMEOUTS.fiveSeconds);
  };

  return (
    <div>
      <CopyToClipboard text={token} onCopy={onCopied}>
        <Button startIcon={<CopyPasteIcon />}>Copy to clipboard</Button>
      </CopyToClipboard>
      <p className={classes.copyNotification}>{copied && <span className="green fadeIn">Copied to clipboard.</span>}</p>
    </div>
  );
};

export const Organization = () => {
  const [hasSingleSignOn, setHasSingleSignOn] = useState(false);
  const [isConfiguringSSO, setIsConfiguringSSO] = useState(false);
  const [isResettingSSO, setIsResettingSSO] = useState(false);
  const [showTokenWarning, setShowTokenWarning] = useState(false);
  const [newSso, setNewSso] = useState('');
  const [selectedSsoItem, setSelectedSsoItem] = useState(undefined);
  const isEnterprise = useSelector(getIsEnterprise);
  const { isAdmin } = useSelector(getUserRoles);
  const canPreview = useSelector(getIsPreview);
  const { isHosted } = useSelector(getFeatures);
  const org = useSelector(getOrganization);
  const ssoConfig = useSelector(getSsoConfig);
  const dispatch = useDispatch();
  const { token } = useSelector(getCurrentSession);
  const { classes } = useStyles();

  useEffect(() => {
    dispatch(getUserOrganization());
  }, [dispatch]);

  useEffect(() => {
    dispatch(getSsoConfigs());
  }, [dispatch]);

  useEffect(() => {
    setHasSingleSignOn(!!ssoConfig);
    setIsConfiguringSSO(!!ssoConfig);
    if (ssoConfig) {
      setSelectedSsoItem(SSO_TYPES[ssoConfig.type]);
    }
  }, [ssoConfig]);

  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);

  const onSaveSSOSettings = useCallback(
    (id, config) => {
      const { contentType } = SSO_TYPES[selectedSsoItem.type];
      if (isResettingSSO) {
        return dispatch(deleteSsoConfig(ssoConfig)).then(() => setIsResettingSSO(false));
      }
      if (id) {
        return dispatch(changeSsoConfig({ id, config, contentType }));
      }
      return dispatch(storeSsoConfig({ config, contentType }));
    },
    [isResettingSSO, dispatch, ssoConfig, selectedSsoItem]
  );

  const onCancelSSOSettings = () => {
    setIsResettingSSO(false);
    setIsConfiguringSSO(hasSingleSignOn);
  };

  const onTokenExpansion = useCallback(() => setShowTokenWarning(true), []);

  const onDownloadReportClick = () =>
    dispatch(downloadLicenseReport()).then(report => createFileDownload(report, `Mender-license-report-${dayjs().format('YYYY-MM-DD')}`, token));

  const onTenantInfoClick = () => {
    copy(`Organization: ${org.name}, Tenant ID: ${org.id}`);
    setSnackbar('Copied to clipboard');
  };

  const onSSOClick = () => {
    if (hasSingleSignOn) {
      setIsConfiguringSSO(false);
      return setIsResettingSSO(true);
    }
    setIsConfiguringSSO(toggle);
  };

  const onSsoSelect = useCallback(
    ({ target: { value: type = '' } }) => {
      if (ssoConfig) {
        setNewSso(type);
      } else {
        setSelectedSsoItem(SSO_TYPES[type]);
      }
    },
    [ssoConfig]
  );

  const changeSSO = () =>
    dispatch(deleteSsoConfig(ssoConfig)).then(() => {
      setSelectedSsoItem(SSO_TYPES[newSso]);
      setIsConfiguringSSO(true);
      setNewSso('');
    });

  return (
    <div className="margin-top-small">
      <h2 className="margin-top-small">Organization and billing</h2>
      <List>
        <OrganizationSettingsItem
          title="Organization name"
          content={{
            action: { action: onTenantInfoClick, internal: true },
            description: (
              <div className={`clickable ${classes.tenantInfo}`} onClick={onTenantInfoClick}>
                {org.name}
                <span>({org.id})</span>
              </div>
            )
          }}
        />
        <OrganizationSettingsItem
          title={<OrgHeader />}
          content={{}}
          secondary={
            <>
              <ExpandableAttribute
                className={classes.tenantToken}
                component="div"
                disableGutters
                dividerDisabled
                key="org_token"
                onExpansion={onTokenExpansion}
                secondary={org.tenant_token}
                textClasses={{ secondary: 'inventory-text tenant-token-text' }}
              />
              {showTokenWarning && (
                <p className="warning">
                  <b>Important</b>
                  <br />
                  Do not share your organization token with others. Treat this token like a password, as it can be used to request authorization for new
                  devices.
                </p>
              )}
            </>
          }
          sideBarContent={<CopyTextToClipboard onCopy={onTokenExpansion} token={org.tenant_token} />}
        />
      </List>

      {isEnterprise && isAdmin && (
        <div>
          <FormControlLabel
            className="margin-bottom-small"
            control={<Checkbox checked={!isResettingSSO && (hasSingleSignOn || isConfiguringSSO)} onChange={onSSOClick} />}
            label="Enable Single Sign-On"
          />
        </div>
      )}

      {isConfiguringSSO && (
        <div>
          <Select className={classes.ssoSelect} displayEmpty onChange={onSsoSelect} value={selectedSsoItem?.type || ''}>
            <MenuItem value="">Select type</MenuItem>
            {Object.values(SSO_TYPES).map(item => (
              <MenuItem key={item.type} value={item.type}>
                <div className="capitalized-start">{item.title}</div>
              </MenuItem>
            ))}
          </Select>
        </div>
      )}

      <div className="flexbox center-aligned">
        {isResettingSSO && !isConfiguringSSO && (
          <>
            <Button onClick={onCancelSSOSettings}>Cancel</Button>
            <Button onClick={onSaveSSOSettings} disabled={!hasSingleSignOn} variant="contained">
              Save
            </Button>
          </>
        )}
      </div>
      {selectedSsoItem && (
        <div className="margin-top">
          <Collapse className="margin-left-large" in={isConfiguringSSO}>
            <SSOConfig
              ssoItem={selectedSsoItem}
              config={ssoConfig}
              onSave={onSaveSSOSettings}
              onCancel={onCancelSSOSettings}
              setSnackbar={dispatchedSetSnackbar}
              token={token}
            />
          </Collapse>
        </div>
      )}

      {isHosted && <Billing />}
      {(canPreview || !isHosted) && isEnterprise && isAdmin && (
        <Button className="margin-top" onClick={onDownloadReportClick} variant="contained">
          Download license report
        </Button>
      )}
      <ChangeSsoDialog dismiss={() => setNewSso(undefined)} open={!!newSso} submit={changeSSO} />
    </div>
  );
};

const ChangeSsoDialog = ({ dismiss, open, submit }) => (
  <Dialog open={open}>
    <DialogTitle>Change Single Sign-On type</DialogTitle>
    <DialogContent style={{ overflow: 'hidden' }}>Are you sure you want to change SSO type? This will lose your current settings.</DialogContent>
    <DialogActions>
      <Button style={{ marginRight: 10 }} onClick={dismiss}>
        Cancel
      </Button>
      <Button variant="contained" color="primary" onClick={() => submit()}>
        Change
      </Button>
    </DialogActions>
  </Dialog>
);

export default Organization;
