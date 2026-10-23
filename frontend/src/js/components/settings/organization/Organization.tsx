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
import { useSelector } from 'react-redux';

import { SaveAlt as SaveAltIcon } from '@mui/icons-material';
// material ui
import { Alert, Button, Chip, Collapse, Typography, formControlClasses, textFieldClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import { SettingsItem, settingsFieldWidths } from '@northern.tech/common-ui/SettingsItem';
import { SettingsSection } from '@northern.tech/common-ui/SettingsSection';
import { Select } from '@northern.tech/common-ui/forms/Select';
import { ALL_DEVICES, SSO_TYPES, uiPermissionsById } from '@northern.tech/store/constants';
import {
  getCurrentSession,
  getFeatures,
  getIsEnterprise,
  getIsPreview,
  getOrganization,
  getSsoConfig,
  getUserCapabilities,
  getUserRoles
} from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import {
  changeSsoConfig,
  deleteSsoConfig,
  downloadLicenseReport,
  getSsoConfigs,
  getTenantToken,
  getUserOrganization,
  storeSsoConfig
} from '@northern.tech/store/thunks';
import { createFileDownload, toggle } from '@northern.tech/utils/helpers';
import dayjs from 'dayjs';

import { SSOConfig } from './SSOConfig';

const useStyles = makeStyles()(() => ({
  tenantToken: { wordWrap: 'break-word' },
  widthLimit: {
    [`.${textFieldClasses.root},.${formControlClasses.root}`]: { width: settingsFieldWidths.large },
    ['.settings-item-main-content']: { gridTemplateColumns: 'minmax(0, max-content) 1fr' }
  }
}));

const ssoConfirmations = {
  change: { header: 'Change Single Sign-On type?', confirmButtonText: 'Change SSO type' },
  disable: { header: 'Disable Single Sign-On?', confirmButtonText: 'Disable SSO' }
};

// unlike the ExpandableAttribute, the token should not be visible by default - thus a separate component
const TenantToken = ({ expanded, onClick, token }: { expanded: boolean; onClick: () => void; token: string }) => (
  <>
    <div className="tenant-token-text">{expanded ? token : `${token.substring(0, 8)}...`}</div>
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
  const [isEditingSSO, setIsEditingSSO] = useState(false);
  const [isDisablingSSO, setIsDisablingSSO] = useState(false);
  const [showTokenWarning, setShowTokenWarning] = useState(false);
  const [newSso, setNewSso] = useState('');
  const [selectedSsoItem, setSelectedSsoItem] = useState(undefined);
  const isEnterprise = useSelector(getIsEnterprise);
  const { isAdmin } = useSelector(getUserRoles);
  const { groupsPermissions } = useSelector(getUserCapabilities);
  const canPreview = useSelector(getIsPreview);
  const { isHosted } = useSelector(getFeatures);
  const { id: tenantId, name: orgName, tenant_token = '' } = useSelector(getOrganization);
  const ssoConfig = useSelector(getSsoConfig);
  const dispatch = useAppDispatch();
  const { token } = useSelector(getCurrentSession);
  const { classes } = useStyles();
  const canAddDevices = groupsPermissions[ALL_DEVICES]?.includes(uiPermissionsById.manage.value);

  useEffect(() => {
    dispatch(getUserOrganization());
  }, [dispatch]);

  useEffect(() => {
    if (!canAddDevices) {
      return;
    }
    dispatch(getTenantToken());
  }, [canAddDevices, dispatch]);

  useEffect(() => {
    if (!isEnterprise) {
      return;
    }
    dispatch(getSsoConfigs());
  }, [dispatch, isEnterprise]);

  useEffect(() => {
    setHasSingleSignOn(!!ssoConfig);
    setIsConfiguringSSO(!!ssoConfig);
    setIsEditingSSO(false);
    if (ssoConfig) {
      setSelectedSsoItem(SSO_TYPES[ssoConfig.type]);
    }
  }, [ssoConfig]);

  const onSaveSSOSettings = useCallback(
    (id, config) => {
      const { contentType } = SSO_TYPES[selectedSsoItem.type];
      if (id) {
        return dispatch(changeSsoConfig({ config: { config, id }, contentType })).unwrap();
      }
      return dispatch(storeSsoConfig({ config, contentType })).unwrap();
    },
    [dispatch, selectedSsoItem]
  );

  const onCancelSSOSettings = () => {
    setIsEditingSSO(false);
    setIsConfiguringSSO(hasSingleSignOn);
    setSelectedSsoItem(ssoConfig ? SSO_TYPES[ssoConfig.type] : undefined);
  };

  const onDisableSSO = () =>
    dispatch(deleteSsoConfig(ssoConfig))
      .unwrap()
      .then(() => setSelectedSsoItem(undefined));

  const onTokenExpansion = useCallback(() => setShowTokenWarning(true), []);

  const onDownloadReportClick = () =>
    dispatch(downloadLicenseReport())
      .unwrap()
      .then(report => createFileDownload(report, `Mender-license-report-${dayjs().format('YYYY-MM-DD')}`, token));

  const onSSOClick = () => setIsConfiguringSSO(toggle);

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

  const ssoConfirmation = newSso ? ssoConfirmations.change : isDisablingSSO ? ssoConfirmations.disable : undefined;

  const onSsoConfirmationClose = () => {
    setNewSso('');
    setIsDisablingSSO(false);
  };

  return (
    <SettingsSection className={classes.widthLimit} title="Organization">
      <div className="flexbox column">
        <SettingsItem title="Organization ID" secondary={tenantId} sideBarContent={<CopyTextToClipboard notify={false} token={tenantId} />} />
        <SettingsItem title="Organization name" secondary={orgName} sideBarContent={<CopyTextToClipboard notify={false} token={orgName} />} />
        {canAddDevices && (
          <SettingsItem
            classes={{ base: '', content: '', main: classes.tenantToken }}
            title="Organization token"
            description="The token is unique for your organization and ensures that only devices that you own are able to connect to your account."
            secondary={<TenantToken expanded={showTokenWarning} onClick={onTokenExpansion} token={tenant_token} />}
            sideBarContent={<CopyTextToClipboard notify={false} onCopy={onTokenExpansion} token={tenant_token} />}
            notification={
              showTokenWarning && (
                <Alert severity="warning">
                  Do not share your organization token. If leaked, unauthorized devices can send authorization requests, making it difficult to identify and
                  manage your own devices.
                </Alert>
              )
            }
          />
        )}
        {isEnterprise && isAdmin && (
          <SettingsItem
            title={
              <div className="flexbox align-items-center gap-small">
                <Typography variant="subtitle1">Single Sign-On</Typography>
                {hasSingleSignOn && <Chip size="small" label="Enabled" variant="outlined" color="success" />}
              </div>
            }
            description="Allow users to sign in with your company’s identity provider."
            secondary={
              hasSingleSignOn && !isEditingSSO && selectedSsoItem ? (
                <div className="flexbox column gap-small">
                  <SSOConfig ssoItem={selectedSsoItem} config={ssoConfig} onSave={onSaveSSOSettings} readOnly />
                  <div className="flexbox gap-small">
                    <Button color="error" variant="outlined" onClick={() => setIsDisablingSSO(true)}>
                      Disable SSO
                    </Button>
                    <Button variant="outlined" color="info" onClick={() => setIsEditingSSO(true)}>
                      Edit
                    </Button>
                  </div>
                </div>
              ) : isConfiguringSSO ? (
                <div className="flexbox column gap-small">
                  <Select
                    displayEmpty
                    placeholder="Select type"
                    onChange={onSsoSelect}
                    options={Object.values(SSO_TYPES)}
                    renderOption={item => <div className="capitalized-start">{item.title}</div>}
                    selectionAttribute="type"
                    value={selectedSsoItem?.type || ''}
                    width={settingsFieldWidths.medium}
                  />
                  {selectedSsoItem && (
                    <Collapse in={isConfiguringSSO}>
                      <SSOConfig ssoItem={selectedSsoItem} config={ssoConfig} onSave={onSaveSSOSettings} />
                    </Collapse>
                  )}
                  <Button className="align-self-start" color="info" variant="outlined" onClick={onCancelSSOSettings}>
                    Cancel
                  </Button>
                </div>
              ) : (
                <Button onClick={onSSOClick}>Set up SSO</Button>
              )
            }
          />
        )}
        {(canPreview || !isHosted) && isEnterprise && isAdmin && (
          <SettingsItem
            title="License report"
            secondary={
              <Button onClick={onDownloadReportClick} endIcon={<SaveAltIcon />}>
                Download report
              </Button>
            }
          />
        )}
      </div>
      {ssoConfirmation && (
        <ConfirmModal
          {...ssoConfirmation}
          description="The current SSO configuration will be deleted, and users will not be able to sign in with your identity provider until SSO is set up again."
          toType="disable"
          open
          close={onSsoConfirmationClose}
          onConfirm={newSso ? changeSSO : onDisableSSO}
        />
      )}
    </SettingsSection>
  );
};

export default Organization;
