// Copyright 2026 Northern.tech AS
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
import { rolloutModes } from './constants';
import { devicesToPercentage, formatDeviceCount, getPhaseDeviceCount, getPhaseMessages, getRemainder, percentageToDevices, toPhaseDescription } from './utils';

describe('getPhaseDeviceCount', () => {
  it('works with empty attributes', () => {
    expect(getPhaseDeviceCount(120, 10, 20, false)).toEqual(12);
    expect(getPhaseDeviceCount(120, 10, 20, true)).toEqual(12);
    expect(getPhaseDeviceCount(120, null, 20, true)).toEqual(24);
    expect(getPhaseDeviceCount(120, null, 20, false)).toEqual(24);
    expect(getPhaseDeviceCount(undefined, null, 20, false)).toEqual(0);
  });
});

describe('getRemainder', () => {
  it('calculates remainder percent correctly', () => {
    const phases = [
      { batch_size: 10, not: 'interested' },
      { batch_size: 10, not: 'interested' },
      { batch_size: 10, not: 'interested' }
    ];
    expect(getRemainder({ phases, rolloutMode: rolloutModes.percentage.key })).toEqual(80);
    expect(
      getRemainder({
        phases: [
          { batch_size: 10, not: 'interested' },
          { batch_size: 90, not: 'interested' }
        ],
        rolloutMode: rolloutModes.percentage.key
      })
    ).toEqual(90);
    expect(
      getRemainder({
        phases: [
          { batch_size: 10, not: 'interested' },
          { batch_size: 95, not: 'interested' }
        ],
        rolloutMode: rolloutModes.percentage.key
      })
    ).toEqual(90);
    expect(
      getRemainder({
        phases: [
          { batch_size: 50, not: 'interested' },
          { batch_size: 55, not: 'interested' },
          { batch_size: 95, not: 'interested' }
        ],
        rolloutMode: rolloutModes.percentage.key
      })
    ).toEqual(0);
  });
  it('calculates remainder devices correctly', () => {
    expect(
      getRemainder({ rolloutMode: rolloutModes.device_count.key, phases: [{ batch_size_devices: 10 }, { batch_size_devices: 20 }, {}], numberDevices: 100 })
    ).toEqual(70);
    expect(getRemainder({ rolloutMode: rolloutModes.device_count.key, phases: [{ batch_size_devices: 50 }, {}], numberDevices: 100 })).toEqual(50);
    expect(getRemainder({ rolloutMode: rolloutModes.device_count.key, phases: [{}], numberDevices: 100 })).toEqual(100);
    expect(
      getRemainder({ rolloutMode: rolloutModes.device_count.key, phases: [{ batch_size_devices: 90 }, { batch_size_devices: 20 }, {}], numberDevices: 100 })
    ).toEqual(0);
  });
});

describe('formatDeviceCount', () => {
  it('returns 0 for negative or NaN inputs', () => {
    expect(formatDeviceCount(-5)).toEqual('0');
    expect(formatDeviceCount(NaN)).toEqual('0');
    expect(formatDeviceCount(Infinity)).toEqual('0');
  });
  it('returns plain numbers below 1000', () => {
    expect(formatDeviceCount(0)).toEqual('0');
    expect(formatDeviceCount(1)).toEqual('1');
    expect(formatDeviceCount(999)).toEqual('999');
  });
  it('formats thousands with one decimal', () => {
    expect(formatDeviceCount(1000)).toEqual('1K');
    expect(formatDeviceCount(1500)).toEqual('1.5K');
    expect(formatDeviceCount(9999)).toEqual('9.9K');
  });
  it('formats tens of thousands without decimal', () => {
    expect(formatDeviceCount(10000)).toEqual('10K');
    expect(formatDeviceCount(50000)).toEqual('50K');
    expect(formatDeviceCount(999999)).toEqual('999K');
  });
  it('formats millions', () => {
    expect(formatDeviceCount(1000000)).toEqual('1M');
    expect(formatDeviceCount(1500000)).toEqual('1.5M');
    expect(formatDeviceCount(9900000)).toEqual('9.9M');
    expect(formatDeviceCount(10000000)).toEqual('10M');
    expect(formatDeviceCount(15000000)).toEqual('15M');
  });
});

