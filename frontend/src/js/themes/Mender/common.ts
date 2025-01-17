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
import { accordionSummaryClasses } from '@mui/material/AccordionSummary';
import { listItemClasses } from '@mui/material/ListItem';

const secondaryText = 'rgba(0, 0, 0, 0.54)';

/**
 * @param qualitative if set is an ordered set of distinct colors availabe for programatic use.
 * @example Chart dataset colors
 */
export const palette = {
  background: {
    terminal: 'rgba(33, 36, 41, 0.25)'
  },
  primary: {
    main: '#337a87'
  },
  border: {
    main: '#e3e3e3'
  },
  secondary: {
    main: '#5d0f43',
    lighter: '#921267'
  },
  error: {
    light: 'rgba(93, 15, 67, 0.075)',
    main: '#ab1000',
    dark: '#770b00' // hardcode same as darken to match less variables
  },
  success: {
    main: '#009e73'
  },
  text: {
    /**
     * color matched from variables.less @text of #404041 but by opacity, same as main
     */
    primary: 'rgba(10, 10, 11, 0.78)',
    secondary: secondaryText,
    hint: secondaryText,
    link: '#679ba5'
  },
  brand: {
    mender: '#015969',
    northernTech: '#28aee4'
  }
};

const generatedColors = {
  a: palette.primary.main,
  b: '#a31773',
  c: '#00859e',
  d: '#14cfda',
  e: '#9bfff0',
  f: '#d5d5d5'
};
const qualitative = {};
for (const [grp, col] of Object.entries(generatedColors)) {
  qualitative[grp] = {
    main: col
  };
}
palette['qualitative'] = qualitative;

export const chartColorPalette = Object.values(generatedColors);

const round = value => Math.round(value * 1e4) / 1e4;
const htmlFontSize = 16;
const fontSize = 13;
const coef = fontSize / 14;
const pxToRem = size => `${round((size / htmlFontSize) * coef)}rem`;

export const typography = {
  fontFamily: 'Lato, sans-serif',
  fontSize, // will be transformed to rem automatically by mui
  body1: {
    lineHeight: 1.5
  },
  pxToRem
};

const componentProps = {
  MuiFormControl: {
    defaultProps: {
      variant: 'outlined'
    }
  },
  MuiInput: {
    defaultProps: {
      variant: 'outlined'
    }
  },
  MuiTextField: {
    defaultProps: {
      variant: 'outlined'
    }
  },
  MuiSelect: {
    defaultProps: {
      autoWidth: true,
      variant: 'outlined'
    }
  }
};

export const overrides = {
  ...componentProps,
  MuiSnackbarContent: {
    styleOverrides: {
      action: {
        color: '#9E6F8E'
      }
    }
  },
  MuiTab: {
    styleOverrides: {
      root: {
        textTransform: 'none'
      }
    }
  },
  MuiAccordion: {
    styleOverrides: {
      root: {
        border: 'none',
        boxShadow: 'none',
        '&:before': {
          display: 'none'
        },
        padding: 0,
        [`&.${accordionClasses.expanded}`]: {
          margin: 'auto'
        }
      }
    }
  },
  MuiAccordionSummary: {
    styleOverrides: {
      root: {
        marginBottom: 0,
        height: 48,
        [`&.${accordionSummaryClasses.expanded}`]: {
          height: 48,
          minHeight: 48
        }
      },
      content: {
        alignItems: 'center',
        [`&.${accordionSummaryClasses.expanded}`]: {
          margin: 0
        },
        '& > :last-child': {
          paddingRight: 12
        }
      }
    }
  },
  MuiAccordionDetails: {
    styleOverrides: {
      root: {
        flexDirection: 'column'
      }
    }
  },
  MuiInput: {
    ...componentProps.MuiInput,
    styleOverrides: {
      underline: {
        '&:before': {
          borderBottom: '1px solid rgb(224, 224, 224)'
        },
        '&:hover:not($disabled):before': {
          borderBottom: `2px solid ${palette.primary.main} !important`
        },
        '&:after': {
          borderBottom: `2px solid ${palette.primary.main}`
        }
      }
    }
  },
  MuiFormControl: {
    ...componentProps.MuiFormControl,
    styleOverrides: {
      root: {
        minWidth: '240px'
      }
    }
  },
  MuiFormControlLabel: {
    styleOverrides: {
      root: {
        marginTop: '18px'
      }
    }
  },
  MuiIconButton: {
    styleOverrides: {
      root: {
        fontSize: '1.2rem'
      }
    }
  },
  MuiButton: {
    styleOverrides: {
      root: {
        borderRadius: 2,
        fontSize: 14,
        fontWeight: 'bold',
        '&:hover': {
          colors: palette.primary.main
        }
      },
      text: {
        padding: '10px 15px'
      }
    }
  },
  MuiSvgIcon: {
    styleOverrides: {
      root: {
        iconButton: {
          marginRight: '8px'
        }
      }
    }
  },
  MuiListItem: {
    styleOverrides: {
      root: {
        paddingTop: 11,
        paddingBottom: 11,
        [`&.${listItemClasses.disabled}`]: {
          opacity: 1
        }
      }
    }
  },
  MuiListItemText: {
    styleOverrides: {
      root: {
        marginTop: 0,
        marginBottom: 0
      }
    }
  },
  MuiTableCell: {
    styleOverrides: {
      root: {
        padding: '0px 24px 0px 24px',
        height: '48px'
      },
      head: {
        height: '56px',
        lineHeight: '1.15rem'
      },
      paddingCheckbox: {
        padding: '0 0 0 6px',
        width: '54px'
      }
    }
  },
  MuiDrawer: {
    styleOverrides: {
      paper: {
        minWidth: '40vw',
        maxWidth: '80vw',
        padding: '30px 75px 5%'
      }
    }
  }
};
