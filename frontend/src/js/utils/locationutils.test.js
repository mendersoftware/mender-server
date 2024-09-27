// Copyright 2022 Northern.tech AS
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
import { AUDIT_LOGS_TYPES, DEPLOYMENT_ROUTES, DEPLOYMENT_STATES, DEVICE_STATES, UNGROUPED_GROUP, listDefaultsByState } from '@northern.tech/store/constants';

import {
  commonProcessor,
  formatAuditlogs,
  formatDeployments,
  formatDeviceSearch,
  formatPageState,
  formatReleases,
  generateDeploymentsPath,
  generateDevicePath,
  generateReleasesPath,
  parseAuditlogsQuery,
  parseDeploymentsQuery,
  parseDeviceQuery,
  parseReleasesQuery
} from './locationutils';

const today = new Date(new Date().setHours(0, 0, 0, 0)).toISOString();
const tonight = new Date(new Date().setHours(23, 59, 59, 999)).toISOString();

const sortDefaults = { sort: { direction: 'asc' } };

describe('locationutils', () => {
  describe('common', () => {
    it('uses working utilties - commonProcessor', () => {
      const startParams = new URLSearchParams('?perPage=234&id=123-324&open=true&sort=asc&issues=issueType1&issues=issueType2');
      const { pageState, params, sort } = commonProcessor(startParams);
      expect(sort).toEqual({ direction: 'asc' });
      expect(pageState).toEqual({ id: ['123-324'], selectedIssues: ['issueType1', 'issueType2'], open: true, perPage: 234 });
      expect(params.has('page')).not.toBeTruthy();
    });
    it('uses working utilities - formatPageState', () => {
      const search = formatPageState(
        { selectedId: 123, selectedIssues: ['1243', 'qweioqwei'], page: 1234, perPage: 1000, sort: { direction: 'desc', key: 'someKey' } },
        { defaults: sortDefaults }
      );
      expect(search).toEqual('sort=someKey:desc&page=1234&perPage=1000&id=123&issues=1243&issues=qweioqwei&open=true');
    });
  });
  describe('auditlog', () => {
    const startDate = new Date('2000-01-01').toISOString();
    it('uses working utilties - formatAuditlogs', () => {
      let search = formatAuditlogs(
        {
          pageState: {
            detail: 'testgroup',
            endDate: new Date().toISOString(),
            startDate,
            type: AUDIT_LOGS_TYPES[1],
            user: { id: 1 }
          }
        },
        { today, tonight }
      );
      expect(search).toEqual('objectId=testgroup&userId=1&objectType=deployment&endDate=2019-01-13&startDate=2000-01-01');
      search = formatAuditlogs(
        {
          pageState: {
            detail: 'testgroup',
            endDate: new Date().toISOString(),
            startDate,
            type: AUDIT_LOGS_TYPES[1].value,
            user: '1'
          }
        },
        { today, tonight }
      );
      expect(search).toEqual('objectId=testgroup&userId=1&objectType=deployment&endDate=2019-01-13&startDate=2000-01-01');
    });
    it('uses working utilities - parseAuditlogsQuery', () => {
      const result = parseAuditlogsQuery(new URLSearchParams('objectId=testgroup&userId=1&objectType=device&endDate=2019-01-13&startDate=2000-01-01'), {
        today,
        tonight
      });
      expect(result).toEqual({
        detail: 'testgroup',
        endDate: '2019-01-13T23:59:59.999',
        startDate: startDate.replace('Z', ''),
        type: AUDIT_LOGS_TYPES[2],
        user: '1'
      });
    });
  });

  describe('deployments', () => {
    describe('parseDeploymentsQuery', () => {
      const defaultArgs = { location: { pathname: '/deployments' }, today, tonight };
      it('works as expected', () => {
        const result = parseDeploymentsQuery(new URLSearchParams('?pending=1:50&inprogress=5'), {
          pageState: { state: DEPLOYMENT_ROUTES.active.key, id: ['testId'] },
          ...defaultArgs,
          location: { pathname: '/deployments/unknown' }
        });
        expect(result).toEqual({
          deploymentObject: {},
          general: { showCreationDialog: false, showReportDialog: false, state: DEPLOYMENT_ROUTES.active.key },
          [DEPLOYMENT_STATES.inprogress]: { page: 5 },
          [DEPLOYMENT_STATES.pending]: { page: 1, perPage: 50 }
        });
      });
      it('works as expected - pt2', () => {
        const result = parseDeploymentsQuery(new URLSearchParams('?type=configuration&search=someSearch'), {
          pageState: { state: DEPLOYMENT_ROUTES.finished.key, id: ['testId'] },
          ...defaultArgs,
          location: { pathname: '/deployments/unknownToo' }
        });
        expect(result).toEqual({
          deploymentObject: {},
          general: { showCreationDialog: false, showReportDialog: false, state: DEPLOYMENT_ROUTES.finished.key },
          [DEPLOYMENT_STATES.finished]: { startDate: '', endDate: tonight, search: 'someSearch', type: 'configuration' }
        });
      });
      it('works as expected - pt3', () => {
        const result = parseDeploymentsQuery(new URLSearchParams('?endDate=2020-05-02&startDate=2000-01-25'), {
          pageState: { state: DEPLOYMENT_ROUTES.finished.key, id: ['testId'] },
          ...defaultArgs,
          location: { pathname: '/deployments/done' }
        });
        expect(result).toEqual({
          deploymentObject: {},
          general: { showCreationDialog: false, showReportDialog: false, state: DEPLOYMENT_ROUTES.finished.key },
          [DEPLOYMENT_STATES.finished]: { startDate: '2000-01-25T00:00:00.000', endDate: '2020-05-02T23:59:59.999', search: '', type: '' }
        });
      });
      it('works as expected - pt4', () => {
        const result = parseDeploymentsQuery(new URLSearchParams('?perPage=60'), {
          pageState: { state: DEPLOYMENT_ROUTES.finished.key, id: ['testId'], perPage: 60 },
          ...defaultArgs,
          location: { pathname: '/deployments/scheduled' }
        });
        expect(result).toEqual({
          deploymentObject: {},
          general: { showCreationDialog: false, showReportDialog: false, state: DEPLOYMENT_ROUTES.scheduled.key },
          [DEPLOYMENT_STATES.scheduled]: { perPage: 60 }
        });
      });
      it('works with release triggered dialogs', () => {
        const result = parseDeploymentsQuery(new URLSearchParams('?release=somereleaseName'), {
          pageState: { state: DEPLOYMENT_ROUTES.active.key, id: ['testId'] },
          ...defaultArgs
        });
        expect(result).toEqual({
          deploymentObject: { release: 'somereleaseName' },
          general: { showCreationDialog: false, showReportDialog: false, state: DEPLOYMENT_ROUTES.active.key }
        });
      });
      it('works with device triggered dialogs', () => {
        const result = parseDeploymentsQuery(new URLSearchParams('?deviceId=someDevice'), {
          pageState: { state: DEPLOYMENT_ROUTES.active.key, id: ['testId'] },
          ...defaultArgs
        });
        expect(result).toEqual({
          deploymentObject: { devices: [{ id: 'someDevice' }] },
          general: { showCreationDialog: false, showReportDialog: false, state: DEPLOYMENT_ROUTES.active.key }
        });
      });
    });
    describe('formatDeployments', () => {
      it('on active', () => {
        const pageState = {
          general: { state: DEPLOYMENT_ROUTES.active.key },
          [DEPLOYMENT_STATES.pending]: { page: 2, perPage: 40 },
          [DEPLOYMENT_STATES.inprogress]: { perPage: 60 },
          selectedId: 'testId'
        };
        const pathname = generateDeploymentsPath({ pageState });
        const search = formatDeployments({ pageState }, { defaults: listDefaultsByState, today, tonight });
        expect(pathname).toEqual('/deployments/active');
        expect(search).toEqual('inprogress=1:60&pending=2:40');
      });

      it('on scheduled', () => {
        const pageState = {
          general: { state: DEPLOYMENT_ROUTES.scheduled.key },
          [DEPLOYMENT_STATES.scheduled]: { page: 2 }
        };
        const pathname = generateDeploymentsPath({ pageState });
        const search = formatDeployments(
          { pageState },
          {
            defaults: listDefaultsByState,
            today,
            tonight
          }
        );
        expect(pathname).toEqual('/deployments/scheduled');
        expect(search).toEqual('page=2');
      });

      it('on finished', () => {
        const pageState = {
          general: { state: DEPLOYMENT_ROUTES.finished.key },
          [DEPLOYMENT_STATES.finished]: { page: 4, search: 'something', type: 'deployment' }
        };
        const search = formatDeployments({ pageState }, { defaults: listDefaultsByState, today, tonight });
        const pathname = generateDeploymentsPath({ pageState });
        expect(pathname).toEqual('/deployments/finished');
        expect(search).toEqual('page=4&search=something&type=deployment');
      });
    });
  });

  describe('devices', () => {
    it('uses working utilties - parseDeviceQuery converts classic url style', () => {
      const { groupName, filters } = parseDeviceQuery(new URLSearchParams('?some=thing&group=testgroup&mac=donalds&existing=filter'), {
        filteringAttributes: { identityAttributes: ['mac'], inventoryAttributes: ['some'] }
      });
      expect(groupName).toEqual('testgroup');
      expect(filters).toEqual([
        { key: 'mac', operator: '$eq', scope: 'identity', value: 'donalds' },
        { key: 'some', operator: '$eq', scope: 'inventory', value: 'thing' },
        { key: 'existing', operator: '$eq', scope: 'inventory', value: 'filter' }
        // { key: 'existing', operator: '$wat', scope: 'special', value: 'filter' }
      ]);
    });
    it('uses working utilties - parseDeviceQuery converts new style', () => {
      const { groupName, filters } = parseDeviceQuery(new URLSearchParams('?inventory=some:eq:thing&inventory=group:eq:testgroup&identity=bla:ne:blubb'));
      expect(groupName).toEqual('testgroup');
      expect(filters).toEqual([
        { key: 'bla', operator: '$ne', scope: 'identity', value: 'blubb' },
        { key: 'some', operator: '$eq', scope: 'inventory', value: 'thing' }
      ]);
    });
    it('uses working utilties - parseDeviceQuery converts new style with device context', () => {
      const { open } = parseDeviceQuery(new URLSearchParams(), { pageState: { id: ['something'] } });
      expect(open).toEqual(true);
    });
    it('uses working utilties - parseDeviceQuery parses properly with many : in the query', () => {
      const { filters } = parseDeviceQuery(new URLSearchParams('?inventory=some:eq:thing:group:eq:testgroup:bla:ne:blubb'));
      expect(filters).toEqual([{ key: 'some', operator: '$eq', scope: 'inventory', value: 'thing:group:eq:testgroup:bla:ne:blubb' }]);
    });

    const devicesPath = '/devices/asd';
    const devicesSearch = '?some=thing&different=thing&entirely=different';
    it('uses working utilties - formatDeviceSearch', () => {
      const pageState = {
        pageState: { state: DEVICE_STATES.pending },
        filters: [{ key: 'some', value: 'thing' }],
        selectedGroup: 'testgroup',
        location: { pathname: devicesPath, search: devicesSearch }
      };
      const pathname = generateDevicePath(pageState);
      const search = formatDeviceSearch(pageState);
      expect(pathname).toEqual('/devices/pending');
      expect(search).toEqual('inventory=some:eq:thing&inventory=group:eq:testgroup');
    });

    it('uses working utilties - formatDeviceSearch - on init', () => {
      const pageState = {
        pageState: { state: DEVICE_STATES.accepted },
        filters: [{ key: 'some', value: 'thing' }],
        selectedGroup: 'testgroup',
        location: { pathname: devicesPath, search: devicesSearch }
      };
      const pathname = generateDevicePath(pageState);
      const search = formatDeviceSearch(pageState);
      expect(pathname).toEqual('/devices/accepted');
      expect(search).toEqual('inventory=some:eq:thing&inventory=group:eq:testgroup');
    });

    it('uses working utilties - formatDeviceSearch - with ungrouped selected', () => {
      const search = formatDeviceSearch({
        filters: [{ key: 'some', value: 'thing' }],
        pageState: {},
        selectedGroup: UNGROUPED_GROUP.id
      });
      expect(search).toEqual('inventory=some:eq:thing&inventory=group:eq:Unassigned');
    });
  });

  describe('releases', () => {
    it('uses working utilties - formatReleases', () => {
      let search = formatReleases({ pageState: { tab: '', selectedTags: [] } });
      expect(search).toEqual('');
      search = formatReleases({ pageState: { tab: 'flump', selectedTags: ['123', '456'] } });
      expect(search).toEqual('tag=123&tag=456&tab=flump');
    });
    it('uses working utilities - parseReleasesQuery', () => {
      const result = parseReleasesQuery(new URLSearchParams('tab=flump&tag=asd&tag=52534'), {
        location: { pathname: '/releases/terst' }
      });
      const endDate = new Date('2019-01-13');
      endDate.setHours(23, 59, 59, 999);
      expect(result).toEqual({
        searchTerm: '',
        selectedRelease: 'terst',
        tab: 'flump',
        tags: ['asd', '52534'],
        type: ''
      });
    });
    it('uses working utilities - generateReleasesPath', () => {
      const pageState = { tab: 'flump', selectedRelease: 'testId', selectedTags: ['123', '456'] };
      const pathname = generateReleasesPath({ pageState });
      const search = formatReleases({ pageState });
      expect(pathname).toEqual('/releases/testId');
      expect(search).toEqual('tag=123&tag=456&tab=flump');
    });
  });
});