describe('getPhaseMessages', () => {
  it('returns empty array for empty phases', () => {
    expect(getPhaseMessages({ phases: [], phaseIndex: 0, deploymentDeviceCount: 100, rolloutMode: rolloutModes.percentage.key, isDynamic: false })).toEqual([]);
  });

  describe('percentage mode', () => {
    it('flags percentage out of range', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size: 0 }, {}],
        phaseIndex: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.percentage.key,
        isDynamic: false
      });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('between 1%');
    });
    it('flags percentage that rounds to 0 devices', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size: 1 }, {}],
        phaseIndex: 0,
        deploymentDeviceCount: 2,
        rolloutMode: rolloutModes.percentage.key,
        isDynamic: false
      });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('rounds down to 0');
    });
    it('does not flag valid percentages', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size: 50 }, {}],
        phaseIndex: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.percentage.key,
        isDynamic: false
      });
      expect(msgs).toEqual([]);
    });
  });

  describe('device_count mode', () => {
    it('warns when batch exceeds group size for dynamic groups', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size_devices: 200 }, {}],
        phaseIndex: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: true
      });
      expect(msgs[0].severity).toEqual('warning');
      expect(msgs[0].message).toContain('exceeds the current target group size');
    });
    it('errors when batch exceeds group size for static groups', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size_devices: 200 }, {}],
        phaseIndex: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: false
      });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('exceeds total target group size');
    });
    it('flags phases with 0 devices', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size_devices: 0 }, {}],
        phaseIndex: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: false
      });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('at least 1 device');
    });
    it('flags last phase with remainder below 1', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size_devices: 100 }, {}],
        phaseIndex: 1,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: false
      });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('at least 1 device');
    });
    it('flags batch exceeding maxDevices', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size_devices: 60 }, {}],
        phaseIndex: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: false,
        maxDevices: 50
      });
      expect(msgs.some(m => m.severity === 'error')).toEqual(true);
      expect(msgs.find(m => m.message.includes('cannot exceed'))).toBeTruthy();
    });
    it('does not flag valid device counts', () => {
      const msgs = getPhaseMessages({
        phases: [{ batch_size_devices: 50 }, {}],
        phaseIndex: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: false
      });
      expect(msgs).toEqual([]);
    });
  });
});

describe('percentageToDevices', () => {
  it('converts percentage to device count', () => {
    expect(percentageToDevices(50, 100)).toEqual(50);
    expect(percentageToDevices(10, 1000)).toEqual(100);
    expect(percentageToDevices(33, 100)).toEqual(33);
  });
  it('returns at least 1 for small percentages', () => {
    expect(percentageToDevices(1, 50)).toEqual(1);
    expect(percentageToDevices(1, 1)).toEqual(1);
  });
  it('returns 0 when no devices', () => {
    expect(percentageToDevices(50, 0)).toEqual(0);
  });
});

describe('devicesToPercentage', () => {
  it('converts device count to percentage', () => {
    expect(devicesToPercentage(50, 100)).toEqual(50);
    expect(devicesToPercentage(100, 1000)).toEqual(10);
  });
  it('clamps to 1-99 range', () => {
    expect(devicesToPercentage(1, 10000)).toEqual(1);
    expect(devicesToPercentage(9999, 10000)).toEqual(99);
  });
  it('returns 10 when no devices', () => {
    expect(devicesToPercentage(50, 0)).toEqual(10);
  });
});

describe('toPhaseDescription', () => {
  it('describes custom percentage phases', () => {
    const phases = [{ batch_size: 30, delay: 2, delayUnit: 'hours' }, { batch_size: 70 }];
    const { phasesDescription } = toPhaseDescription(phases, 100);
    expect(phasesDescription).toContain('2 phases');
    expect(phasesDescription).toContain('30%');
    expect(phasesDescription).toContain('70%');
  });
  it('describes custom device-count phases', () => {
    const phases = [{ batch_size_devices: 30, delay: 2, delayUnit: 'hours' }, { batch_size_devices: 70 }];
    const { phasesDescription } = toPhaseDescription(phases, 100);
    expect(phasesDescription).toContain('2 phases');
    expect(phasesDescription).toContain('30');
    expect(phasesDescription).toContain('70');
  });
  it('computes remainder for last custom percentage phase', () => {
    const phases = [{ batch_size: 40 }, {}];
    const { phasesDescription } = toPhaseDescription(phases, 100);
    expect(phasesDescription).toContain('60%');
  });
  it('describes uniform percentage phases', () => {
    const phases = [{ isUniform: true, batch_size: 20, delay: 1, delayUnit: 'hours' }];
    const { phasesDescription } = toPhaseDescription(phases, 100);
    expect(phasesDescription).toContain('Uniform');
    expect(phasesDescription).toContain('20%');
  });
  it('describes uniform device-count phases', () => {
    const phases = [{ isUniform: true, batch_size_devices: 50, delay: 1, delayUnit: 'hours' }];
    const { phasesDescription } = toPhaseDescription(phases, 200);
    expect(phasesDescription).toContain('Uniform');
    expect(phasesDescription).toContain('50 devices');
  });
});
