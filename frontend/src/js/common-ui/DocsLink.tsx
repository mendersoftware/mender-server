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
import { ReactNode, forwardRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { Description as DescriptionIcon, Launch as LaunchIcon } from '@mui/icons-material';
import { Chip, Collapse, Typography, TypographyProps, chipClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { getDocsVersion, getFeatures } from '@northern.tech/store/selectors';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { yes } from '@northern.tech/utils/helpers';

import { MenderTooltipClickable } from './helptips/MenderTooltip';

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
  deviceIdentity: { id: 'deviceIdentity', path: 'client-installation/identity' },
  dynamicGroups: { id: 'dynamicGroups', path: 'overview/device-group#dynamic-group' },
  limitedDeployments: { id: 'limitedDeployments', path: 'overview/deployment#deployment-to-dynamic-groups' },
  phasedDeployments: { id: 'phasedDeployments', path: 'overview/customize-the-update-process' },
  pausedDeployments: { id: 'pausedDeployments', path: 'overview/customize-the-update-process#synchronized-updates' },
  retryDeployments: { id: 'retryDeployments', path: 'overview/deployment' },
  releases: { id: 'releases', path: 'overview/artifact' },
  rbac: { id: 'rbac', path: 'overview/role.based.access.control' },
  webhookSecret: { id: 'webhookSecret', path: 'server-integration/webhooks#signature-header' }
};

export const DocsTooltip = ({ anchor = {}, id = '', ...props }) => {
  const [isHovering, setIsHovering] = useState(false);
  const debouncedHovering = useDebounce(isHovering, TIMEOUTS.debounceDefault);
  const { classes } = useStyles();

  if (!DOCSTIPS[id]) {
    return null;
  }
  const { content, path } = DOCSTIPS[id];

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
      <DocsLink path={path}>
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
          onDelete={yes}
          onMouseOver={() => setIsHovering(true)}
          onMouseOut={() => setIsHovering(false)}
          variant="outlined"
        />
      </DocsLink>
    </MenderTooltipClickable>
  );
};

export const InlineLaunchIcon = () => <LaunchIcon style={{ verticalAlign: 'sub' }} fontSize="small" />;

interface DocsTextLinkProps {
  [key: string]: unknown;
  capitalizedStart?: boolean;
  children?: ReactNode;
  id: keyof typeof DOCSTIPS;
  typographyProps: Partial<TypographyProps>;
}

const textLinkDefaultProps: TypographyProps = { variant: 'body1' };

export const DocsTextLink = ({ capitalizedStart = true, children, id, typographyProps = textLinkDefaultProps, ...props }: DocsTextLinkProps) => {
  if (!DOCSTIPS[id]) {
    return null;
  }
  const { path } = DOCSTIPS[id];
  return (
    <DocsLink path={path} {...props}>
      <Typography className={`inline ${capitalizedStart ? 'capitalized-start' : ''}`} color="primary" {...typographyProps}>
        {children || 'learn more'}
      </Typography>
    </DocsLink>
  );
};

export const DocsLink = forwardRef(({ children, className = '', path = '', title = '', ...remainder }, ref) => {
  const docsVersion = useSelector(getDocsVersion);
  const { isHosted } = useSelector(getFeatures);
  const target = `https://docs.mender.io/${path}`;

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
