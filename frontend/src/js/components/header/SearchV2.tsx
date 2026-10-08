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
import type { KeyboardEvent } from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router';

// material ui
import { ArrowForward as ArrowForwardIcon, Close as CloseIcon, DeveloperBoard as DeviceIcon, Search as SearchIcon } from '@mui/icons-material';
import {
  Alert,
  Breadcrumbs,
  ButtonBase,
  Dialog,
  Divider,
  IconButton,
  InputAdornment,
  InputBase,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  ListSubheader,
  Skeleton,
  Typography,
  breadcrumbsClasses,
  buttonBaseClasses,
  listItemIconClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { defaultTextRender, getDeviceIdentityText } from '@northern.tech/common-ui/DeviceIdentity';
import { ApproximateRelativeDate } from '@northern.tech/common-ui/Time';
import Select from '@northern.tech/common-ui/forms/Select';
import { ALL_DEVICE_STATES, DEVICE_FILTERING_OPTIONS, DEVICE_STATES, TIMEOUTS } from '@northern.tech/store/constants';
import type { Device } from '@northern.tech/store/devicesSlice';
import { formatDeviceSearch } from '@northern.tech/store/locationutils';
import { getDeviceIdentityAttributes, getIdAttribute, getUserSettings } from '@northern.tech/store/selectors';
import { useAppDispatch, useAppSelector } from '@northern.tech/store/store';
import { saveUserSettings, searchIdentities, setDeviceListState } from '@northern.tech/store/thunks';
import { isDarkMode } from '@northern.tech/store/utils';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import { getDeviceSoftwareText } from '../devices/BaseDevices';
import DeviceStatus from '../devices/DeviceStatus';

const triggerPlaceholder = 'Find a device';
const shortcut = '/';
const editableTags = ['INPUT', 'SELECT', 'TEXTAREA'];
const resultsPerPage = 10;
const skeletonRows = Array.from({ length: 4 }, (_, index) => index);
const navigationOffsets = { ArrowDown: 1, ArrowUp: -1 };

interface SearchResultItem {
  checkIn?: string;
  device: Device;
  label: string;
  metadata: string[];
  value: string;
}

const toOptionKey = ({ attribute, scope }) => `${scope}:${attribute}`;

const findOption = (options, { attribute, scope }) => options.find(option => option.attribute === attribute && option.scope === scope);

const toResult = (device: Device, idAttribute, { attribute, scope }): SearchResultItem => ({
  checkIn: device.check_in_time_exact ?? device.check_in_time_rounded,
  device,
  label: getDeviceIdentityText({ device, idAttribute }),
  metadata: [device.attributes.device_type?.join(', '), getDeviceSoftwareText(device.attributes)].filter(Boolean),
  value: String(defaultTextRender({ column: { attribute: { name: attribute, scope } }, device }))
});

const useStyles = makeStyles()(theme => ({
  emptyState: { minHeight: 200 },
  highlight: { backgroundColor: theme.palette.highlight?.main },
  inlineTime: { display: 'inline', fontSize: 'inherit' },
  italic: { fontStyle: 'italic' },
  inputPlaceholder: { '&::placeholder': { color: theme.palette.text.secondary, opacity: 1 } },
  subheader: { backgroundColor: 'transparent' },
  viewAll: { background: theme.palette.action.hover },
  metadata: {
    [`&.${breadcrumbsClasses.root}`]: { color: 'inherit', font: 'inherit', letterSpacing: 'inherit' },
    [`& .${breadcrumbsClasses.separator}`]: { marginInline: theme.spacing(0.5) }
  },
  listItemIcon: {
    [`&.${listItemIconClasses.root}`]: { alignSelf: 'flex-start', marginTop: theme.spacing(1), minWidth: 'auto' }
  },
  chipSelect: {
    borderRadius: 16,
    minWidth: 'unset',
    border: `1px solid ${theme.palette.action.selected}`,
    '&:hover': { backgroundColor: theme.palette.action.focus },
    '&&': {
      '& .MuiSelect-select': {
        minHeight: 'unset',
        padding: '1px 25px 1px 12px',
        borderRadius: 16
      }
    },
    '& .MuiOutlinedInput-notchedOutline': { border: 0 },
    '& .MuiSelect-icon': { right: 6, fontSize: 18 }
  },
  shortcut: {
    border: `1px solid ${theme.palette.divider}`,
    background: isDarkMode(theme.palette.mode) ? theme.palette.divider : theme.palette.grey[100],
    borderRadius: theme.shape.borderRadius,
    flexShrink: 0,
    fontFamily: theme.typography.fontFamily,
    fontSize: theme.typography.pxToRem(11),
    lineHeight: 2.2,
    minWidth: 24,
    textAlign: 'center'
  },
  trigger: {
    border: `1px solid ${theme.palette.divider}`,
    borderRadius: theme.shape.borderRadius,
    color: theme.palette.text.secondary,
    maxWidth: 400,
    [`&.${buttonBaseClasses.root}`]: { justifyContent: 'start', padding: theme.spacing(1, 1.5) },
    '&:hover': { borderColor: theme.palette.text.secondary },
    [`&.${buttonBaseClasses.focusVisible}`]: {
      borderColor: theme.palette.primary.main,
      boxShadow: `0 0 0 1px ${theme.palette.primary.main}`
    }
  },
  triggerLabel: { flexGrow: 1, textAlign: 'start' }
}));

const isEditableTarget = (target: EventTarget | null) => {
  const element = target as HTMLInputElement | null;
  return !!element && (element.isContentEditable || (editableTags.includes(element.tagName) && !element.readOnly));
};

const HighlightedMatch = ({ className, term, text }) => {
  const index = term ? text.indexOf(term) : -1;
  if (index < 0) {
    return text;
  }
  return (
    <>
      {text.slice(0, index)}
      <mark className={className}>{text.slice(index, index + term.length)}</mark>
      {text.slice(index + term.length)}
    </>
  );
};

const SearchAdornment = () => (
  <InputAdornment position="start">
    <SearchIcon color="inherit" fontSize="small" />
  </InputAdornment>
);

const ResultsSkeleton = () => {
  const { classes } = useStyles();
  return (
    <List disablePadding>
      {skeletonRows.map(row => (
        <ListItem key={row}>
          <ListItemIcon className={`margin-right-x-small ${classes.listItemIcon}`}>
            <Skeleton height={20} variant="circular" width={20} />
          </ListItemIcon>
          <ListItemText primary={<Skeleton width="25%" />} secondary={<Skeleton width="40%" />} />
        </ListItem>
      ))}
    </List>
  );
};

const SearchTrigger = ({ className, onOpen }) => {
  const { classes } = useStyles();
  return (
    <ButtonBase
      aria-haspopup="dialog"
      aria-keyshortcuts={shortcut}
      aria-label={triggerPlaceholder}
      className={`full-width ${classes.trigger} ${className}`}
      onClick={onOpen}
    >
      <SearchIcon className="margin-right-x-small" color="inherit" fontSize="small" />
      <Typography className={classes.triggerLabel} variant="body1">
        {triggerPlaceholder}
      </Typography>
      <kbd className={classes.shortcut}>{shortcut}</kbd>
    </ButtonBase>
  );
};

const SearchDialog = ({ onClose, open }) => {
  const [activeIndex, setActiveIndex] = useState(0);
  const [{ results, term: searchedTerm, total }, setSearch] = useState<{ results: SearchResultItem[]; term: string; total: number }>({
    results: [],
    term: '',
    total: 0
  });
  const [term, setTerm] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const idAttribute = useAppSelector(getIdAttribute);
  const { searchHintDismissed, searchAttribute } = useAppSelector(getUserSettings);
  const identityAttributes = useAppSelector(getDeviceIdentityAttributes);
  const [pickedAttribute, setPickedAttribute] = useState<{ attribute: string; scope: string } | null>(null);

  const attributeOptions = useMemo(
    () => identityAttributes.map(({ label, scope, value }) => ({ attribute: value, key: toOptionKey({ attribute: value, scope }), label, scope })),
    [identityAttributes]
  );
  const selectedOption =
    findOption(attributeOptions, pickedAttribute ?? searchAttribute ?? idAttribute) ?? findOption(attributeOptions, idAttribute) ?? attributeOptions[0];
  const { attribute, scope } = selectedOption;

  const dispatch = useAppDispatch();
  const { classes } = useStyles();
  const debouncedTerm = useDebounce(term, TIMEOUTS.debounceDefault);

  const placeholder = `Find devices by ${selectedOption.label} starting with...`;
  const isIdAttributeSelected = idAttribute.scope === scope && idAttribute.attribute === attribute;
  const hasSearched = !!searchedTerm;
  const showResults = !!results.length;
  const showPlaceholder = !showResults && !!term;
  const showSkeleton = showPlaceholder && !hasSearched;
  const showEmptyState = showPlaceholder && hasSearched;
  const showViewAll = total > results.length;
  const optionCount = results.length + (showViewAll ? 1 : 0);

  useEffect(() => {
    if (!debouncedTerm) {
      setSearch({ results: [], term: '', total: 0 });
      return;
    }
    let isCurrent = true;
    const searchedAttribute = { attribute, scope };
    dispatch(
      searchIdentities({
        attributes: [idAttribute, searchedAttribute],
        name: attribute,
        per_page: resultsPerPage,
        scope,
        value_prefix: debouncedTerm
      })
    )
      .unwrap()
      .then(({ devices, total }) => {
        if (!isCurrent) {
          return;
        }
        setSearch({ results: devices.map(device => toResult(device, idAttribute, searchedAttribute)), term: debouncedTerm, total });
      })
      .catch(() => isCurrent && setSearch({ results: [], term: debouncedTerm, total: 0 }));
    return () => {
      isCurrent = false;
    };
  }, [attribute, debouncedTerm, dispatch, idAttribute, scope]);

  useEffect(() => {
    setActiveIndex(0);
  }, [results]);

  const onDismissHint = () => dispatch(saveUserSettings({ searchHintDismissed: true }));

  const closeAndNavigate = (to, options?) => {
    onClose();
    setTimeout(() => navigate(to, options), TIMEOUTS.debounceShort);
  };

  const onChangeAttribute = ({ target: { value } }) => {
    const option = attributeOptions.find(({ key }) => key === value);
    if (!option) {
      return;
    }
    const nextAttribute = { attribute: option.attribute, scope: option.scope };
    setSearch({ results: [], term: '', total: 0 });
    setPickedAttribute(nextAttribute);
    dispatch(saveUserSettings({ searchAttribute: nextAttribute }));
  };

  const onSelect = ({ id, status }: Device) => {
    const deviceState = Object.values(DEVICE_STATES).includes(status) ? status : ALL_DEVICE_STATES;
    dispatch(setDeviceListState({ selectedId: id, state: deviceState }));
    closeAndNavigate(`/devices/${deviceState}?id=${id}`, { state: { internal: true } });
  };

  const onViewAll = () => {
    const filters = [{ key: attribute, operator: DEVICE_FILTERING_OPTIONS.$regex.key, scope, value: debouncedTerm }];
    closeAndNavigate({ pathname: `/devices/${ALL_DEVICE_STATES}`, search: formatDeviceSearch({ filters, pageState: {} }) });
  };

  const onKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Enter') {
      event.preventDefault();
      if (results[activeIndex]) {
        onSelect(results[activeIndex].device);
      } else if (showViewAll) {
        onViewAll();
      }
      return;
    }
    const offset = navigationOffsets[event.key];
    if (!offset || !optionCount) {
      return;
    }
    event.preventDefault();
    setActiveIndex(index => (index + offset + optionCount) % optionCount);
  };

  return (
    <Dialog
      fullWidth
      maxWidth="md"
      onClose={onClose}
      open={open}
      slotProps={{
        paper: { className: 'margin-top-x-small' },
        container: { className: 'align-items-start' },
        transition: { onEntered: () => inputRef.current?.focus() }
      }}
    >
      <InputBase
        className="padding-x-small padding-left-small"
        endAdornment={
          <InputAdornment position="end">
            <IconButton aria-label="close search" onClick={onClose} size="small">
              <CloseIcon fontSize="small" />
            </IconButton>
          </InputAdornment>
        }
        fullWidth
        inputRef={inputRef}
        onChange={({ target: { value } }) => setTerm(value)}
        onKeyDown={onKeyDown}
        placeholder={placeholder}
        slotProps={{ input: { className: classes.inputPlaceholder } }}
        startAdornment={
          <div className="flexbox align-items-center margin-right-x-small">
            <SearchAdornment />
            <Select
              className={classes.chipSelect}
              labelAttribute="label"
              onChange={onChangeAttribute}
              options={attributeOptions}
              selectionAttribute="key"
              value={selectedOption.key}
              width="auto"
            />
          </div>
        }
        value={term}
      />
      <Divider />
      {!searchHintDismissed && (
        <Alert onClose={onDismissHint} severity="info">
          Search is case-sensitive and matches the start of the value (e.g., ABC-1 matches ABC-123, abc-1 does not).
        </Alert>
      )}
      {(showResults || showSkeleton) && (
        <ListSubheader className={classes.subheader} component="div">
          Search results
        </ListSubheader>
      )}
      {showResults && (
        <List disablePadding>
          {results.map(({ checkIn, device, label, metadata, value }, index) => (
            <ListItemButton key={device.id} onClick={() => onSelect(device)} selected={index === activeIndex}>
              <ListItemIcon className={`margin-right-x-small ${classes.listItemIcon}`}>
                <DeviceIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText
                primary={
                  isIdAttributeSelected ? (
                    <Typography variant="subtitle2" component="span">
                      <HighlightedMatch className={classes.highlight} term={debouncedTerm} text={label} />
                    </Typography>
                  ) : (
                    <>
                      <Typography variant="subtitle2" component="span" className="margin-right-x-small">
                        {label}
                      </Typography>
                      <span className={classes.italic}>
                        ({attribute}: <HighlightedMatch className={classes.highlight} term={debouncedTerm} text={value} />)
                      </span>
                    </>
                  )
                }
                secondary={
                  <Breadcrumbs className={classes.metadata} component="div" separator="●">
                    {metadata.map(item => (
                      <span key={item}>{item}</span>
                    ))}
                    <span>
                      Latest activity: <ApproximateRelativeDate className={classes.inlineTime} updateTime={checkIn} />
                    </span>
                  </Breadcrumbs>
                }
                slotProps={{ primary: { className: 'text-overflow', variant: 'body2' }, secondary: { component: 'div', variant: 'caption' } }}
              />
              <DeviceStatus device={{ ...device, isOffline: device.status !== DEVICE_STATES.pending && device.isOffline }} />
            </ListItemButton>
          ))}
          {showViewAll && (
            <>
              <Divider />
              <ListItemButton className={classes.viewAll} onClick={onViewAll} selected={activeIndex === results.length}>
                <ListItemText
                  primary={
                    <div className="flexbox align-items-center">
                      View all {total.toLocaleString()} results
                      <ArrowForwardIcon fontSize="small" className="margin-left-x-small" />
                    </div>
                  }
                  slotProps={{ primary: { color: 'primary', variant: 'subtitle2' } }}
                />
              </ListItemButton>
            </>
          )}
        </List>
      )}
      {showSkeleton && <ResultsSkeleton />}
      {showEmptyState && (
        <div className={`flexbox centered column align-center ${classes.emptyState}`}>
          <Typography color="textSecondary" variant="subtitle1">
            No matching devices found
          </Typography>
          <Typography className="margin-top-x-small" color="textSecondary" variant="body2">
            Try adjusting your search term using the exact prefix and correct casing of your device identity attribute.
          </Typography>
        </div>
      )}
    </Dialog>
  );
};

export const SearchV2 = ({ className = '' }) => {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (open) {
      return;
    }
    const onGlobalKeyDown = event => {
      if (event.key !== shortcut || event.defaultPrevented || isEditableTarget(event.target)) {
        return;
      }
      event.preventDefault();
      setOpen(true);
    };
    window.addEventListener('keydown', onGlobalKeyDown);
    return () => window.removeEventListener('keydown', onGlobalKeyDown);
  }, [open]);

  return (
    <>
      <SearchTrigger className={className} onOpen={() => setOpen(true)} />
      <SearchDialog onClose={() => setOpen(false)} open={open} />
    </>
  );
};

export default SearchV2;
