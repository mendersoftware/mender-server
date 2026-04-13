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
import type { ReactNode, SyntheticEvent } from 'react';

import { Divider, Drawer, dividerClasses, drawerClasses } from '@mui/material';
import type { Breakpoint, DrawerProps } from '@mui/material';
import { styled } from '@mui/material/styles';

import { DrawerTitle } from './DrawerTitle';
import type { DrawerTitleProps } from './DrawerTitle';

export type DrawerSize = Exclude<Breakpoint, 'xs'> | 'auto';

const sizes: Record<string, { min: number; vw: string }> = {
  sm: { vw: '50vw', min: 480 },
  md: { vw: '60vw', min: 600 },
  lg: { vw: '67vw', min: 720 },
  xl: { vw: '75vw', min: 900 }
};

const SizedDrawer = styled(Drawer, { shouldForwardProp: prop => prop !== 'size' })<{ size?: DrawerSize }>(({ theme, size = 'md' }) => ({
  [`& .${drawerClasses.paper}`]: {
    width: '100vw',
    minWidth: 0,
    [theme.breakpoints.up('sm')]:
      size === 'auto' ? { width: 'auto', minWidth: 0 } : { width: sizes[size as Breakpoint].vw, minWidth: sizes[size as Breakpoint].min },
    [`& > .${dividerClasses.root}`]: { marginLeft: theme.spacing(-8), marginRight: theme.spacing(-8) }
  }
}));

interface BaseDrawerSlotProps extends Pick<DrawerProps, 'slotProps'> {
  header: Omit<DrawerTitleProps, 'onClose'>;
}

interface BaseDrawerProps extends Omit<DrawerProps, 'title' | 'slotProps'> {
  notification?: ReactNode;
  size?: DrawerSize;
  slotProps: BaseDrawerSlotProps;
}

const BaseDrawer = ({ children, className = '', notification, onClose, open, size = 'md', slotProps, ...rest }: BaseDrawerProps) => {
  const { header: headerProps, ...drawerSlotProps } = slotProps ?? {};
  const handleHeaderClose = (event: SyntheticEvent) => onClose?.(event, 'escapeKeyDown');
  return (
    <SizedDrawer
      className={`${open ? 'fadeIn' : 'fadeOut'} ${className}`}
      anchor="right"
      open={open}
      onClose={onClose}
      size={size}
      slotProps={drawerSlotProps as DrawerProps['slotProps']}
      {...rest}
    >
      {headerProps && <DrawerTitle {...headerProps} onClose={handleHeaderClose} />}
      {notification}
      <Divider className="margin-top-x-small margin-bottom-small" />
      {children}
    </SizedDrawer>
  );
};

export default BaseDrawer;
