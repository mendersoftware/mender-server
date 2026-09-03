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
import { Collapse, IconButton, Table, TableBody, TableCell, TableHead, TableRow, Typography, tableCellClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { InventoryTable } from '@northern.tech/common-ui/InventoryTable';
import { LastChangedNote } from '@northern.tech/common-ui/LastChangedNote';
import { Link } from '@northern.tech/common-ui/Link';
import Pagination from '@northern.tech/common-ui/Pagination';
import { ControlledSearch } from '@northern.tech/common-ui/Search';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import storeActions from '@northern.tech/store/actions';
import { DEVICE_LIST_DEFAULTS, manifestVersion } from '@northern.tech/store/constants';
import { formatReleases, generateReleasesPath } from '@northern.tech/store/locationutils';
import { useAppDispatch } from '@northern.tech/store/store';
import { getDeviceComponents } from '@northern.tech/store/thunks';
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
  { title: 'Component Type', prop: 'component_type' },
  { title: 'Version', prop: 'version' }
];
const ComponentRow = props => {
  const [open, setOpen] = useState(false);
  const { className, component, onCopy } = props;
  return (
    <>
      <TableRow className={className}>
        <TableCell>
          <IconButton aria-label="expand row" size="small" onClick={() => setOpen(!open)}>
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        {columnsConfig.map(column => (
          <TableCell key={component.id + column.prop}>{component[column.prop]}</TableCell>
        ))}
      </TableRow>
      <TableRow>
        <TableCell className="padding-none" colSpan={columnsConfig.length + 1}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Typography variant="subtitle1">Component inventory</Typography>
            <InventoryTable config={component.attributes} setSnackbar={onCopy} />
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
};
const ComponentTable = props => {
  const { components } = props;
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
        </TableRow>
      </TableHead>
      <TableBody>
        {components.map(component => (
          <ComponentRow key={component.id} className={classes.componentRow} component={component} onCopy={onCopy} />
        ))}
      </TableBody>
    </Table>
  );
};

export const DeviceSystem = ({ device }) => {
  const { attributes = {}, updated_ts: updateTime, isOffline } = device;
  const { device_type: deviceTypes = [] } = attributes;
  const dispatch = useAppDispatch();
  const methods = useForm({ mode: 'onChange', defaultValues: { search: '' } });
  const searchTerm = useWatch({ control: methods.control, name: 'search' }).trim();
  const search = searchTerm.toLowerCase();
  const [page, setPage] = useState(defaultPage);
  const [perPage, setPerPage] = useState(defaultPerPage);
  const manifestName = attributes[manifestVersion];
  const systemType = deviceTypes.join(',') || '-';
  const manifestPath = `${generateReleasesPath({ pageState: {} })}?${formatReleases({ pageState: { tab: 'manifests', id: manifestName } })}`;
  const manifest = manifestName ? (
    <Link target="_blank" rel="noopener noreferrer" to={manifestPath}>
      {manifestName}
    </Link>
  ) : (
    '-'
  );
  useEffect(() => {
    dispatch(getDeviceComponents(device.id));
  }, [dispatch, device.id]);

  useEffect(() => {
    setPage(defaultPage);
  }, [search]);

  const filteredComponents = Object.entries(device.components ?? {})
    .map(toComponent)
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
            <ComponentTable components={paginatedComponents} />
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
