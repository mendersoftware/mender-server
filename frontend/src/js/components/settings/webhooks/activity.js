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
import React, { useEffect, useMemo, useState } from 'react';

// material ui
import { ArrowDropDown as ArrowDropDownIcon, ArrowDropUp as ArrowDropUpIcon, Circle as CircleIcon } from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary, accordionClasses } from '@mui/material';
import { accordionSummaryClasses } from '@mui/material/AccordionSummary';
import { makeStyles } from 'tss-react/mui';

import { DEVICE_LIST_DEFAULTS } from '../../../constants/deviceConstants';
import { toggle } from '../../../helpers';
import Pagination from '../../common/pagination';
import Time from '../../common/time';

const useStyles = makeStyles()(theme => ({
  activityList: {
    display: 'flexbox',
    flexDirection: 'column',
    ['.header']: {
      padding: `0px ${theme.spacing(2)}`,
      marginBottom: theme.spacing()
    },
    [`.header, .${accordionSummaryClasses.content}`]: {
      display: 'grid',
      gridTemplateColumns: '1fr 2fr 1fr 50px'
    },
    [`.${accordionSummaryClasses.content}`]: {
      cursor: 'pointer',
      [`&>time, &>.trigger-type`]: {
        color: theme.palette.secondary.main
      },
      [`&>:last-child`]: {
        paddingRight: 'initial'
      }
    },
    [`.${accordionClasses.root}`]: {
      backgroundColor: theme.palette.background.default
    }
  },
  divider: { marginTop: theme.spacing(), marginBottom: theme.spacing() }
}));

const triggerMap = {
  'device-decommissioned': 'Device decommissioned',
  'device-provisioned': 'Device provisioned',
  'device-status-changed': 'Device status updated'
};

const iconStyles = {
  status: { fontSize: 12, marginRight: 8 },
  dropDown: { justifySelf: 'right', marginRight: 8 }
};

const DeliveryStatus = ({ entry, webhook = {} }) => {
  const { delivery_statuses = [] } = entry;

  const status = useMemo(() => {
    const status = delivery_statuses.find(status => status.integration_id === webhook.id) ?? delivery_statuses[0];
    if (status) {
      return { code: status.status_code, signal: status.success ? 'green' : 'red' };
    }
    return { code: 418, signal: 'disabled' };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(delivery_statuses), webhook.id]);

  return (
    <div className="flexbox center-aligned">
      <CircleIcon className={status.signal} style={iconStyles.status} />
      <div className={status.code >= 400 ? 'muted' : ''}>{status.code}</div>
    </div>
  );
};

const columns = [
  { key: 'created_ts', title: 'Time', render: ({ entry }) => <Time value={entry.time} /> },
  { key: 'trigger', title: 'Event trigger', render: ({ entry }) => <div className="trigger-type">{triggerMap[entry.type] ?? entry.type}</div> },
  { key: 'status', title: 'Status', render: ({ entry, webhook }) => <DeliveryStatus entry={entry} webhook={webhook} /> },
  {
    key: 'details',
    title: 'Details',
    render: ({ isExpanded }) => (isExpanded ? <ArrowDropUpIcon style={iconStyles.dropDown} /> : <ArrowDropDownIcon style={iconStyles.dropDown} />)
  }
];

const ListItem = ({ entry = {}, webhook }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const { data = {} } = entry;

  return (
    <Accordion square expanded={isExpanded}>
      <AccordionSummary onClick={() => setIsExpanded(toggle)}>
        {columns.map(({ key, render: Column }) => (
          <Column key={key} entry={entry} isExpanded={isExpanded} webhook={webhook} />
        ))}
      </AccordionSummary>
      <AccordionDetails>
        <code>{JSON.stringify(data)}</code>
      </AccordionDetails>
    </Accordion>
  );
};

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;
const WebhookActivity = ({ events = [], getWebhookEvents, eventTotal, webhook }) => {
  const { classes } = useStyles();
  const [page, setPage] = useState(defaultPage);

  useEffect(() => {
    getWebhookEvents({ page, perPage: defaultPerPage });
  }, [getWebhookEvents, page]);

  return (
    <div className={classes.activityList}>
      {events.length ? (
        <>
          <div className="header">
            {columns.map(({ key, title }) => (
              <div key={key}>{title}</div>
            ))}
          </div>
          <div className="body">
            {events.map(entry => (
              <ListItem key={entry.id} entry={entry} webhook={webhook} />
            ))}
            {eventTotal > defaultPerPage && (
              <Pagination
                className="margin-top-none"
                count={eventTotal ? eventTotal : defaultPerPage}
                showCountInfo={false}
                rowsPerPageOptions={[defaultPerPage]}
                page={page}
                rowsPerPage={defaultPerPage}
                onChangePage={setPage}
              />
            )}
          </div>
        </>
      ) : (
        <div className="margin-top-large flexbox centered disabled">No webhook activity yet.</div>
      )}
    </div>
  );
};

export default WebhookActivity;
