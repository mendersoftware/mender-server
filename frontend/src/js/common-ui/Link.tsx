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
import { Link as RouterLink } from 'react-router-dom';

import type { LinkProps as MuiLinkProps } from '@mui/material';
import { Link as MuiLink } from '@mui/material';

export type LinkProps = Omit<MuiLinkProps, 'href'> & { external?: boolean; href?: string; to?: string };

export const Link = ({ href, to, external = false, onClick, target, rel, component, ...rest }: LinkProps) => {
  if (to) {
    return <MuiLink component={RouterLink} to={to} onClick={onClick} {...rest} />;
  }
  if (href) {
    return (
      <MuiLink
        href={href}
        onClick={onClick}
        target={target ?? (external ? '_blank' : undefined)}
        rel={rel ?? (external ? 'noopener noreferrer' : undefined)}
        {...rest}
      />
    );
  }
  if (onClick) {
    return <MuiLink component={component ?? 'button'} onClick={onClick} {...rest} />;
  }
  return <MuiLink component={component ?? 'span'} {...rest} />;
};

export default Link;
