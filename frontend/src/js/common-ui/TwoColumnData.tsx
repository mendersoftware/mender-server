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
import React, { CSSProperties, Fragment, ReactNode, createContext, useContext, useEffect, useRef, useState } from 'react';

// material ui
import { FileCopyOutlined as CopyToClipboardIcon } from '@mui/icons-material';
import { Chip, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import copy from 'copy-to-clipboard';

const getGridTemplateColumnSizing = (columnWidth: string): string => `${columnWidth} minmax(auto, 650px)`;

type DataValue = ReactNode | string;

const useStyles = makeStyles()(theme => ({
  textContent: {
    display: '-webkit-box',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    WebkitLineClamp: 5,
    WebkitBoxOrient: 'vertical',
    wordBreak: 'break-word'
  },
  copyIconOverride: {
    '&.copy-to-clipboard svg': {
      fill: theme.palette.action.active
    }
  },
  copyIconVisible: {
    opacity: 1
  },
  root: {
    '&.two-columns.column-data': {
      gridTemplateColumns: getGridTemplateColumnSizing('max-content'),
      maxWidth: 'initial',
      rowGap: theme.spacing()
    }
  }
}));

const ValueColumn = ({ setSnackbar, value = '' }: { setSnackbar?: (message: string) => void; value?: DataValue }) => {
  const { classes } = useStyles();
  const [tooltipVisible, setTooltipVisible] = useState(false);
  const isComponent = React.isValidElement(value);

  const onClick = () => {
    if (setSnackbar) {
      let copyable = value;
      if (isComponent) {
        copyable = value.props.value;
      }
      copy(copyable);
      setSnackbar('Value copied to clipboard');
    }
  };

  return (
    <div className={`flexbox center-aligned ${setSnackbar ? 'copy-to-clipboard' : ''} ${classes.copyIconOverride}`}>
      <Typography
        className={`${classes.textContent} ${setSnackbar ? 'clickable' : ''}`}
        component="div"
        onClick={onClick}
        title={isComponent ? value.props.value : value}
        onMouseEnter={() => setTooltipVisible(true)}
        onMouseLeave={() => setTooltipVisible(false)}
        variant="body2"
      >
        {value}
      </Typography>
      {setSnackbar && (
        <Tooltip title="Copy to clipboard" placement="top">
          <CopyToClipboardIcon color="action" fontSize="small" className={`margin-left-x-small ${tooltipVisible ? classes.copyIconVisible : ''}`} />
        </Tooltip>
      )}
    </div>
  );
};

const KeyColumn = ({ chipLikeKey, setColumnWidth, value }: { chipLikeKey?: boolean; setColumnWidth?: (width: number) => void; value: string }) => {
  const ref = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    if (ref.current && setColumnWidth) {
      const width = ref.current.scrollWidth;
      setColumnWidth(width);
    }
  }, [setColumnWidth, value]);

  return chipLikeKey ? (
    <Chip label={value} size="small" style={{ justifySelf: 'end' }} />
  ) : (
    <Typography ref={ref} className="key" variant="subtitle2" style={{ width: 'max-content' }}>
      {value}
    </Typography>
  );
};

export interface TwoColumnDataProps {
  chipLikeKey?: boolean;
  className?: string;
  columnWidth?: number;
  data?: Record<string, DataValue>;
  setColumnWidth?: (width: number) => void;
  setSnackbar?: (message: string) => void;
  style?: CSSProperties;
}

export const TwoColumnData = ({ chipLikeKey = false, className = '', columnWidth, data = {}, setColumnWidth, setSnackbar, style = {} }: TwoColumnDataProps) => {
  const { classes } = useStyles();
  return (
    <div
      className={`break-all two-columns ${classes.root} column-data ${className}`}
      style={{ ...style, gridTemplateColumns: columnWidth ? getGridTemplateColumnSizing(`${columnWidth}px`) : undefined }}
    >
      {Object.entries(data).map(([key, value]) => (
        <Fragment key={key}>
          <KeyColumn chipLikeKey={chipLikeKey} setColumnWidth={setColumnWidth} value={key} />
          <ValueColumn setSnackbar={setSnackbar} value={value} />
        </Fragment>
      ))}
    </div>
  );
};

const ColumnWidthContext = createContext<{ columnWidth: number; setColumnWidth: (width: number) => void } | null>(null);

export const ColumnWidthProvider = ({ children }: { children: ReactNode }) => {
  const [columnWidth, setColumnWidth] = useState<number>(0);

  const updateColumnWidth = (width: number) => {
    setColumnWidth(prevWidth => Math.max(prevWidth, width));
  };

  return <ColumnWidthContext.Provider value={{ columnWidth, setColumnWidth: updateColumnWidth }}>{children}</ColumnWidthContext.Provider>;
};

export const SynchronizedTwoColumnData = (props: Omit<TwoColumnDataProps, 'columnWidth' | 'setColumnWidth'>) => {
  const { columnWidth, setColumnWidth } = useContext(ColumnWidthContext)!;
  return <TwoColumnData {...props} columnWidth={columnWidth} setColumnWidth={setColumnWidth} />;
};
