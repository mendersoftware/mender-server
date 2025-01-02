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
import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Navigate, useParams } from 'react-router-dom';

// material ui
import { Payment as PaymentIcon } from '@mui/icons-material';

import LeftNav from '@northern.tech/common-ui/left-nav';
import { TIMEOUTS, canAccess } from '@northern.tech/store/constants';
import { getCurrentUser, getFeatures, getOrganization, getTenantCapabilities, getUserCapabilities, getUserRoles } from '@northern.tech/store/selectors';
import { Elements } from '@stripe/react-stripe-js';

import Upgrade from './Upgrade';
import Global from './global';
import Integrations from './integrations';
import Organization from './organization/organization';
import { RoleManagement } from './role-management/RoleManagement';
import SelfUserManagement from './user-management/selfusermanagement';
import UserManagement from './user-management/usermanagement';

let stripePromise = null;

const sectionMap = {
  'global-settings': {
    component: Global,
    text: () => 'Global settings',
    canAccess: ({ organization: { service_provider }, userCapabilities: { canManageUsers } }) => !service_provider && canManageUsers
  },
  'my-profile': { component: SelfUserManagement, text: () => 'My profile', canAccess },
  'organization-and-billing': {
    component: Organization,
    text: () => 'Organization and billing',
    canAccess: ({ hasMultitenancy }) => hasMultitenancy
  },
  'user-management': {
    component: UserManagement,
    text: () => 'User management',
    canAccess: ({ userCapabilities: { canManageUsers } }) => canManageUsers
  },
  'role-management': {
    component: RoleManagement,
    text: () => 'Roles',
    canAccess: ({ currentUser, userRoles: { isAdmin } }) => currentUser && isAdmin
  },
  integrations: {
    component: Integrations,
    text: () => 'Integrations',
    canAccess: ({ organization: { service_provider }, userRoles: { isAdmin } }) => !service_provider && isAdmin
  },
  upgrade: {
    component: Upgrade,
    icon: <PaymentIcon />,
    text: ({ organization: { trial } }) => (trial ? 'Upgrade to a plan' : 'Upgrades and add-ons'),
    canAccess: ({ hasMultitenancy, organization: { service_provider } }) => !service_provider && hasMultitenancy
  }
};

export const Settings = () => {
  const currentUser = useSelector(getCurrentUser);
  const { hasMultitenancy } = useSelector(getFeatures);
  const organization = useSelector(getOrganization);
  const stripeAPIKey = useSelector(state => state.app.stripeAPIKey);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const userCapabilities = useSelector(getUserCapabilities);
  const userRoles = useSelector(getUserRoles);
  const [loadingFinished, setLoadingFinished] = useState(!stripeAPIKey);
  const { section: sectionParam } = useParams();

  useEffect(() => {
    // Make sure to call `loadStripe` outside of a component’s render to avoid recreating
    // the `Stripe` object on every render - but don't initialize twice.
    if (!stripePromise) {
      import(/* webpackChunkName: "stripe" */ '@stripe/stripe-js').then(({ loadStripe }) => {
        if (stripeAPIKey) {
          stripePromise = loadStripe(stripeAPIKey).finally(() => setLoadingFinished(true));
        }
      });
    } else {
      const notStripePromise = new Promise(resolve => setTimeout(resolve, TIMEOUTS.debounceDefault));
      Promise.race([stripePromise, notStripePromise]).then(result => setLoadingFinished(result !== notStripePromise));
    }
  }, [stripeAPIKey]);

  const checkDenyAccess = item =>
    currentUser.id && !item.canAccess({ currentUser, hasMultitenancy, organization, tenantCapabilities, userCapabilities, userRoles });

  const getCurrentSection = (sections, section = sectionParam) => {
    if (!sections.hasOwnProperty(section) || checkDenyAccess(sections[section])) {
      return;
    }
    return sections[section];
  };

  const links = Object.entries(sectionMap).reduce((accu, [key, item]) => {
    if (!checkDenyAccess(item)) {
      accu.push({
        path: `/settings/${key}`,
        icon: item.icon,
        title: item.text({ organization })
      });
    }
    return accu;
  }, []);

  const section = getCurrentSection(sectionMap, sectionParam);
  if (!section) {
    return <Navigate to="/settings/my-profile" replace />;
  }
  const Component = section.component;
  return (
    <div className="tab-container with-sub-panels" style={{ minHeight: '95%' }}>
      <LeftNav sections={[{ itemClass: 'settingsNav', items: links, title: 'Settings' }]} />
      <div className="rightFluid padding-right-large" style={{ paddingBottom: '15%' }}>
        {loadingFinished && (
          <Elements stripe={stripePromise}>
            <Component />
          </Elements>
        )}
      </div>
    </div>
  );
};

export default Settings;