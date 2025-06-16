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
import { useCallback, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { Alert, Button, Collapse, DialogActions, DialogContent, FormControlLabel, MenuItem, Select, Switch, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import storeActions from '@northern.tech/store/actions';
import { SSO_TYPES } from '@northern.tech/store/constants';
import { getCurrentSession, getFeatures, getIsEnterprise, getIsPreview, getOrganization, getSsoConfig, getUserRoles } from '@northern.tech/store/selectors';
import { changeSsoConfig, deleteSsoConfig, downloadLicenseReport, getSsoConfigs, getUserOrganization, storeSsoConfig } from '@northern.tech/store/thunks';
import { createFileDownload, toggle } from '@northern.tech/utils/helpers';
import dayjs from 'dayjs';

import OrganizationSettingsItem from './OrganizationSettingsItem';
import { SSOConfig } from './SSOConfig';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(({ spacing }) => ({
  orgInfo: { gap: spacing(2) },
  ssoSelect: { minWidth: 265 }
}));

// unlike the ExpandableAttribute, the token should not be visible by default - thus a separate component
const TenantToken = ({ expanded, onClick, token }: { expanded: boolean; onClick: () => void; token: string }) => (
  <>
    <div className="tenant-token-text">{expanded ? token : `${token.substring(0, token.length / 2)}...`}</div>
    {!expanded && (
      <div className="clickable link-color margin-top-x-small margin-left-x-small" onClick={onClick}>
        <b>Show more</b>
      </div>
    )}
  </>
);

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
  const { id: tenantId, name: orgName, tenant_token = '' } = useSelector(getOrganization);
  const ssoConfig = useSelector(getSsoConfig);
  const dispatch = useDispatch();
  const { token } = useSelector(getCurrentSession);
  const { classes } = useStyles();

  useEffect(() => {
    dispatch(getUserOrganization());
  }, [dispatch]);

  useEffect(() => {
    if (!isEnterprise) {
      return;
    }
    dispatch(getSsoConfigs());
  }, [dispatch, isEnterprise]);

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
    dispatch(downloadLicenseReport())
      .unwrap()
      .then(report => createFileDownload(report, `Mender-license-report-${dayjs().format('YYYY-MM-DD')}`, token));

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
    <div style={{ maxWidth: 750 }}>
      <Typography variant="h6">My organization</Typography>
      <div className={`flexbox column ${classes.orgInfo}`}>
        <OrganizationSettingsItem title="Organization ID" secondary={tenantId} sideBarContent={<CopyTextToClipboard notify={false} token={tenantId} />} />
        <OrganizationSettingsItem title="Organization name" secondary={orgName} sideBarContent={<CopyTextToClipboard notify={false} token={orgName} />} />
        <OrganizationSettingsItem
          title="Organization token"
          description="The token is unique for your organization and ensures that only devices that you own are able to connect to your account."
          secondary={<TenantToken expanded={showTokenWarning} onClick={onTokenExpansion} token={tenant_token} />}
          sideBarContent={<CopyTextToClipboard notify={false} onCopy={onTokenExpansion} token={tenant_token} />}
          notification={
            showTokenWarning && (
              <Alert severity="warning">
                Do not share your organization token with others. Treat this token like a password, as it can be used to request authorization for new devices.
              </Alert>
            )
          }
        />
        {isEnterprise && isAdmin && (
          <div>
            <FormControlLabel
              className="margin-bottom-small margin-left-none"
              control={<Switch checked={!isResettingSSO && (hasSingleSignOn || isConfiguringSSO)} className="margin-left-small" onChange={onSSOClick} />}
              label="Enable Single Sign-On"
              labelPlacement="start"
            />
          </div>
        )}
      </div>

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
  <BaseDialog open={open} title="Change Single Sign-On type" onClose={dismiss}>
    <DialogContent style={{ overflow: 'hidden' }}>Are you sure you want to change SSO type? This will lose your current settings.</DialogContent>
    <DialogActions>
      <Button style={{ marginRight: 10 }} onClick={dismiss}>
        Cancel
      </Button>
      <Button variant="contained" color="primary" onClick={() => submit()}>
        Change
      </Button>
    </DialogActions>
  </BaseDialog>
);

export default Organization;
