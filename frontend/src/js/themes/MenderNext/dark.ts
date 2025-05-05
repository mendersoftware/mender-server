// Copyright 2025 Northern.tech AS
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
import { PaletteOptions, ThemeOptions, outlinedInputClasses } from '@mui/material';
import { common, red } from '@mui/material/colors';

import { colors, typography } from './common';

const info = {
  main: colors.grey[700],
  dark: colors.grey[800],
  light: colors.grey[300],
  contrastText: common.white
};

const palette: PaletteOptions = {
  ...colors,
  mode: 'dark',
  primary: {
    main: colors.purple[200],
    dark: colors.purple[400],
    light: colors.purple[50]
  },
  secondary: {
    main: colors.cyan[200],
    dark: colors.cyan[400],
    light: colors.cyan[50]
  },
  error: {
    main: red[400],
    dark: red[700],
    light: red[300],
    contrastText: common.white
  },
  info,
  neutral: info
};

export const dark: ThemeOptions = {
  palette,
  typography,
  components: {
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          '&:hover': {
            borderColor: common.white
          },
          [`&.${outlinedInputClasses.disabled}`]: {
            borderColor: common.white
          }
        }
      }
    }
  }
};
