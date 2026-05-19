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
import { type ReactNode, type Ref, useState } from 'react';

import { ClickAwayListener, SpeedDial, SpeedDialAction, SpeedDialIcon, Typography, getOverlayAlpha, lighten } from '@mui/material';
import { speedDialActionClasses } from '@mui/material/SpeedDialAction';
import { makeStyles } from 'tss-react/mui';

import { isDarkMode } from '@northern.tech/store/utils';
import { toggle } from '@northern.tech/utils/helpers';

export interface QuickAction {
  icon: ReactNode;
  key: string;
  onClick: () => void;
  title: ReactNode;
}

interface BaseQuickActionsProps {
  actions: QuickAction[];
  ariaLabel: string;
  label: string;
  onboardingComponent?: ReactNode;
  onToggle?: (open: boolean) => void;
  speedDialRef?: Ref<HTMLDivElement>;
  titleRef?: Ref<HTMLDivElement>;
}

const useStyles = makeStyles()(theme => ({
  container: {
    display: 'flex',
    position: 'fixed',
    bottom: theme.spacing(6.5),
    right: theme.spacing(6.5),
    zIndex: 10,
    minWidth: 'max-content',
    alignItems: 'flex-end',
    justifyContent: 'flex-end',
    pointerEvents: 'none',
    [`& .${speedDialActionClasses.staticTooltipLabel}`]: {
      minWidth: 'max-content'
    }
  },
  fab: { marginBottom: theme.spacing(2), marginRight: theme.spacing(2) },
  label: {
    background: isDarkMode(theme.palette.mode) ? lighten(theme.palette.background.paper, getOverlayAlpha(6)) : theme.palette.common.white,
    boxShadow: isDarkMode(theme.palette.mode) ? 'none' : theme.shadows[6],
    color: theme.palette.action?.active,
    padding: `${theme.spacing(1)} ${theme.spacing(2)}`,
    borderRadius: theme.spacing(0.5),
    marginRight: theme.spacing(1),
    marginBottom: theme.spacing(3),
    pointerEvents: 'auto'
  }
}));

export const BaseQuickActions = ({ actions, ariaLabel, titleRef, label, speedDialRef, onboardingComponent = null, onToggle }: BaseQuickActionsProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const { classes } = useStyles();

  const handleToggle = (e?: React.MouseEvent) => {
    e?.stopPropagation();
    setIsOpen(toggle);
    onToggle?.(!isOpen);
  };

  const handleClickAway = () => {
    setIsOpen(false);
    onToggle?.(false);
  };

  if (!actions.length) {
    return null;
  }

  return (
    <div className={classes.container}>
      <Typography variant="body1" className={`clickable ${classes.label}`} onClick={handleToggle} ref={titleRef}>
        {label}
      </Typography>
      <ClickAwayListener onClickAway={handleClickAway}>
        <SpeedDial className={classes.fab} ariaLabel={ariaLabel} icon={<SpeedDialIcon />} onClick={handleToggle} open={isOpen} ref={speedDialRef}>
          {actions.map(action => (
            <SpeedDialAction
              key={action.key}
              aria-label={action.key}
              icon={action.icon}
              slotProps={{ tooltip: { title: action.title, open: true } }}
              onClick={action.onClick}
            />
          ))}
        </SpeedDial>
      </ClickAwayListener>
      {onboardingComponent}
    </div>
  );
};

export default BaseQuickActions;
