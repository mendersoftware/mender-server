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
import { common } from '@mui/material/colors';

import { colors, typography } from './common';

const info = {
  main: common.white,
  dark: colors.grey[300],
  light: colors.grey[100],
  contrastText: common.black
};

const palette: PaletteOptions = {
  ...colors,
  mode: 'light',
  primary: {
    main: colors.purple[700],
    dark: colors.purple[800],
    light: colors.purple[400]
  },
  secondary: {
    main: colors.cyan[700],
    dark: colors.cyan[900],
    light: colors.cyan[500]
  },
  info,
  neutral: info,
  text: {
    primary: colors.grey[900],
    secondary: colors.grey[700],
    disabled: colors.grey[500]
  }
};

export const light: ThemeOptions = {
  palette,
  typography,
  components: {
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          '&:hover': {
            borderColor: common.black
          },
          [`&.${outlinedInputClasses.disabled}`]: {
            borderColor: common.black
          }
        }
      }
    }
  }
};
