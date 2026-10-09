// Copyright 2026 Northern.tech AS
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
import { FormProvider, useForm, useWatch } from 'react-hook-form';

import { KeyboardArrowDown as KeyboardArrowDownIcon, KeyboardArrowUp as KeyboardArrowUpIcon } from '@mui/icons-material';
import { Chip, Collapse, IconButton, Table, TableBody, TableCell, TableHead, TableRow, Tooltip, Typography, tableCellClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { InventoryTable } from '@northern.tech/common-ui/InventoryTable';
import { LastChangedNote } from '@northern.tech/common-ui/LastChangedNote';
import { Link } from '@northern.tech/common-ui/Link';
import Pagination from '@northern.tech/common-ui/Pagination';
import { ControlledSearch } from '@northern.tech/common-ui/Search';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import storeActions from '@northern.tech/store/actions';
import { DEVICE_LIST_DEFAULTS } from '@northern.tech/store/constants';
import type { DeviceComponents } from '@northern.tech/store/devicesSlice';
import { formatReleases, generateReleasesPath } from '@northern.tech/store/locationutils';
import { useAppDispatch } from '@northern.tech/store/store';
import pluralize from 'pluralize';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(() => ({
  componentRow: {
    [`& > .${tableCellClasses.root}`]: {
      borderBottom: 'unset'
    }
  }
}));

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

const inconsistentSoftwareMarker = '_INCONSISTENT';

export const systemComponentAttribute = 'system_component';
export const orchestratorManifestAttribute = 'mender-orchestrator-manifest.manifest_name';

export const getSystemComponentId = (components: DeviceComponents = {}) =>
  Object.keys(components).find(id => components[id].some(({ name, value }) => name === systemComponentAttribute && value === 'true'));

const systemComponentTooltipBase = 'This component runs the Mender Orchestrator and reports to the Mender Server as the System device.';
const systemComponentTooltips = {
  inventory: `${systemComponentTooltipBase} See the inventory of the other components in the System tab.`,
  software: `${systemComponentTooltipBase} See the software of the other components in the System tab.`,
  system: `${systemComponentTooltipBase} Its own data is shown in the Inventory and Software tabs.`
};

type TooltipVariant = 'inventory' | 'software' | 'system';

export const SystemDeviceChip = ({ className = '', tooltipVariant }: { className?: string; tooltipVariant: TooltipVariant }) => (
  <Tooltip title={systemComponentTooltips[tooltipVariant]} placement="bottom">
    <Chip className={className} label="System device" size="small" variant="outlined" />
  </Tooltip>
);

export const SystemComponentNote = ({ componentId, variant }: { componentId: string; variant: TooltipVariant }) => (
  <div className="flexbox align-items-center margin-bottom-small">
    <Typography variant="body2">Component: {componentId}</Typography>
    <SystemDeviceChip className="margin-left-small" tooltipVariant={variant} />
  </div>
);

const visibleAttributes = ['component_type', 'version'];
const searchableAttributes = ['id', 'component_type'];
const toComponent = ([id, attributes]) =>
  attributes.reduce(
    (accu, { name, value }) => {
      if (visibleAttributes.includes(name)) {
        accu[name] = value;
      } else {
        accu.attributes[name] = value;
      }
      return accu;
    },
    { id, attributes: {} }
  );
const matchesSearch = (component, keys, search) => !search || keys.some(key => `${component[key] ?? ''}`.toLowerCase().includes(search));

const columnsConfig = [
  { title: 'ID', prop: 'id' },
  { title: 'Component type', prop: 'component_type' },
  { title: 'Software version', prop: 'version' }
];
const ComponentCells = ({ component, idSuffix = null }) =>
  columnsConfig.map(({ prop }) => (
    <TableCell className="flexbox align-items-center" key={component.id + prop}>
      {component[prop]}
      {prop === 'id' && idSuffix}
    </TableCell>
  ));

const SystemComponentRow = ({ className, component, setDetailsTab }) => (
  <TableRow className={className}>
    <TableCell />
    <ComponentCells component={component} idSuffix={<SystemDeviceChip className="margin-left-small" tooltipVariant="system" />} />
    <TableCell align="right">
      <Link onClick={() => setDetailsTab('inventory')}>Inventory</Link>
      <Link className="margin-left-small" onClick={() => setDetailsTab('software')}>
        Software
      </Link>
    </TableCell>
  </TableRow>
);

const ComponentRow = ({ className, component, onCopy }) => {
  const [open, setOpen] = useState(false);
  return (
    <>
      <TableRow aria-label="expand row" className={`clickable ${className}`} hover onClick={() => setOpen(!open)}>
        <TableCell>
          <IconButton size="small">{open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}</IconButton>
        </TableCell>
        <ComponentCells component={component} />
        <TableCell />
      </TableRow>
      <TableRow>
        <TableCell className="padding-none" colSpan={columnsConfig.length + 1}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Typography variant="subtitle1">Inventory</Typography>
            <InventoryTable config={component.attributes} setSnackbar={onCopy} />
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
};

const ComponentTable = ({ components, setDetailsTab, systemComponentId }) => {
  const { classes } = useStyles();
  const dispatch = useAppDispatch();
  const onCopy = useCallback(message => dispatch(setSnackbar(message)), [dispatch]);
  return (
    <Table>
      <TableHead>
        <TableRow>
          <TableCell />
          {columnsConfig.map(column => (
            <TableCell key={column.title}>{column.title}</TableCell>
          ))}
          <TableCell />
        </TableRow>
      </TableHead>
      <TableBody>
        {components.map(component =>
          component.id === systemComponentId ? (
            <SystemComponentRow key={component.id} className={classes.componentRow} component={component} setDetailsTab={setDetailsTab} />
          ) : (
            <ComponentRow key={component.id} className={classes.componentRow} component={component} onCopy={onCopy} />
          )
        )}
      </TableBody>
    </Table>
  );
};

export const DeviceSystem = ({ device, setDetailsTab }) => {
  const { attributes = {}, components = {}, updated_ts: updateTime, isOffline } = device;
  const { system_type: systemType } = attributes;
  const methods = useForm({ mode: 'onChange', defaultValues: { search: '' } });
  const searchTerm = useWatch({ control: methods.control, name: 'search' }).trim();
  const search = searchTerm.toLowerCase();
  const [page, setPage] = useState(defaultPage);
  const [perPage, setPerPage] = useState(defaultPerPage);
  const manifestName = attributes[orchestratorManifestAttribute];
  const manifestId = manifestName?.endsWith(inconsistentSoftwareMarker) ? manifestName.slice(0, -inconsistentSoftwareMarker.length) : manifestName;
  const manifestPath = `${generateReleasesPath({ pageState: {} })}?${formatReleases({ pageState: { tab: 'manifests', id: manifestId } })}`;
  const manifest = manifestName ? (
    <Link target="_blank" rel="noopener noreferrer" to={manifestPath}>
      {manifestName}
    </Link>
  ) : (
    '-'
  );

  useEffect(() => {
    setPage(defaultPage);
  }, [search]);

  const systemComponentId = getSystemComponentId(components);
  const filteredComponents = Object.entries(components)
    .map(toComponent)
    .sort((a, b) => Number(b.id === systemComponentId) - Number(a.id === systemComponentId))
    .filter(component => matchesSearch(component, searchableAttributes, search));
  const paginatedComponents = filteredComponents.slice((page - 1) * perPage, page * perPage);

  const onChangeRowsPerPage = newPerPage => {
    setPage(defaultPage);
    setPerPage(newPerPage);
  };

  return (
    <>
      <ContentSection title="System information">
        <TwoColumnData data={{ 'System type': systemType, Manifest: manifest }} />
      </ContentSection>
      <ContentSection title="Components" titleEnd={<LastChangedNote updateTime={updateTime} isOffline={isOffline} />}>
        <FormProvider {...methods}>
          <ControlledSearch asFormField placeholder="ID or component type" />
        </FormProvider>
        {!!searchTerm && (
          <Typography className="margin-top-small" variant="subtitle2">
            Showing {filteredComponents.length} {pluralize('result', filteredComponents.length)} for ‘{searchTerm}’
          </Typography>
        )}
        {filteredComponents.length === 0 && searchTerm ? (
          <div className="flexbox centered margin-top-small">
            <Typography variant="body1">No components were found. Try adjusting your search query</Typography>
          </div>
        ) : (
          <>
            <ComponentTable components={paginatedComponents} setDetailsTab={setDetailsTab} systemComponentId={systemComponentId} />
            {!!filteredComponents.length && (
              <div className="flexbox">
                <Pagination
                  key={filteredComponents.length}
                  className="margin-top-none"
                  count={filteredComponents.length}
                  rowsPerPage={perPage}
                  page={page}
                  onChangePage={setPage}
                  onChangeRowsPerPage={onChangeRowsPerPage}
                />
              </div>
            )}
          </>
        )}
      </ContentSection>
    </>
  );
};
