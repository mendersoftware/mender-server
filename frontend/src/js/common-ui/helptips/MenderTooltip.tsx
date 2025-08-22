// Copyright 2021 Northern.tech AS
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
//@ts-nocheck
import { useEffect, useState } from 'react';

import type { TooltipProps } from '@mui/material';
import { ClickAwayListener, Tooltip } from '@mui/material';
import { withStyles } from 'tss-react/mui';

import { toggle } from '@northern.tech/utils/helpers';
import type { PositioningStrategy } from '@popperjs/core';

export const MenderTooltip = withStyles(Tooltip, ({ palette, shadows, spacing }) => ({
  arrow: {
    color: palette.background.paper
  },
  tooltip: {
    backgroundColor: palette.background.paper,
    boxShadow: shadows[1],
    color: palette.text.primary,
    padding: spacing(2),
    fontSize: 'small',
    maxWidth: 600,
    info: {
      maxWidth: 300,
      color: palette.text.hint,
      backgroundColor: palette.grey[500]
    }
  }
}));

export interface MenderTooltipClickableProps extends TooltipProps {
  onboarding?: boolean;
  onOpenChange?: (open: boolean) => void;
  startOpen?: boolean;
  tooltipComponent?: typeof MenderTooltip;
  visibility?: boolean;
}

export const MenderTooltipClickable = ({
  children,
  onboarding,
  startOpen = false,
  visibility = startOpen,
  onOpenChange,
  tooltipComponent = MenderTooltip,
  ...remainingProps
}): MenderTooltipClickableProps => {
  const [open, setOpen] = useState(startOpen || false);

  useEffect(() => {
    setOpen(visibility);
  }, [visibility]);

  useEffect(() => {
    if (!onOpenChange) {
      return;
    }
    onOpenChange(open);
  }, [open, onOpenChange]);

  const toggleVisibility = () => setOpen(toggle);

  const hide = () => setOpen(false);

  const Component = tooltipComponent as typeof Tooltip;
  const extraProps = onboarding
    ? {
        PopperProps: {
          disablePortal: true,
          popperOptions: {
            strategy: 'fixed' as PositioningStrategy,
            modifiers: [
              { name: 'flip', enabled: false },
              { name: 'preventOverflow', enabled: true, options: { boundary: window, altBoundary: false } }
            ]
          }
        }
      }
    : {};
  return (
    <ClickAwayListener onClickAway={hide}>
      <Component
        arrow={!onboarding}
        open={open}
        disableFocusListener
        disableHoverListener
        disableTouchListener
        onOpen={() => setOpen(true)}
        {...extraProps}
        {...remainingProps}
      >
        <div onClick={toggleVisibility}>{children}</div>
      </Component>
    </ClickAwayListener>
  );
};

export default MenderTooltip;
