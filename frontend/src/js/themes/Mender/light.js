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
import { formLabelClasses } from '@mui/material/FormLabel';
import { listItemTextClasses } from '@mui/material/ListItemText';

import { LIGHT_MODE } from '@northern.tech/store/constants';

import { palette as commonPalette, overrides, typography } from './common';

const grey = {
  '900': '#969696',
  '800': '#a9a9a9',
  '700': '#bcbcbc',
  '600': '#cfcfcf',
  '500': '#e9e9e9',
  '400': '#f7f7f7',
  '300': '#e6f2f1',
  '200': '#ddedec',
  '100': '#d8ebe9',
  '50': '#d4e9e7'
};

const palette = {
  ...commonPalette,
  mode: LIGHT_MODE,
  grey,
  background: {
    ...commonPalette.background,
    light: '#fdfdfd',
    lightgrey: grey[400],
    default: '#fff',
    dark: 'rgb(50, 50, 50)',
    darkBlue: '#284d68'
  },
  secondary: {
    ...commonPalette.secondary,
    lighter: '#8e577b',
    main: '#5d0f43'
  },
  tooltip: {
    ...commonPalette.tooltip,
    text: grey[50],
    tierTipBackground: '#f7fafb'
  },
  text: {
    ...commonPalette.text,
    primary: 'rgba(10, 10, 11, 0.78)',
    entryLink: '#7adce6'
  }
};

export const light = {
  palette,
  typography,
  components: {
    ...overrides,
    MuiCheckbox: {
      styleOverrides: {
        root: {
          color: palette.text.primary
        }
      }
    },
    MuiAccordion: {
      ...overrides.MuiAccordion,
      styleOverrides: {
        root: {
          ...overrides.MuiAccordion.styleOverrides.root,
          [`&.${accordionClasses.expanded}`]: {
            ...overrides.MuiAccordion.styleOverrides.root[`&.${accordionClasses.expanded}`],
            backgroundColor: palette.grey[400]
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
    MuiFormLabel: {
      styleOverrides: {
        root: {
          color: palette.text.hint,
          [`&.${formLabelClasses.focused}`]: {
            color: palette.primary.main
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
          color: palette.text.hint
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
          },
          [`&.navLink, &.navLink .${listItemTextClasses.root}`]: {
            color: palette.grey[900]
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
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: palette.background.dark
        }
      }
    }
  }
};
