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
import { ReactElement } from 'react';
import { Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom';

import AuditLogs from '../components/auditlogs/AuditLogs';
import Dashboard from '../components/dashboard/Dashboard';
import Deployments from '../components/deployments/Deployments';
import Devices from '../components/devices/DeviceGroups';
import Help from '../components/help/Help';
import Login from '../components/login/Login';
import Password from '../components/login/Password';
import PasswordReset from '../components/login/PasswordReset';
import Signup from '../components/login/Signup';
import Releases from '../components/releases/Releases';
import Settings from '../components/settings/Settings';
import { TenantPage } from '../components/tenants/TenantPage';

type RouteConfig = { element: ReactElement; isPublic?: boolean; path: string; title: string };
type RouteConfigs = Record<string, RouteConfig>;

export const routeConfigs: RouteConfigs = {
  auditlog: { path: 'auditlog', element: <AuditLogs />, title: 'Audit log' },
  dashboard: { path: '', element: <Dashboard />, title: 'Dashboard' },
  deployments: { path: 'deployments', element: <Deployments />, title: 'Deployments' },
  devices: { path: 'devices', element: <Devices />, title: 'Devices' },
  help: { path: 'help', element: <Help />, title: 'Help & support' },
  login: { path: 'login', element: <Login />, title: 'Tenants', isPublic: true },
  password: { path: 'password', element: <Password />, title: 'Tenants', isPublic: true },
  passwordReset: { path: 'password/:secretHash', element: <PasswordReset />, title: 'Tenants' },
  releases: { path: 'releases', element: <Releases />, title: 'Releases' },
  settings: { path: 'settings', element: <Settings />, title: 'Settings' },
  signup: { path: 'signup', element: <Signup />, title: 'Tenants', isPublic: true },
  tenants: { path: 'tenants', element: <TenantPage />, title: 'Tenants' }
};

const publicRoutes: string[] = Object.values(routeConfigs).reduce((accu, { path, isPublic }) => (isPublic ? [...accu, `/${path}`] : accu), [] as string[]);

const LocationValidator = () => {
  const location = useLocation();

  if (publicRoutes.some(publicRoute => location.pathname.startsWith(publicRoute))) {
    window.location.replace('/ui/');
    return;
  }
  return <Outlet />;
};

export const PrivateRoutes = () => (
  <Routes>
    <Route element={<LocationValidator />}>
      <Route path={routeConfigs.auditlog.path} element={routeConfigs.auditlog.element} />
      <Route path={routeConfigs.devices.path} element={routeConfigs.devices.element}>
        <Route path=":status" element={null} />
      </Route>
      <Route path={routeConfigs.releases.path} element={routeConfigs.releases.element}>
        <Route path=":artifactVersion" element={null} />
      </Route>
      <Route path={routeConfigs.deployments.path} element={routeConfigs.deployments.element}>
        <Route path=":tab" element={null} />
      </Route>
      <Route path={routeConfigs.settings.path} element={routeConfigs.settings.element}>
        <Route path=":section" element={null} />
      </Route>
      <Route path={routeConfigs.help.path} element={routeConfigs.help.element}>
        <Route path=":section" element={null} />
      </Route>
      <Route path="*" element={routeConfigs.dashboard.element} />
    </Route>
  </Routes>
);

export const PrivateSPRoutes = () => (
  <Routes>
    <Route element={<LocationValidator />}>
      <Route path={routeConfigs.auditlog.path} element={routeConfigs.auditlog.element} />
      <Route path={routeConfigs.settings.path} element={routeConfigs.settings.element}>
        <Route path=":section" element={null} />
      </Route>
      <Route path={routeConfigs.help.path} element={routeConfigs.help.element}>
        <Route path=":section" element={null} />
      </Route>
      <Route path={routeConfigs.tenants.path} element={routeConfigs.tenants.element}>
        <Route path=":tenantId" element={null} />
      </Route>
      <Route path="*" element={<Navigate to={routeConfigs.tenants.path} replace />} />
    </Route>
  </Routes>
);

export const PublicRoutes = () => (
  <Routes>
    <Route path={routeConfigs.password.path} element={routeConfigs.password.element} />
    <Route path={routeConfigs.passwordReset.path} element={routeConfigs.passwordReset.element} />
    <Route path={routeConfigs.signup.path} element={routeConfigs.signup.element}>
      <Route path=":campaign" element={null} />
    </Route>
    <Route path="*" element={routeConfigs.login.element} />
  </Routes>
);
