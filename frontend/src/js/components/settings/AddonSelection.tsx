// Copyright 2021 Northern.tech AS
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
import React, { Fragment, useMemo, useState } from 'react';

import { Launch as LaunchIcon } from '@mui/icons-material';
import { Button, Chip } from '@mui/material';
import { buttonClasses } from '@mui/material/Button';
import { makeStyles } from 'tss-react/mui';

import { ConfirmAddon } from '@northern.tech/common-ui/dialogs/ConfirmAddon';
import { ADDONS, Addon, AvailablePlans, PLANS } from '@northern.tech/store/constants';
import { Organization } from '@northern.tech/store/organizationSlice/types';
import { useAppDispatch } from '@northern.tech/store/store';
import { requestPlanChange } from '@northern.tech/store/thunks';

import { addOnsToString, clearLocalStorageEntry, updateLocalStorage } from './Upgrade';

const useStyles = makeStyles()(theme => ({
  chip: {
    textTransform: 'uppercase',
    fontWeight: 'bold'
  },
  planPanelAddonItem: {
    borderBottom: `1px solid ${theme.palette.grey[300]}`
  },
  chipContainer: {
    paddingTop: '4px'
  },
  price: {
    paddingTop: '5px'
  },
  planPanelAddon: {
    display: 'grid',
    gridTemplateColumns: 'max-content 1fr max-content max-content',
    gridTemplateRows: 'repeat(3, minmax(35px, auto))'
  },
  button: {
    textTransform: 'none',
    padding: '2px 5px',
    [`&.${buttonClasses.text}.addButton`]: {
      color: theme.palette.primary.main
    },
    [`&.${buttonClasses.text}.removeButton`]: {
      color: theme.palette.secondary.lighter,
      backgroundColor: theme.palette.background.default
    }
  },
  placeholder: {
    width: '140px'
  }
}));
interface RelatableAddon extends Addon {
  name: string;
  isEnabled: boolean;
  pending: { isAdd: string; name: string };
  isEligible: boolean;
}
interface AddOnSelectionProps {
  org: Organization;
  currentPlan: AvailablePlans;
  addons: { enabled: boolean; name: string }[];
  features: string[];
  isTrial: boolean;
}
export const AddOnSelection = ({ org, currentPlan, addons = [], features, isTrial }: AddOnSelectionProps) => {
  const [action, setAction] = useState<{ name: string; isAdd: boolean } | null>(null);
  const currentPlanName = PLANS[currentPlan].name;
  const { classes } = useStyles();
  const dispatch = useAppDispatch();
  const onAddOnClick = (name: string, isAdd: boolean) => {
    setAction({ name, isAdd });
  };
  const requestAddon = async () => {
    if (!action) return;
    const { name: addonName, isAdd } = action;
    let requested_addons = addOnsToString(org.addons).split(', ');
    if (isAdd) {
      requested_addons.push(addonName);
    } else {
      requested_addons = requested_addons.filter(addon => addon !== addonName);
    }
    try {
      await dispatch(
        requestPlanChange({
          tenantId: org.id,
          content: {
            current_plan: currentPlanName,
            requested_plan: currentPlanName,
            current_addons: addOnsToString(org.addons) || '-',
            requested_addons: requested_addons.filter(addon => !!addon).join(', ') || '-',
            user_message: ''
          }
        })
      ).unwrap();
    } catch (error) {
      console.error(error);
      return;
    }
    updateLocalStorage(org.id, addonName, isAdd);
    setAction(null);
  };
  const relevantAddons = useMemo(
    () => {
      const currentState = JSON.parse(localStorage.getItem(org.id + '_upgrades') || '{}');
      return Object.entries(ADDONS)
        .reduce((acc: RelatableAddon[], [addOnName, addOn]) => {
          const isEnabled = addons.some(orgAddOn => orgAddOn.enabled && addOnName === orgAddOn.name);
          let pending = currentState[addOnName];

          if (pending && pending.pending && pending.isAdd === isEnabled) {
            clearLocalStorageEntry(org.id, addOnName);
            pending = null;
          }
          acc.push({ ...addOn, name: addOnName, isEnabled, pending, isEligible: addOn.eligible.indexOf(currentPlan) > -1 });
          return acc;
        }, [])
        .sort((a, b) => a.name.localeCompare(b.name));
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(addons), JSON.stringify(features), action, org]
  );
  return (
    <>
      {action && <ConfirmAddon name={action.name} onClose={() => setAction(null)} onConfirm={() => requestAddon()} variant={action.isAdd ? 'add' : 'remove'} />}
      <h3 className="margin-top-large">Get more features with add-ons</h3>

      <div className="flexbox column">
        <p>
          Enhance your Mender plan with optional add-on packages – offering additional features to easily manage your software and devices over-the-air. Below
          you can request which add-ons you’d like to be included with your plan.
        </p>
        <div className={classes.planPanelAddon}>
          {relevantAddons.map(addOn => {
            const addonItemClasses = `${addOn.isEnabled ? 'active' : ''} ${addOn.isEligible ? '' : 'muted'} ${classes.planPanelAddonItem}`;
            return (
              <Fragment key={addOn.name}>
                <a
                  className={`flexbox center-aligned bold ${addonItemClasses}`}
                  href={`https://mender.io/pricing/add-ons/${addOn.name}`}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Mender {addOn.title}
                  <LaunchIcon className="link-color margin-left-x-small" fontSize="small" />
                </a>
                {currentPlan !== PLANS.enterprise.id ? (
                  <div className={`${classes.planPanelAddonItem} ${addonItemClasses} ${classes.price} padding-left-small`}>
                    {addOn.isEligible ? (
                      <>
                        starting <b>{addOn[currentPlan].price}</b>
                      </>
                    ) : (
                      <>not available on {currentPlanName} plan</>
                    )}
                  </div>
                ) : (
                  <div className={`${classes.planPanelAddonItem} ${addonItemClasses}`} />
                )}
                {addOn.isEnabled ? (
                  <>
                    <div className={`${addonItemClasses} ${classes.chipContainer}`}>
                      <Chip className={`${classes.chip} muted margin-right-small`} size="small" label="active" />
                    </div>
                    {addOn.pending && !addOn.pending.isAdd ? (
                      <div className={`${addonItemClasses} ${classes.chipContainer}`}>
                        <Chip label="removal pending" size="small" className={`${classes.chip} muted`} />
                      </div>
                    ) : (
                      !isTrial && (
                        <Button
                          className={`${classes.button} ${addonItemClasses} removeButton`}
                          variant="text"
                          disabled={!addOn.isEligible}
                          onClick={() => onAddOnClick(addOn.name, false)}
                        >
                          Remove from plan
                        </Button>
                      )
                    )}
                  </>
                ) : (
                  <>
                    {addOn.pending && addOn.pending.isAdd ? (
                      <div className={`${addonItemClasses} ${classes.chipContainer}`}>
                        <Chip className={`${classes.chip} muted`} size="small" label="pending" />
                      </div>
                    ) : (
                      <Button
                        className={`${classes.button} ${addonItemClasses} addButton`}
                        variant="text"
                        disabled={!addOn.isEligible}
                        onClick={() => (addOn.isEligible ? onAddOnClick(addOn.name, true) : () => false)}
                      >
                        Add to plan
                      </Button>
                    )}
                    <div className={`${classes.placeholder} ${addonItemClasses}`} />
                  </>
                )}
              </Fragment>
            );
          })}
        </div>
      </div>
    </>
  );
};

export default AddOnSelection;
