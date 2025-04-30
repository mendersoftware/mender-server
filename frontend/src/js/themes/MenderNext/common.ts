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
import { blue, cyan, grey } from '@mui/material/colors';

export const typography = {
  fontFamily: 'Red Hat Display',
  body1: {
    fontFamily: 'Red Hat Text'
  }
};

export const colors = {
  grey: {
    ...grey,
    50: '#fafafb',
    100: '#f4f4f5',
    200: '#ededee',
    300: '#dededf',
    400: '#bbbbbc',
    500: '#9c9c9d',
    600: '#737374',
    700: '#5f5f60',
    800: '#404041',
    900: '#1f1f20'
  },
  purple: {
    ...blue, // ...remainder colors (A100 - A700) should be based on 'blue'
    50: '#f4e1eb',
    100: '#e6b5cf',
    200: '#d886af',
    300: '#cb5790',
    400: '#c33579',
    500: '#bc0a63',
    600: '#ac0d5e',
    700: '#970f57',
    800: '#820f50',
    900: '#5d0f43'
  },
  cyan: {
    ...cyan,
    50: '#e1f5fd',
    100: '#b5e7f9',
    200: '#85d8f4',
    300: '#56c8ed',
    400: '#39bde5',
    500: '#2fb2dd',
    600: '#28a3c9',
    700: '#1d8faf',
    800: '#147b96',
    900: '#015969'
  }
};
