// Copyright 2020 Northern.tech AS
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
import { useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { BarChart as BarChartIcon } from '@mui/icons-material';
import { Typography } from '@mui/material';

import Loader from '@northern.tech/common-ui/Loader';
import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { MAX_PAGE_SIZE, TIMEOUTS, defaultReportType, defaultReports, rootfsImageVersion, softwareTitleMap } from '@northern.tech/store/constants';
import {
  getAcceptedDevices,
  getAttributesList,
  getDeviceReports,
  getDeviceReportsForUser,
  getGroupsByIdWithoutUngrouped,
  getIsEnterprise,
  getUserSettingsInitialized
} from '@northern.tech/store/selectors';
import { getDeviceAttributes, getReportDataWithoutBackendSupport, saveUserSettings } from '@northern.tech/store/thunks';
import { isEmpty } from '@northern.tech/utils/helpers';

import { extractSoftwareInformation } from '../devices/device-details/InstalledSoftware';
import BaseWidget from './widgets/BaseWidget';
import ChartAdditionWidget from './widgets/ChartAddition';
import DistributionReport from './widgets/Distribution';

const reportTypes = {
  distribution: DistributionReport
};

const getLayerKey = ({ title, key }, parent) => `${parent.length ? `${parent}.` : parent}${key.length <= title.length ? key : title}`;

const generateLayer = (softwareLayer, parentKey = '', nestingLevel = 0) => {
  const { children, key, title } = softwareLayer;
  const suffix = title === key ? '.version' : '';
  const layerKey = getLayerKey(softwareLayer, parentKey);
  const layerTitle = `${layerKey}${suffix}`;
  let headerItems = [{ title, nestingLevel, value: layerKey }];
  if (softwareTitleMap[layerTitle]) {
    headerItems = [
      { subheader: title, nestingLevel, value: `${layerTitle}-subheader` },
      { title: softwareTitleMap[layerTitle].title, nestingLevel: nestingLevel + 1, value: layerTitle }
    ];
  } else if (!isEmpty(children)) {
    headerItems = [{ subheader: title, nestingLevel, value: `${layerTitle}-subheader` }];
  }
  return Object.values(softwareLayer.children).reduce((accu, childLayer) => {
    const layerData = generateLayer(childLayer, getLayerKey(softwareLayer, parentKey), nestingLevel + 1);
    accu.push(...layerData);
    return accu;
  }, headerItems);
};

const listSoftware = attributes => {
  const enhancedAttributes = attributes.reduce((accu, attribute) => ({ ...accu, [attribute]: attribute }), {});
  const softwareTree = extractSoftwareInformation(enhancedAttributes, false);
  const { rootFs, remainder } = Object.values(softwareTree).reduce(
    (accu, layer) => {
      if (layer.key.startsWith('rootfs-image')) {
        return { ...accu, rootFs: layer };
      }
      accu.remainder.push(layer);
      return accu;
    },
    { rootFs: undefined, remainder: [] }
  );

  return (rootFs ? [rootFs, ...remainder] : remainder).flatMap(softwareLayer => generateLayer(softwareLayer));
};

const DeviceDataLimitWarning = () => (
  <div className="dashboard margin-bottom-large">
    <Typography variant="subtitle2">Device and Group Limit Exceeded</Typography>
    <Typography variant="caption">
      Your current number of devices and groups exceeds the limits of our present implementation. To ensure you continue to gain optimal insights and to better
      understand your specific requirements, we encourage you to reach out to <SupportLink variant="ourTeam" />. By providing us with more details about your
      use case, we can improve potential solutions to best accommodate your needs when the feature gets added to our backend.
    </Typography>
  </div>
);

const checkRequestLimitReached = (reports, deviceRetrievalLimit, total) => {
  const requestLimit = deviceRetrievalLimit / MAX_PAGE_SIZE;
  const { hasTooManyDevices } = reports.reduce(
    (accu, report) => {
      let { hasTooManyDevices, requestCounter } = accu;
      // as the attribute per report can be different for a given group or for all devices, count them both
      // + we assume smaller (sub 500 device) groups for now - the staggered widget rendering should allow some flexibility with the rate limits
      requestCounter += report.group ? 1 : Math.ceil(total / MAX_PAGE_SIZE);
      hasTooManyDevices = accu.hasTooManyDevices || accu.requestCounter > requestLimit;
      return { hasTooManyDevices, requestCounter };
    },
    { hasTooManyDevices: false, requestCounter: 0 }
  );
  return hasTooManyDevices;
};

export const SoftwareDistribution = () => {
  const reports = useSelector(getDeviceReportsForUser);
  const groups = useSelector(getGroupsByIdWithoutUngrouped);
  const attributes = useSelector(getAttributesList);
  const { total } = useSelector(getAcceptedDevices);
  const hasDevices = !!total;
  const isEnterprise = useSelector(getIsEnterprise);
  const hasUserSettingsInitialized = useSelector(getUserSettingsInitialized);
  const deviceRetrievalLimit = useSelector(state => state.deployments.deploymentDeviceLimit);
  const reportsData = useSelector(getDeviceReports);
  const hasReportsData = reportsData.reduce((accu, report) => accu && !isEmpty(report), true);
  const [visibleCount, setVisibleCount] = useState(hasReportsData ? reportsData.length : 1);
  const dispatch = useDispatch();
  const hasTooManyDevices = checkRequestLimitReached(reports, deviceRetrievalLimit, total);

  useEffect(() => {
    dispatch(getDeviceAttributes());
  }, [dispatch]);

  useEffect(() => {
    if (visibleCount < reports.length) {
      // this is purely to stagger the device retrieval and reduce overlap between the repeated queries to the backend
      const timeout = setTimeout(() => setVisibleCount(visibleCount + 1), TIMEOUTS.oneSecond);
      return () => clearTimeout(timeout);
    }
  }, [reports.length, visibleCount]);

  const addCurrentSelection = selection => {
    const newReports = [...reports, { ...defaultReports[0], ...selection }];
    dispatch(saveUserSettings({ reports: newReports }));
  };

  const onSaveChangedReport = (change, index) => {
    const newReports = [...reports];
    newReports.splice(index, 1, change);
    dispatch(saveUserSettings({ reports: newReports }));
  };

  const removeReport = removedReport => dispatch(saveUserSettings({ reports: reports.filter(report => report !== removedReport) }));

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const software = useMemo(() => listSoftware([rootfsImageVersion]), [JSON.stringify(attributes)]);

  if (!isEnterprise) {
    return (
      <div className="dashboard margin-bottom-large">
        <ChartAdditionWidget groups={groups} onAdditionClick={addCurrentSelection} software={software} />
      </div>
    );
  }
  if (hasTooManyDevices) {
    return <DeviceDataLimitWarning />;
  }
  if (!hasUserSettingsInitialized) {
    return (
      <div className="dashboard margin-bottom-large">
        <BaseWidget className="chart-widget flexbox centered" main={<Loader show style={{ width: '100%' }} />} />
      </div>
    );
  }
  return hasDevices ? (
    <div className="dashboard margin-bottom-large">
      {reports.slice(0, visibleCount).map((report, index) => {
        const Component = reportTypes[report.type || defaultReportType];
        return (
          <Component
            key={`report-${report.group}-${index}`}
            onClick={() => removeReport(report)}
            onSave={change => onSaveChangedReport(change, index)}
            selection={{ ...report, index }}
            software={software}
          />
        );
      })}
      {visibleCount < reports.length && (
        <div className="widget chart-widget flexbox centered">
          <Loader show style={{ width: '100%' }} />
        </div>
      )}
      <ChartAdditionWidget groups={groups} onAdditionClick={addCurrentSelection} software={software} />
    </div>
  ) : (
    <div className="dashboard-placeholder margin-top-large">
      <BarChartIcon style={{ transform: 'scale(5)' }} />
      <p className="margin-top-large">Software distribution charts will appear here once you connected a device. </p>
    </div>
  );
};

export default SoftwareDistribution;
