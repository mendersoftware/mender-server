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
import React from 'react';
import { Outlet, Route, Routes, useLocation } from 'react-router-dom';

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

const publicRoutes = ['/password', '/signup', '/login'];

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
      <Route path="auditlog" element={<AuditLogs />} />
      <Route path="devices" element={<Devices />}>
        <Route path=":status" element={null} />
      </Route>
      <Route path="releases" element={<Releases />}>
        <Route path=":artifactVersion" element={null} />
      </Route>
      <Route path="deployments" element={<Deployments />}>
        <Route path=":tab" element={null} />
      </Route>
      <Route path="settings" element={<Settings />}>
        <Route path=":section" element={null} />
      </Route>
      <Route path="help" element={<Help />}>
        <Route path=":section" element={null} />
      </Route>
      <Route path="*" element={<Dashboard />} />
    </Route>
  </Routes>
);
export const PrivateSPRoutes = () => {
  return (
    <Routes>
      <Route element={<LocationValidator />}>
        <Route path="auditlog" element={<AuditLogs />} />
        <Route path="settings" element={<Settings />}>
          <Route path=":section" element={null} />
        </Route>
        <Route path="help" element={<Help />}>
          <Route path=":section" element={null} />
        </Route>
        <Route path="*" element={<TenantPage />} />
      </Route>
    </Routes>
  );
};

export const PublicRoutes = () => (
  <Routes>
    <Route path="password" element={<Password />} />
    <Route path="password/:secretHash" element={<PasswordReset />} />
    <Route path="signup" element={<Signup />}>
      <Route path=":campaign" element={null} />
    </Route>
    <Route path="*" element={<Login />} />
  </Routes>
);
