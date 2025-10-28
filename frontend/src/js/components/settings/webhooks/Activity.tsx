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
import { SetStateAction, useEffect, useRef, useState } from 'react';

import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { ClassesOverrides } from '@northern.tech/common-ui/List';
import Pagination from '@northern.tech/common-ui/Pagination';
import { Event } from '@northern.tech/store/api/types';
import { DEVICE_LIST_DEFAULTS, TIMEOUTS, Webhook } from '@northern.tech/store/constants';

import { WebhookColumns } from './Management';

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

interface WebhookActivityProps extends ClassesOverrides {
  columns: WebhookColumns;
  events?: Event[] | undefined;
  eventsTotal: number;
  getWebhookEvents: () => void;
  setSelectedEvent: SetStateAction<Event | undefined>;
  webhook: Webhook;
}

const WebhookActivity = ({ classes, columns, events = [], eventsTotal, getWebhookEvents, setSelectedEvent, webhook }: WebhookActivityProps) => {
  const [page, setPage] = useState(defaultPage);
  const tableRef = useRef();
  const timer = useRef<ReturnType<typeof setInterval> | undefined>();

  useEffect(() => {
    clearInterval(timer.current);
    timer.current = setInterval(() => getWebhookEvents({ page, perPage: defaultPerPage }), TIMEOUTS.refreshDefault);
    getWebhookEvents({ page, perPage: defaultPerPage });
    return () => {
      clearInterval(timer.current);
    };
  }, [getWebhookEvents, page]);

  const mappedColumns = columns.map(column => ({ ...column, extras: { webhook, classes } }));

  if (!events.length) {
    return <div className="margin-top-large flexbox centered disabled">No webhook activity yet.</div>;
  }

  return (
    <>
      <DetailsTable columns={mappedColumns} items={events} onItemClick={setSelectedEvent} tableRef={tableRef} />
      {eventsTotal > defaultPerPage && (
        <Pagination
          className="margin-top-none"
          count={eventsTotal ? eventsTotal : defaultPerPage}
          showCountInfo={false}
          rowsPerPageOptions={[defaultPerPage]}
          page={page}
          rowsPerPage={defaultPerPage}
          onChangePage={setPage}
        />
      )}
    </>
  );
};

export default WebhookActivity;
