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
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Navigate, useLocation, useParams } from 'react-router';

import { Launch as LaunchIcon } from '@mui/icons-material';
import { useTheme } from '@mui/material';

import LeftNav from '@northern.tech/common-ui/LeftNav';
import PageTitle from '@northern.tech/common-ui/PageTitle';
import { SettingsSection } from '@northern.tech/common-ui/SettingsSection';
import { getFeatures } from '@northern.tech/store/selectors';

import Downloads from './Downloads';
import GetStarted from './GettingStarted';
import MenderHub from './MenderHub';
import Support from './Support';

const components = {
  'get-started': {
    title: 'Getting started',
    component: GetStarted
  },
  downloads: {
    title: 'Downloads',
    component: Downloads,
    hosted: true
  },
  support: {
    title: 'Contact support',
    component: Support
  },
  'mender-hub': {
    title: 'Mender Hub',
    component: MenderHub
  },
  documentation: {
    title: 'Documentation',
    url: `https://docs.mender.io/`
  }
};

// build array of link list components
const eachRecursive = (obj, path, level, accu, isHosted, spacing) =>
  Object.entries(obj).reduce((bag, [key, value]) => {
    if (!isHosted && value.hosted) {
      return bag;
    }
    if (typeof value == 'object' && value !== null && key !== 'component') {
      const this_path = `${path}/${key}`;
      bag.push({
        title: value.title,
        level,
        path: this_path,
        hosted: value.hosted,
        style: { paddingLeft: `calc(${level} * ${spacing})` },
        exact: true,
        icon: value.url ? <LaunchIcon fontSize="small" /> : null,
        url: value.url ? value.url : ''
      });
      bag = eachRecursive(value, this_path, level + 1, bag, isHosted, spacing);
    }
    return bag;
  }, accu);

const helpPath = 'help/';
export const Help = () => {
  const spacing = useTheme().spacing(2);
  const [links, setLinks] = useState([]);
  const { pathname } = useLocation();
  const { section } = useParams();

  const { isHosted } = useSelector(getFeatures);

  useEffect(() => {
    // generate sidebar links
    setLinks(eachRecursive(components, '/help', 1, [], isHosted, spacing));
  }, [isHosted, spacing]);

  if (!section) {
    return <Navigate replace to="/help/get-started" />;
  }

  let ComponentToShow = GetStarted;
  let breadcrumbs = '';
  const routeParams = pathname.includes(helpPath) ? pathname.substring(pathname.indexOf(helpPath) + helpPath.length) : '';
  if (routeParams) {
    const splitsplat = routeParams.split('/');
    let copyOfComponents = components;

    for (let i = 0; i < splitsplat.length; i++) {
      if (i === splitsplat.length - 1) {
        ComponentToShow = copyOfComponents[splitsplat[i]].component;
      } else {
        copyOfComponents = copyOfComponents[splitsplat[i]];
      }
    }

    breadcrumbs = splitsplat[0] ? '  >  ' + components[splitsplat[0]].title : '';
    breadcrumbs = splitsplat[1] ? breadcrumbs + '  >  ' + components[splitsplat[0]][splitsplat[1]].title : breadcrumbs;
  }

  return (
    <div className="tab-container with-sub-panels" style={{ minHeight: '95%', paddingTop: 0 }}>
      <PageTitle segments={[components[section]?.title, 'Help & support']} />
      <LeftNav sections={[{ itemClass: 'helpNav', items: links, title: 'Help & support' }]} />
      <div className="rightFluid padding-right-large" style={{ borderLeft: 'none', paddingBottom: '15%' }}>
        <SettingsSection title={`Help & support ${breadcrumbs}`}>
          <ComponentToShow />
        </SettingsSection>
      </div>
    </div>
  );
};

export default Help;
