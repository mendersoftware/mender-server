// Copyright 2023 Northern.tech AS
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
import React, { forwardRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { Description as DescriptionIcon } from '@mui/icons-material';
import { Chip, Collapse, chipClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { getDocsVersion, getFeatures } from '@northern.tech/store/selectors';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import { MenderTooltipClickable } from './mendertooltip';

const useStyles = makeStyles()(theme => ({
  iconAura: {
    position: 'absolute',
    top: -6,
    bottom: -4,
    left: -7,
    right: -6.5,
    border: `1px dashed ${theme.palette.text.disabled}`,
    borderRadius: '50%',
    '&.hovering': {
      borderColor: 'transparent'
    }
  },
  chip: {
    borderStyle: 'dashed',
    [`.${chipClasses.deleteIcon}`]: {
      fontSize: 'smaller'
    },
    '&.not-hovering': {
      borderColor: 'transparent',
      color: theme.palette.text.disabled,
      [`.${chipClasses.deleteIcon}`]: {
        color: theme.palette.text.disabled
      },
      [`.${chipClasses.label}`]: {
        paddingLeft: 0,
        visibility: 'collapse'
      }
    }
  }
}));

export const DOCSTIPS = {
  deviceConfig: { id: 'deviceConfig', path: 'add-ons/configure' },
  dynamicGroups: { id: 'dynamicGroups', path: 'overview/device-group#dynamic-group' },
  limitedDeployments: { id: 'limitedDeployments', path: 'overview/deployment#deployment-to-dynamic-groups' },
  phasedDeployments: { id: 'phasedDeployments', path: 'overview/customize-the-update-process' },
  pausedDeployments: { id: 'pausedDeployments', path: 'overview/customize-the-update-process#synchronized-updates' },
  retryDeployments: { id: 'retryDeployments', path: 'overview/deployment' },
  releases: { id: 'releases', path: 'overview/artifact' },
  webhookSecret: { id: 'webhookSecret', path: 'server-integration/webhooks#signature-header' }
};

export const DocsTooltip = ({ anchor = {}, id = '', ...props }) => {
  const [isHovering, setIsHovering] = useState(false);
  const debouncedHovering = useDebounce(isHovering, TIMEOUTS.debounceDefault);
  const docsVersion = useSelector(getDocsVersion);
  const { isHosted } = useSelector(getFeatures);
  const { classes } = useStyles();
  const { content, path } = DOCSTIPS[id] || {};
  const target = `https://docs.mender.io/${docsVersion}${path}`;

  const onClick = () => {
    const docsParams = { headers: { 'x-mender-docs': docsVersion } };
    fetch(target, isHosted ? {} : docsParams);
    window.open(target, '_blank');
  };

  const hoverClass = debouncedHovering ? 'hovering' : 'not-hovering';
  return (
    <MenderTooltipClickable
      placement="bottom-start"
      disableFocusListener={false}
      disableHoverListener={false}
      disableTouchListener={false}
      style={anchor}
      title={content}
      {...props}
    >
      <Chip
        color="primary"
        className={`${classes.chip} ${hoverClass}`}
        label={
          <Collapse in={debouncedHovering} orientation="horizontal">
            Learn more
          </Collapse>
        }
        deleteIcon={
          <div className="relative">
            <DescriptionIcon fontSize="small" />
            <div className={`${classes.iconAura} ${hoverClass}`} />
          </div>
        }
        onClick={onClick}
        onDelete={onClick}
        onMouseOver={() => setIsHovering(true)}
        onMouseOut={() => setIsHovering(false)}
        variant="outlined"
      />
    </MenderTooltipClickable>
  );
};

export const DocsLink = forwardRef(({ children, className = '', path, title = '', ...remainder }, ref) => {
  const docsVersion = useSelector(getDocsVersion);
  const { isHosted } = useSelector(getFeatures);
  const target = `https://docs.mender.io/${docsVersion}${path}`;

  const onClickHandler = () => {
    const docsParams = { headers: { 'x-mender-docs': docsVersion } };
    fetch(target, isHosted ? {} : docsParams);
  };

  return (
    // eslint-disable-next-line react/jsx-no-target-blank
    <a className={className} {...remainder} href={target} onClick={onClickHandler} ref={ref} target="_blank" rel={isHosted ? 'noopener' : ''}>
      {children ? children : title}
    </a>
  );
});

DocsLink.displayName = 'DocsLink';

export default DocsLink;
