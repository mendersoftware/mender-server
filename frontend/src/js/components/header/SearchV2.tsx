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
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router';

// material ui
import { ArrowForward as ArrowForwardIcon, Close as CloseIcon, DeveloperBoard as DeviceIcon, Search as SearchIcon } from '@mui/icons-material';
import {
  Alert,
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
  buttonBaseClasses,
  listItemIconClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { getDeviceIdentityText } from '@northern.tech/common-ui/DeviceIdentity';
import { ApproximateRelativeDate } from '@northern.tech/common-ui/Time';
import { ALL_DEVICE_STATES, DEVICE_FILTERING_OPTIONS, DEVICE_STATES, TIMEOUTS } from '@northern.tech/store/constants';
import type { Device } from '@northern.tech/store/devicesSlice';
import { formatDeviceSearch } from '@northern.tech/store/locationutils';
import { getIdAttribute, getUserSettings } from '@northern.tech/store/selectors';
import { useAppDispatch, useAppSelector } from '@northern.tech/store/store';
import { saveUserSettings, searchIdentities, setDeviceListState } from '@northern.tech/store/thunks';
import type { SearchIdentityParams } from '@northern.tech/types/MenderTypes';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import { idAttributeTitleMap } from '../devices/AuthorizedDevices';
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
  metadata: string;
}

const toResult = (device: Device, idAttribute): SearchResultItem => ({
  checkIn: device.check_in_time_exact ?? device.check_in_time_rounded,
  device,
  label: getDeviceIdentityText({ device, idAttribute }),
  metadata: [device.attributes.device_type?.join(', '), getDeviceSoftwareText(device.attributes)].filter(Boolean).join(' · ')
});

const useStyles = makeStyles()(theme => ({
  emptyState: { minHeight: 200 },
  highlight: { backgroundColor: theme.palette.highlight?.main },
  inlineTime: { display: 'inline', fontSize: 'inherit' },
  inputPlaceholder: { '&::placeholder': { color: theme.palette.text.secondary, opacity: 1 } },
  viewAll: { background: theme.palette.action.hover },
  listItemIcon: {
    [`&.${listItemIconClasses.root}`]: { minWidth: 'auto' }
  },
  shortcut: {
    border: `1px solid ${theme.palette.divider}`,
    background: theme.palette.grey[100],
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

  const { attribute, scope: idScope } = useAppSelector(getIdAttribute);
  const { searchHintDismissed } = useAppSelector(getUserSettings);

  const scope = idScope as SearchIdentityParams['scope'];
  const dispatch = useAppDispatch();
  const { classes } = useStyles();
  const debouncedTerm = useDebounce(term, TIMEOUTS.debounceDefault);

  const placeholder = `Search by ${idAttributeTitleMap[attribute] ?? attribute} starting with...`;
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
    const idAttribute = { attribute, scope };
    dispatch(
      searchIdentities({
        attributes: [idAttribute],
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
        setSearch({ results: devices.map(device => toResult(device, idAttribute)), term: debouncedTerm, total });
      })
      .catch(() => isCurrent && setSearch({ results: [], term: debouncedTerm, total: 0 }));
    return () => {
      isCurrent = false;
    };
  }, [attribute, debouncedTerm, dispatch, scope]);

  useEffect(() => {
    setActiveIndex(0);
  }, [results]);

  const onDismissHint = () => dispatch(saveUserSettings({ searchHintDismissed: true }));

  const closeAndNavigate = (to, options?) => {
    onClose();
    setTimeout(() => navigate(to, options), TIMEOUTS.debounceShort);
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
      maxWidth="lg"
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
        startAdornment={<SearchAdornment />}
        value={term}
      />
      <Divider />
      {!searchHintDismissed && (
        <Alert onClose={onDismissHint} severity="info">
          Search is case-sensitive and matches the start of the value (e.g., ABC-1 matches ABC-123, abc-1 does not).
        </Alert>
      )}
      {(showResults || showSkeleton) && <ListSubheader component="div">Search results</ListSubheader>}
      {showResults && (
        <List disablePadding>
          {results.map(({ checkIn, device, label, metadata }, index) => (
            <ListItemButton key={device.id} onClick={() => onSelect(device)} selected={index === activeIndex}>
              <ListItemIcon className={`margin-right-x-small ${classes.listItemIcon}`}>
                <DeviceIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText
                primary={<HighlightedMatch className={classes.highlight} term={debouncedTerm} text={label} />}
                secondary={
                  <>
                    {metadata && `${metadata} · `}
                    Latest activity: <ApproximateRelativeDate className={classes.inlineTime} updateTime={checkIn} />
                  </>
                }
                slotProps={{ primary: { className: 'text-overflow', variant: 'body2' }, secondary: { className: 'text-overflow', variant: 'caption' } }}
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
                      View all {total} results
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
