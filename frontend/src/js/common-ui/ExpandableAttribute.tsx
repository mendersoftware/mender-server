// Copyright 2019 Northern.tech AS
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
import { useCallback, useEffect, useRef, useState } from 'react';

// material ui
import { FileCopyOutlined as CopyToClipboardIcon } from '@mui/icons-material';
import { ListItem, ListItemText, Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { toggle } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';

import { Link } from './Link';

const useStyles = makeStyles()(theme => ({
  copyable: {
    cursor: 'pointer',
    '& > svg': {
      color: theme.palette.primary.main,
      opacity: 0,
      transition: 'opacity 0.2s ease-in-out'
    },
    '&:hover > svg': {
      opacity: 1
    }
  }
}));

const defaultClasses = { root: 'attributes' };

export const ExpandableAttribute = ({
  className = '',
  copyToClipboard,
  dividerDisabled,
  onExpansion,
  primary,
  secondary,
  secondaryTypographyProps = {},
  setSnackbar,
  style,
  textClasses,
  ...remainder
}) => {
  const { classes } = useStyles();
  const textContent = useRef(null);
  const [expanded, setExpanded] = useState(false);
  const [overflowActive, setOverflowActive] = useState(false);
  const [tooltipVisible, setTooltipVisible] = useState(false);

  useEffect(() => {
    if (textContent.current) {
      const overflowActiveCurrently =
        textContent.current.scrollWidth > textContent.current.clientWidth || textContent.current.scrollHeight > textContent.current.clientHeight;
      if (overflowActive !== overflowActiveCurrently && !expanded) {
        setOverflowActive(overflowActiveCurrently);
      }
    }
  }, [expanded, overflowActive, textContent]);

  const onClick = useCallback(() => {
    if (copyToClipboard) {
      // Date/Time components
      copy(secondary);
      setSnackbar('Value copied to clipboard');
    }
    if (!expanded && !!onExpansion) {
      onExpansion();
    }
    setExpanded(toggle);
  }, [copyToClipboard, expanded, onExpansion, secondary, setSnackbar]);

  const currentTextClasses = `${textClasses ? textClasses.secondary : 'inventory-text'}${expanded && overflowActive ? ' expanded-attribute' : ''}`;
  const secondaryText = (
    <>
      <span className={currentTextClasses} ref={textContent}>
        {secondary}
      </span>{' '}
      {overflowActive ? <Link>show {expanded ? 'less' : 'more'}</Link> : null}
    </>
  );

  const cssClasses = { ...defaultClasses, root: `${defaultClasses.root} ${copyToClipboard ? classes.copyable : ''}`.trim() };

  return (
    <div className={className} onClick={onClick} onMouseEnter={() => setTooltipVisible(true)} onMouseLeave={() => setTooltipVisible(false)} style={style}>
      <ListItem classes={cssClasses} divider={!dividerDisabled} {...remainder}>
        <ListItemText
          primary={primary}
          secondary={secondaryText}
          slotProps={{
            secondary: { title: secondary, component: 'div', ...secondaryTypographyProps }
          }}
        />
        {copyToClipboard ? (
          <Tooltip title={'Copy to clipboard'} placement="top" open={tooltipVisible}>
            <CopyToClipboardIcon fontSize="small" />
          </Tooltip>
        ) : null}
      </ListItem>
    </div>
  );
};

export default ExpandableAttribute;
