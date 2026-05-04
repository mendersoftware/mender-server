// Copyright 2024 Northern.tech AS
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
import { useFormContext } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { ErrorOutlined as ErrorOutlineIcon } from '@mui/icons-material';
import { Alert, Checkbox, FormControlLabel, FormHelperText, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import InfoHint from '@northern.tech/common-ui/InfoHint';
import { Link } from '@northern.tech/common-ui/Link';
import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import Form from '@northern.tech/common-ui/forms/Form';
import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import NumberInput from '@northern.tech/common-ui/forms/NumberInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { TIMEOUTS, rolesByName } from '@northern.tech/store/constants';
import { getSpLimits, getSsoConfig } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { addTenant, checkEmailExists, getSsoConfigs } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import { HELPTOOLTIPS } from '../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../helptips/MenderTooltip';

const useStyles = makeStyles()(theme => ({
  buttonWrapper: {
    '&.button-wrapper': {
      justifyContent: 'start',
      marginTop: 0
    }
  },
  formWrapper: {
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(5),
    maxWidth: '726px'
  },
  limitsContainer: {
    marginBottom: theme.spacing(2),
    maxWidth: '550px'
  }
}));

interface UserInputsProps {
  adminExists: boolean;
  checkEmailExists: (email: string) => Promise<void>;
}
export const convertToBackendLimits = (limits, spLimits) =>
  Object.fromEntries(
    Object.entries(limits)
      .filter(([key]) => !!spLimits[key])
      .map(([key, limit]) => {
        const { backendId } = spLimits[key];
        return [backendId, { Name: backendId, value: Number(limit) || 0 }];
      })
  );
const userExistsInfo =
  'This user already has a Mender account, and will be assigned as admin to the new tenant. If you want to create a brand new user, try a different email address.';
const newUserInfo = 'This will create a new user as admin of the new tenant.';

const UserInputs = (props: UserInputsProps) => {
  const { checkEmailExists, adminExists } = props;
  const [emailInfoText, setEmailInfoText] = useState<string>('');

  const { watch, getFieldState } = useFormContext();

  const enteredEmail = watch('email');
  const debouncedEmail = useDebounce(enteredEmail, TIMEOUTS.debounceDefault);

  useEffect(() => {
    const { invalid: isInvalidEmail } = getFieldState('email');
    if (!debouncedEmail || isInvalidEmail) {
      return;
    }
    checkEmailExists(debouncedEmail);
  }, [debouncedEmail, getFieldState, checkEmailExists]);

  useEffect(() => {
    const { invalid: isInvalidEmail } = getFieldState('email');
    if (!debouncedEmail || isInvalidEmail) {
      return;
    }
    if (adminExists) {
      setEmailInfoText(userExistsInfo);
      return;
    }
    setEmailInfoText(newUserInfo);
  }, [debouncedEmail, getFieldState, adminExists]);

  return (
    <>
      <TextInput
        validations="isEmail,trim"
        required
        requiredRendered={false}
        id="email"
        label="Admin user"
        width={430}
        helperText="This user will have the admin role for this tenant. It can be an existing Mender account, or a brand new user."
      />
      {emailInfoText ? <InfoHint content={emailInfoText} /> : <div />}
    </>
  );
};

export const DeviceLimitsInput = props => {
  const { spLimits, currentLimits, isEdit = false } = props;

  const [deviceTierEnabled, setDeviceTierEnabled] = useState(Object.fromEntries(Object.keys(spLimits).map(limit => [limit, isEdit])));
  const { classes } = useStyles();
  const { setFocus, resetField } = useFormContext();

  const inputProps = Object.values(spLimits).map(limit => {
    const unlimited = limit.limit === -1;
    const quotaLeft = limit.quotaLeft + currentLimits[limit.id];
    const rules = unlimited
      ? {}
      : {
          min: { value: 0, message: 'The limit must be 0 or more' },
          max: { value: quotaLeft, message: `The device limit must be ${quotaLeft} or fewer` }
        };
    return {
      ...limit,
      rules,
      maxPlaceholder: unlimited ? 'Device limit' : `Maximum: ${(quotaLeft || 0).toLocaleString()}`
    };
  });

  const onToggleDeviceTier = (id: string) => {
    const isEnabled = !deviceTierEnabled[id];
    if (isEnabled) {
      setTimeout(() => setFocus(id), 0);
    } else {
      resetField(id);
    }
    setDeviceTierEnabled(tiersEnabled => ({ ...tiersEnabled, [id]: !tiersEnabled[id] }));
  };
  return (
    <div>
      {inputProps.map(limit => (
        <div key={limit.id} className={classes.limitsContainer}>
          <FormControlLabel
            control={
              <Checkbox
                className="margin-left-x-small"
                disabled={limit.limit !== -1 && limit.current >= limit.limit && !isEdit}
                checked={deviceTierEnabled[limit.id]}
                onChange={() => onToggleDeviceTier(limit.id)}
              />
            }
            label={
              <div className="flexbox">
                <Typography color="textPrimary" className="capitalized-start">
                  {limit.name} devices
                </Typography>
                <MenderHelpTooltip id={HELPTOOLTIPS[`${limit.id}Device`].id} className="margin-left-small" />
              </div>
            }
          />
          <NumberInput
            className="margin-top-x-small"
            label="Device limit"
            disabled={!deviceTierEnabled[limit.id]}
            id={limit.id}
            required={deviceTierEnabled[limit.id]}
            requiredRendered={false}
            min={0}
            max={limit.quotaLeft}
            size="small"
            rules={limit.rules}
            helperText={deviceTierEnabled[limit.id] ? limit.maxPlaceholder : ''}
            width="550px"
          />
          {limit.limitReached && !isEdit && (
            <FormHelperText className="margin-left-small">You have already allocated your overall limit of standard devices.</FormHelperText>
          )}
        </div>
      ))}
      <Typography variant="body2">
        To increase your overall device limits{inputProps.length === 1 ? ' or add more device tiers,' : ','} <SupportLink variant="us" />
      </Typography>
    </div>
  );
};
const tenantAdminDefaults = { email: '', name: '', sso: false };

interface TenantCreateFormProps {
  onCloseClick: () => void;
  open: boolean;
}
export const TenantCreateForm = (props: TenantCreateFormProps) => {
  const { onCloseClick, open } = props;
  const [adminExists, setAdminExists] = useState<boolean>(false);
  const [hasError, setHasError] = useState<boolean>(false);
  const ssoConfig = useSelector(getSsoConfig);
  const spLimits = useSelector(getSpLimits);
  const dispatch = useAppDispatch();

  const { classes } = useStyles();
  const tenantLimitsDefaults = Object.fromEntries(Object.keys(spLimits).map(limit => [limit, 0]));

  const formInitialValues = { ...tenantAdminDefaults, ...tenantLimitsDefaults };

  useEffect(() => {
    dispatch(getSsoConfigs());
  }, [dispatch]);

  const onCheckEmailExists = useCallback(
    async (email: string) => {
      const exists = await dispatch(checkEmailExists(email)).unwrap();
      setAdminExists(exists);
    },
    [dispatch]
  );

  const submitNewTenant = useCallback(
    async data => {
      const { email, sso, name, ...deviceLimits } = data;
      let selectionState = {
        deviceLimits: deviceLimits,
        sso,
        name,
        binary_delta: true
      };
      if (adminExists) {
        selectionState = { users: [{ role: rolesByName.admin, email }], ...selectionState };
      } else {
        selectionState = { admin: { email, send_reset_password: true }, ...selectionState };
      }
      try {
        await dispatch(addTenant(selectionState)).unwrap(); // only awaiting the thunk resolution to not get rejected
        onCloseClick();
      } catch {
        setHasError(true);
      }
    },
    [adminExists, dispatch, onCloseClick]
  );

  return (
    <BaseDrawer open={open} onClose={onCloseClick} size="lg" slotProps={{ header: { title: 'Create a tenant' } }}>
      <Form
        initialValues={formInitialValues}
        classes={classes}
        className={classes.formWrapper}
        handleCancel={() => onCloseClick()}
        showButtons
        buttonColor="secondary"
        onSubmit={submitNewTenant}
        validationMode="onSubmit"
        submitLabel="Create tenant"
        autocomplete="off"
      >
        {hasError && (
          <Alert icon={<ErrorOutlineIcon />} severity="error">
            There was an error while creating the tenant. Please try again, or contact support.
          </Alert>
        )}
        <div>
          <Typography className="margin-bottom-x-small" variant="subtitle1">
            Tenant name
          </Typography>
          <TextInput validations="isLength:3:256,trim" required requiredRendered={false} id="name" hint="Name" label="Name" width={430} />
        </div>
        <div>
          <Typography className="margin-bottom-x-small" variant="subtitle1">
            Admin user email
          </Typography>
          <UserInputs adminExists={adminExists} checkEmailExists={onCheckEmailExists} />
        </div>
        <div>
          <Typography className="margin-bottom-x-small" variant="subtitle1">
            Device limits
          </Typography>
          <DeviceLimitsInput spLimits={spLimits} currentLimits={tenantLimitsDefaults} />
        </div>
        {!!ssoConfig && (
          <div>
            <Typography variant="subtitle1">Single Sign On</Typography>
            <Typography className="margin-top-x-small" variant="body2">
              Inherit the Single Sign On (SSO) configuration from the Service Provider. The created tenant’s admin user will not be able to change these
              settings later.{' '}
            </Typography>
            <Link target="_blank" to="/settings/organization" color="inherit">
              View your SSO settings{' '}
            </Link>
            <FormCheckbox className="margin-top-x-small margin-left-small" id="sso" label="Use Service Provider's SSO settings" />
          </div>
        )}
      </Form>
    </BaseDrawer>
  );
};
