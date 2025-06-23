// Copyright 2020 Northern.tech AS
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
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { Clear as ClearIcon, Settings, Square } from '@mui/icons-material';
import { IconButton, LinearProgress, linearProgressClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Loader from '@northern.tech/common-ui/Loader';
import { ALL_DEVICES, TIMEOUTS, chartTypes, rootfsImageVersion, softwareTitleMap } from '@northern.tech/store/constants';
import { getDeviceReports, getGroupsById } from '@northern.tech/store/selectors';
import { getReportDataWithoutBackendSupport, updateReportData } from '@northern.tech/store/thunks';
import { ensureVersionString } from '@northern.tech/store/utils';
import { chartColorPalette } from '@northern.tech/themes/Mender';
import { isEmpty, toggle } from '@northern.tech/utils/helpers';
import { VictoryBar, VictoryContainer, VictoryPie, VictoryStack } from 'victory';

import BaseWidget from './BaseWidget';
import { ChartEditWidget, Header, RemovalWidget } from './ChartAddition';

const seriesOther = '__OTHER__';

const createColorClassName = hexColor => `color-${hexColor.slice(1)}`;

const useStyles = makeStyles()(theme => ({
  indicator: { fontSize: 10, minWidth: 'initial', marginLeft: 4 },
  legendItem: {
    alignItems: 'center',
    display: 'grid',
    gridTemplateColumns: '1fr max-content',
    columnGap: theme.spacing(2),
    '&.indicating': {
      gridTemplateColumns: 'min-content 1fr max-content',
      columnGap: theme.spacing()
    }
  },
  wrapper: {
    display: 'grid',
    gridTemplateColumns: '200px 1fr',
    columnGap: theme.spacing(2),
    marginBottom: theme.spacing(2),
    '&>.flexbox.column > *': {
      height: 20
    },
    '.barchart': {
      [`.${linearProgressClasses.root}`]: {
        backgroundColor: theme.palette.grey[400]
      },
      ...Object.values(chartColorPalette).reduce(
        (accu, color) => ({
          ...accu,
          [`.${createColorClassName(color)} .${linearProgressClasses.barColorPrimary}`]: { backgroundColor: color }
        }),
        {
          [`.${createColorClassName(theme.palette.grey[400])} .${linearProgressClasses.barColorPrimary}`]: { backgroundColor: theme.palette.grey[400] }
        }
      )
    }
  }
}));

const ChartLegend = ({ classes, data, events = [], showIndicators = true }) => {
  const { eventHandlers = {} } = events[0];
  const { onClick } = eventHandlers;
  return (
    <div className="flexbox column">
      {data.map(({ fill, x, title, tip, value }) => (
        <div
          className={`clickable ${classes.legendItem} ${showIndicators ? 'indicating' : ''}`}
          key={x}
          onClick={e => onClick(e, { datum: { x } })}
          title={tip}
        >
          {showIndicators && <Square className={classes.indicator} style={{ fill }} />}
          <div className="text-overflow">{title}</div>
          <div>{value.toLocaleString()}</div>
        </div>
      ))}
    </div>
  );
};

const VictoryBarChart = ({ data, totals, ...remainder }) => (
  <VictoryStack {...remainder} animate={{ duration: 700, onLoad: { duration: 700 } }} horizontal padding={{ left: 0, right: 0, top: 0, bottom: 15 }}>
    <VictoryBar alignment="start" barWidth={16} sortKey={['y']} sortOrder="ascending" data={data} />
    <VictoryBar alignment="start" barWidth={16} sortKey={['y']} sortOrder="descending" data={totals} />
  </VictoryStack>
);

const BarChart = ({ data, events }) => {
  const [chartData, setChartData] = useState([]);
  const timer = useRef();

  useEffect(() => {
    setChartData(data.map(item => ({ ...item, y: 0 })));
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setChartData(data), TIMEOUTS.debounceDefault);
    return () => {
      clearTimeout(timer.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(data)]);

  const { eventHandlers = {} } = events[0];
  const { onClick } = eventHandlers;
  return (
    <div className="flexbox column">
      {chartData.map(({ fill, x, y, tip }) => (
        <div className="clickable flexbox column barchart" key={x} onClick={e => onClick(e, { datum: { x } })} title={tip} style={{ justifyContent: 'center' }}>
          <LinearProgress className={createColorClassName(fill)} variant="determinate" value={y} style={{ height: 8 }} />
        </div>
      ))}
    </div>
  );
};

const ChartContainer = ({ className, children, innerRef, style = {} }) => (
  <div className={className} ref={innerRef} style={style}>
    {children}
  </div>
);

const BarChartContainer = ({ classes = {}, data, events, ...remainder }) => (
  <ChartContainer className={classes.wrapper}>
    <ChartLegend classes={classes} data={data} events={events} showIndicators={false} />
    <BarChart {...remainder} data={data} events={events} />
  </ChartContainer>
);

const PieChart = props => <VictoryPie {...props} padding={{ left: 0, right: 0, top: 0, bottom: 15 }} />;

const padding = 10;
const PieChartContainer = ({ classes = {}, ...chartProps }) => {
  const ref = useRef();
  let height;
  if (ref.current) {
    // use the widget height, remove the space the header takes up and account for widget padding (top + padding) + own padding for the chart
    height = ref.current.parentElement.offsetHeight - ref.current.parentElement.children[0].offsetHeight - 3 * padding;
  }
  return (
    <ChartContainer className={classes.wrapper} innerRef={ref} style={{ height }}>
      <ChartLegend {...chartProps} classes={classes} />
      {height && <PieChart {...chartProps} containerComponent={<VictoryContainer style={{ height }} />} />}
    </ChartContainer>
  );
};

const VictoryBarChartContainer = ({ classes = {}, ...chartProps }) => (
  <ChartContainer className={classes.wrapper}>
    <ChartLegend {...chartProps} classes={classes} />
    <VictoryBarChart {...chartProps} />
  </ChartContainer>
);

const chartTypeComponentMap = {
  [chartTypes.bar.key]: BarChartContainer,
  [`${chartTypes.bar.key}-alternative`]: VictoryBarChartContainer,
  [chartTypes.pie.key]: PieChartContainer
};

const initDistribution = ({ data, theme }) => {
  const { items, otherCount, total } = data;
  const numberOfItems = items.length > chartColorPalette.length ? chartColorPalette.length - 1 : items.length;
  const colors = chartColorPalette.slice(0, numberOfItems).reverse();
  const distribution = items.slice(0, colors.length).reduce(
    (accu, { key, count }, index) => [
      {
        x: key || '-',
        y: (count / total) * 100, //value,
        title: key || '-',
        tip: key || '-',
        fill: chartColorPalette[index],
        value: count
      },
      ...accu
    ],
    []
  );
  if (items.length > chartColorPalette.length || otherCount) {
    distribution.splice(0, 0, {
      x: seriesOther,
      title: 'Other',
      tip: 'Other',
      y: (otherCount / total) * 100,
      fill: chartColorPalette[chartColorPalette.length - 1],
      value: otherCount
    });
  }
  distribution.sort((pairA, pairB) => pairB.y - pairA.y);
  // y: formatValue(item.y, total)
  const totals = distribution.map(({ x, y }) => ({ value: total, x, y: 100 - y, fill: theme.palette.grey[400] }));
  return { distribution, totals };
};

interface DistributionReport {
  attribute: string;
  chartType: string;
  group: string;
  index: number;
  software: string;
  type: string;
}

type SoftwareLayer = {
  [key: string]: SoftwareLayer | string;
};

interface DistributionReportProps {
  onClick?: () => void;
  onSave: (selection: Partial<DistributionReport>) => void;
  selection?: Partial<DistributionReport>;
  software?: SoftwareLayer;
}

export const DistributionReport = ({ onClick, onSave, selection = {}, software: softwareTree }: DistributionReportProps) => {
  const { attribute, chartType = chartTypes.bar.key, group, index: reportIndex, software: softwareSelection } = selection;
  const software = softwareSelection || attribute || rootfsImageVersion;
  const [editing, setEditing] = useState(false);
  const [removing, setRemoving] = useState(false);
  const navigate = useNavigate();
  const { classes, theme } = useStyles();
  const reportsData = useSelector(getDeviceReports);
  const groupsById = useSelector(getGroupsById);
  const dispatch = useDispatch();
  const hasGroupDefinition = !!groupsById[group];
  const report = reportsData[reportIndex];
  const hasData = !isEmpty(report);

  useEffect(() => {
    if (group && !hasGroupDefinition) {
      return;
    }
    // this will retrieve device data and on repeat renders of the widget rely on stored data only - this increases the risk to miss data from the backend or
    // device initiated data changes, but assumes this is preferable over repeat queries to the backend causing rate limiting
    if (!hasData) {
      dispatch(getReportDataWithoutBackendSupport(reportIndex));
      return;
    }
    dispatch(updateReportData(reportIndex));
  }, [dispatch, reportIndex, group, hasData, hasGroupDefinition]);

  const { distribution, totals } = useMemo(() => {
    if (isEmpty(report)) {
      return { distribution: [], totals: [] };
    }
    return initDistribution({ data: report, theme });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(report), JSON.stringify(selection)]);

  const onSliceClick = useCallback(
    (e, { datum: { x: target } }) => {
      if (target === seriesOther) {
        return;
      }
      navigate(`/devices/accepted?inventory=${group ? `group:eq:${group}&` : ''}${ensureVersionString(software, attribute)}:eq:${target}`);
    },
    [attribute, group, navigate, software]
  );

  const toggleRemoving = () => setRemoving(toggle);

  const onToggleEditClick = () => setEditing(toggle);

  const onSaveClick = selection => {
    onSave(selection);
    setEditing(false);
  };

  const Chart = chartTypeComponentMap[chartType];
  const chartProps = {
    classes,
    data: distribution,
    domainPadding: 0,
    events: [{ target: 'data', eventHandlers: { onClick: onSliceClick } }],
    standalone: true,
    style: { data: { fill: ({ datum }) => datum.fill } },
    labels: () => null
  };
  const couldHaveDevices = !group || groupsById[group]?.deviceIds.length;
  if (removing) {
    return <RemovalWidget onCancel={toggleRemoving} onClick={onClick} />;
  }
  if (editing) {
    return <ChartEditWidget groups={groupsById} onSave={onSaveClick} onCancel={onToggleEditClick} selection={selection} software={softwareTree} />;
  }
  if (!report) {
    return <BaseWidget className="chart-widget flexbox centered" main={<Loader show />} />;
  }
  return (
    <div className="widget chart-widget">
      <div className="margin-bottom-small">
        <div className="flexbox space-between margin-bottom-small">
          <Header chartType={chartType} />
          <div className="flexbox center-aligned" style={{ zIndex: 1 }}>
            <IconButton onClick={onToggleEditClick} size="small">
              <Settings fontSize="small" />
            </IconButton>
            <IconButton onClick={toggleRemoving} size="small">
              <ClearIcon fontSize="small" />
            </IconButton>
          </div>
        </div>
        <div className="flexbox space-between slightly-smaller">
          <div>{softwareTitleMap[software] ? softwareTitleMap[software].title : software}</div>
          <div>{group || ALL_DEVICES}</div>
        </div>
      </div>
      {distribution.length || totals.length ? (
        <Chart {...chartProps} totals={totals} />
      ) : couldHaveDevices ? (
        <div className="muted flexbox centered">There are no devices that match the selected criteria.</div>
      ) : (
        <Loader show={true} />
      )}
    </div>
  );
};

export default DistributionReport;
