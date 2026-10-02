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
import type { Ref } from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';

import { Alert, Typography } from '@mui/material';

import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import LinedHeader from '@northern.tech/common-ui/LinedHeader';
import Pagination from '@northern.tech/common-ui/Pagination';
import { MaybeTime, Time } from '@northern.tech/common-ui/Time';
import { SynchronizedTwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { DEPLOYMENT_STATES } from '@northern.tech/store/constants';
import type { Deployment } from '@northern.tech/store/deploymentsSlice';
import { DEVICE_LIST_DEFAULTS } from '@northern.tech/utils/constants';
import { formatTime } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

import { phaseLimits, rolloutModes } from '../deployment-wizard/phases/constants';
import { getRemainder } from '../deployment-wizard/phases/utils';
import { SubstateProgressBar } from '../progress/SubstateProgressBar';
import type { DisplayablePhase } from '../progress/utils';
import { getDeploymentStartTime, getDisplayablePhases } from '../progress/utils';
import PhaseStatus from './PhaseStatus';

const { perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

const phaseColumns = [
  {
    key: 'phase',
    title: 'Phases',
    render: ({ id, isFinal }: DisplayablePhase) => (
      <>
        <div>Phase {id}</div>
        {isFinal && <Typography variant="caption">(Final phase)</Typography>}
      </>
    )
  },
  {
    key: 'size',
    title: 'Batch size',
    cellProps: { align: 'right' },
    render: ({ isGrowing, device_count, initial_batch_device_count }: DisplayablePhase) =>
      `${(initial_batch_device_count || device_count)?.toLocaleString()}${isGrowing ? '*' : ''}`
  },
  { key: 'startTs', title: 'Phase start time', cellProps: { align: 'right' }, render: ({ start_ts }: DisplayablePhase) => <MaybeTime value={start_ts} /> },
  {
    key: 'status',
    title: 'Status',
    cellProps: { className: 'padding-top-none padding-bottom-none', style: { minWidth: 200 } },
    render: (phase: DisplayablePhase) => <PhaseStatus phase={phase} />
  }
];

interface RolloutScheduleProps {
  deployment: Deployment;
  innerRef?: Ref<HTMLDivElement>;
  onAbort: (id: string) => void;
  onUpdateControlChange: (update: { states: Record<string, { action: string }> }) => void;
}

const getCurrentPhasePage = (phases: DisplayablePhase[], perPage: number): number => {
  const index = phases.findIndex(({ status }) => status === DEPLOYMENT_STATES.inprogress);
  return index < 0 ? 1 : Math.floor(index / perPage) + 1;
};

const getRolloutPatternSummary = (deployment: Deployment): string => {
  const { phases = [], uniform_phases } = deployment;
  if (uniform_phases) {
    const { batch_size, batch_size_devices = 0 } = uniform_phases;
    return batch_size ? `Uniform, ${batch_size}% per phase` : `Uniform, ${batch_size_devices} ${pluralize('device', batch_size_devices)} per phase`;
  }
  if (phases.length < 2) {
    return 'Standard (single phase)';
  }
  const prefix = `${phases.length} ${pluralize('phase', phases.length)}: `;
  if (phases.some(({ batch_size_devices }) => !!batch_size_devices)) {
    const lastSize = phases[phases.length - 1]?.batch_size_devices ?? 0;
    return `${prefix}${phases.map(({ batch_size_devices }) => batch_size_devices).join(', ')} ${pluralize('device', lastSize)}`;
  }
  const remainder = getRemainder({ phases, numberDevices: phaseLimits.fullBatchPercentage, rolloutMode: rolloutModes.percentage.key });
  return `${prefix}${phases.map(({ batch_size = 0 }, index) => `${index === phases.length - 1 ? remainder : batch_size}%`).join(', ')}`;
};

export const RolloutSchedule = ({ deployment, innerRef, onAbort, onUpdateControlChange }: RolloutScheduleProps) => {
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(defaultPerPage);
  const pagedDeploymentId = useRef<string | undefined>(undefined);
  const { filter, finished, id, max_devices = 0, phases = [], update_control_map } = deployment;
  const displayablePhases = useMemo(() => getDisplayablePhases(deployment), [deployment]);
  const isPaginated = displayablePhases.length > defaultPerPage;

  useEffect(() => {
    if (!isPaginated || pagedDeploymentId.current === id) {
      return;
    }
    pagedDeploymentId.current = id;
    setPage(getCurrentPhasePage(displayablePhases, defaultPerPage));
  }, [displayablePhases, id, isPaginated]);

  const onChangeRowsPerPage = (newPerPage: number) => {
    setPerPage(newPerPage);
    setPage(getCurrentPhasePage(displayablePhases, newPerPage));
  };

  const isCapped = !!(filter?.name && !!max_devices);
  const isMaybeGrowing = displayablePhases.some(({ isGrowing }) => isGrowing);
  const visiblePhases = isPaginated ? displayablePhases.slice((page - 1) * perPage, page * perPage) : displayablePhases;

  return (
    <>
      <LinedHeader className="margin-top-large" heading="Schedule details" ref={innerRef} />
      <SynchronizedTwoColumnData
        className="margin-bottom"
        data={{
          'Rollout pattern': getRolloutPatternSummary(deployment),
          'Start time': <Time value={formatTime(getDeploymentStartTime(deployment))} />,
          ...(finished ? { 'Finished time': <Time value={formatTime(finished)} /> } : {})
        }}
      />
      {phases.length <= 1 && update_control_map && (
        <SubstateProgressBar className="margin-bottom" deployment={deployment as Deployment} onAbort={onAbort} onUpdateControlChange={onUpdateControlChange} />
      )}
      <DetailsTable className="margin-bottom-none" columns={phaseColumns} items={visiblePhases} />
      {isPaginated && (
        <Pagination count={displayablePhases.length} onChangePage={setPage} onChangeRowsPerPage={onChangeRowsPerPage} page={page} rowsPerPage={perPage} />
      )}
      {isMaybeGrowing && (
        <Alert className="margin-top" severity="info">
          *This deployment targets a dynamic device group, so the final phase may adjust as devices change. The last phase stays active to keep all devices
          updated. <DocsTextLink id="dynamicDeployments" typographyProps={{ variant: 'inherit' }} />
        </Alert>
      )}
      {isCapped && (
        <Alert className="margin-top" severity="info">
          This deployment will finish at {max_devices.toLocaleString()} devices
        </Alert>
      )}
    </>
  );
};

export default RolloutSchedule;
