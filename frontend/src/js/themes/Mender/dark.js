// Copyright 2022 Northern.tech AS
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
import { accordionClasses } from '@mui/material/Accordion';
import { buttonClasses } from '@mui/material/Button';

import { palette as commonPalette, overrides, typography } from './common';

const grey = {
  'A400': '#1d1f20',
  '900': '#4dc3bc',
  '800': '#89c6c1',
  '700': '#a0c7c4',
  '600': '#d7d7d7',
  '500': '#484747',
  '400': '#2a2828',
  '300': '#696868',
  '200': '#b1b1b1',
  '100': '#8c8c8c',
  '50': '#4a4a4a'
};
const paper = '#23252a';

const palette = {
  ...commonPalette,
  mode: 'dark',
  primary: {
    main: '#4d8b96'
  },
  border: {
    main: '#e7e7e7'
  },
  secondary: {
    lighter: '#9e6f8e',
    light: '#8e577b',
    main: '#7d3f69'
  },
  grey,
  background: {
    light: '#1d1f20',
    lightgrey: paper,
    default: '#1b1c22',
    paper
  },
  text: {
    primary: 'hsl(240deg 5% 93% / 67%)',
    hint: 'hsl(240deg 5% 93% / 67%)'
  },
  error: {
    main: '#ab1000',
    dark: '#dc6a5e'
  },
  tooltip: {
    ...commonPalette.tooltip,
    text: grey[600],
    tierTipBackground: grey[50]
  }
};

export const dark = {
  palette,
  typography,
  components: {
    ...overrides,
    MuiAccordion: {
      ...overrides.MuiAccordion,
      styleOverrides: {
        root: {
          ...overrides.MuiAccordion.styleOverrides.root,
          [`&.${accordionClasses.expanded}`]: {
            ...overrides.MuiAccordion.styleOverrides.root[`&.${accordionClasses.expanded}`],
            backgroundColor: palette.grey['A400']
          }
        }
      }
    },
    MuiButton: {
      ...overrides.MuiButton,
      styleOverrides: {
        ...overrides.MuiButton.styleOverrides,
        root: {
          ...overrides.MuiButton.styleOverrides.root,
          [`&.${buttonClasses.text}`]: {
            ...overrides.MuiButton.styleOverrides.root[`&.${buttonClasses.text}`],
            color: palette.text.primary
          }
        }
      }
    },
    MuiIconButton: {
      ...overrides.MuiIconButton,
      styleOverrides: {
        ...overrides.MuiIconButton.styleOverrides,
        root: {
          ...overrides.MuiIconButton.styleOverrides.root,
          color: palette.text.primary
        }
      }
    },
    MuiListItem: {
      ...overrides.MuiListItem,
      styleOverrides: {
        ...overrides.MuiListItem.styleOverrides,
        root: {
          ...overrides.MuiListItem.styleOverrides.root,
          [`&.active`]: {
            backgroundColor: palette.background.default
          },
          [`&.leftNav.active`]: {
            borderTop: `1px solid ${palette.grey[50]}`,
            borderBottom: `1px solid ${palette.grey[50]}`
          }
        }
      }
    },
    MuiListItemText: {
      ...overrides.MuiListItemText,
      styleOverrides: {
        ...overrides.MuiListItemText.styleOverrides,
        root: {
          ...overrides.MuiListItemText.styleOverrides.root,
          color: palette.text.primary
        }
      }
    }
  }
};
