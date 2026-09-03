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
import { delayDefaults, delayUnits, rolloutModes, rolloutPatterns } from './constants';
import {
  convertDefinitionsToMode,
  devicesToPercentage,
  formatDeviceCount,
  getDefinitionsRemainder,
  getPhaseDeviceCount,
  getPhaseMessages,
  getRemainder,
  parseInterval,
  parsePreviousPhases,
  percentageToDevices,
  toPhaseDescription
} from './utils';

describe('parseInterval', () => {
  it('returns defaults when no interval is provided', () => {
    expect(parseInterval()).toEqual(delayDefaults);
    expect(parseInterval(undefined)).toEqual(delayDefaults);
  });

  it('parses whole hours', () => {
    expect(parseInterval('7200s')).toEqual({ delay: 2, delayUnit: delayUnits.hours });
    expect(parseInterval('3600s')).toEqual({ delay: 1, delayUnit: delayUnits.hours });
  });

  it('parses whole minutes', () => {
    expect(parseInterval('300s')).toEqual({ delay: 5, delayUnit: delayUnits.minutes });
    expect(parseInterval('60s')).toEqual({ delay: 1, delayUnit: delayUnits.minutes });
  });

  it('parses whole days', () => {
    expect(parseInterval('86400s')).toEqual({ delay: 1, delayUnit: delayUnits.days });
    expect(parseInterval('172800s')).toEqual({ delay: 2, delayUnit: delayUnits.days });
  });

  it('prefers the largest whole-number unit', () => {
    expect(parseInterval('86400s')).toEqual({ delay: 1, delayUnit: delayUnits.days });
    expect(parseInterval('7200s')).toEqual({ delay: 2, delayUnit: delayUnits.hours });
    expect(parseInterval('5400s')).toEqual({ delay: 90, delayUnit: delayUnits.minutes });
    expect(parseInterval('90000s')).toEqual({ delay: 25, delayUnit: delayUnits.hours });
  });

  it('falls back to rounded hours when no unit component is >= 1', () => {
    expect(parseInterval('30s')).toEqual({ delay: 1, delayUnit: delayUnits.hours });
    expect(parseInterval('45s')).toEqual({ delay: 1, delayUnit: delayUnits.hours });
  });

  it('handles non-numeric input by using the default delay', () => {
    expect(parseInterval('abcs')).toEqual({ delay: 2, delayUnit: delayUnits.hours });
  });
});

describe('getPhaseDeviceCount', () => {
  it('works with empty attributes', () => {
    expect(getPhaseDeviceCount(120, 10, 20, false)).toEqual(12);
    expect(getPhaseDeviceCount(120, 10, 20, true)).toEqual(12);
    expect(getPhaseDeviceCount(120, 0, 20, true)).toEqual(24);
    expect(getPhaseDeviceCount(120, 0, 20, false)).toEqual(24);
    expect(getPhaseDeviceCount(undefined, 0, 20, false)).toEqual(0);
  });
});

describe('getRemainder', () => {
  it('calculates remainder percent of stored deployment phases correctly', () => {
    expect(
      getRemainder({ phases: [{ batch_size: 10 }, { batch_size: 10 }, { batch_size: 10 }], numberDevices: 100, rolloutMode: rolloutModes.percentage.key })
    ).toEqual(80);
    expect(getRemainder({ phases: [{ batch_size: 10 }, { batch_size: 90 }], numberDevices: 100, rolloutMode: rolloutModes.percentage.key })).toEqual(90);
    expect(getRemainder({ phases: [{ batch_size: 10 }, { batch_size: 95 }], numberDevices: 100, rolloutMode: rolloutModes.percentage.key })).toEqual(90);
    expect(
      getRemainder({ phases: [{ batch_size: 50 }, { batch_size: 55 }, { batch_size: 95 }], numberDevices: 100, rolloutMode: rolloutModes.percentage.key })
    ).toEqual(0);
  });
  it('calculates remainder devices of stored deployment phases correctly', () => {
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

describe('getDefinitionsRemainder', () => {
  it('leaves the derived final phase whatever the sized definitions do not claim', () => {
    expect(getDefinitionsRemainder({ phases: [{ batchSize: 10 }, { batchSize: 20 }], numberDevices: 100, rolloutMode: rolloutModes.percentage.key })).toEqual(
      70
    );
    expect(getDefinitionsRemainder({ phases: [], numberDevices: 100, rolloutMode: rolloutModes.percentage.key })).toEqual(100);
    expect(getDefinitionsRemainder({ phases: [{ batchSize: 10 }, { batchSize: 20 }], numberDevices: 50, rolloutMode: rolloutModes.device_count.key })).toEqual(
      20
    );
    expect(getDefinitionsRemainder({ phases: [{ batchSize: 60 }, { batchSize: 60 }], numberDevices: 100, rolloutMode: rolloutModes.device_count.key })).toEqual(
      0
    );
  });
});

describe('convertDefinitionsToMode', () => {
  it('converts batch sizes between percentages and device counts', () => {
    expect(convertDefinitionsToMode([{ batchSize: 10, delay: 2, delayUnit: 'hours' }], rolloutModes.device_count.key, 200)).toEqual([
      { batchSize: 20, delay: 2, delayUnit: 'hours' }
    ]);
    expect(convertDefinitionsToMode([{ batchSize: 20 }], rolloutModes.percentage.key, 200)).toEqual([{ batchSize: 10 }]);
  });
  it('leaves unsized definitions alone', () => {
    expect(convertDefinitionsToMode([{ delay: 2, delayUnit: 'hours' }], rolloutModes.device_count.key, 200)).toEqual([{ delay: 2, delayUnit: 'hours' }]);
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
  describe('percentage mode', () => {
    it('flags percentage out of range', () => {
      const msgs = getPhaseMessages({ batchSize: 0, remainder: 100, deploymentDeviceCount: 100, rolloutMode: rolloutModes.percentage.key, isDynamic: false });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('between 1%');
    });
    it('flags percentage that rounds to 0 devices', () => {
      const msgs = getPhaseMessages({ batchSize: 1, remainder: 99, deploymentDeviceCount: 2, rolloutMode: rolloutModes.percentage.key, isDynamic: false });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('rounds down to 0');
    });
    it('flags a final phase left without a share', () => {
      const msgs = getPhaseMessages({
        isFinal: true,
        remainder: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.percentage.key,
        isDynamic: false
      });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('at least 1 device');
    });
    it('does not flag valid percentages', () => {
      const msgs = getPhaseMessages({ batchSize: 50, remainder: 50, deploymentDeviceCount: 100, rolloutMode: rolloutModes.percentage.key, isDynamic: false });
      expect(msgs).toEqual([]);
    });
  });

  describe('device_count mode', () => {
    it('warns when batch exceeds group size for dynamic groups', () => {
      const msgs = getPhaseMessages({
        batchSize: 200,
        remainder: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: true
      });
      expect(msgs[0].severity).toEqual('warning');
      expect(msgs[0].message).toContain('exceeds the current target group size');
    });
    it('errors when batch exceeds group size for static groups', () => {
      const msgs = getPhaseMessages({
        batchSize: 200,
        remainder: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: false
      });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('exceeds total target group size');
    });
    it('flags phases with 0 devices', () => {
      const msgs = getPhaseMessages({ batchSize: 0, remainder: 0, deploymentDeviceCount: 100, rolloutMode: rolloutModes.device_count.key, isDynamic: false });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('at least 1 device');
    });
    it('flags a final phase with remainder below 1', () => {
      const msgs = getPhaseMessages({
        isFinal: true,
        remainder: 0,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: false
      });
      expect(msgs[0].severity).toEqual('error');
      expect(msgs[0].message).toContain('at least 1 device');
    });
    it('flags batch exceeding maxDevices', () => {
      const msgs = getPhaseMessages({
        batchSize: 60,
        remainder: 40,
        deploymentDeviceCount: 100,
        rolloutMode: rolloutModes.device_count.key,
        isDynamic: false,
        maxDevices: 50
      });
      expect(msgs.some(m => m.severity === 'error')).toEqual(true);
      expect(msgs.find(m => typeof m.message === 'string' && m.message.includes('cannot exceed'))).toBeTruthy();
    });
    it('does not flag valid device counts', () => {
      const msgs = getPhaseMessages({
        batchSize: 50,
        remainder: 50,
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

describe('parsePreviousPhases', () => {
  it('adopts stored custom phases as definitions without their closing remainder phase', () => {
    const stored = [
      { batch_size: 30, delay: 5, delayUnit: 'days', start_ts: 1 },
      { batch_size: 20, delay: 15, delayUnit: 'hours', start_ts: 1 },
      { batch_size: 50, start_ts: 2 }
    ];
    expect(parsePreviousPhases(stored)).toEqual({
      pattern: rolloutPatterns.custom.key,
      phases: [
        { batchSize: 30, delay: 5, delayUnit: 'days' },
        { batchSize: 20, delay: 15, delayUnit: 'hours' }
      ],
      rolloutMode: rolloutModes.percentage.key
    });
  });
  it('detects device count phases', () => {
    const stored = [{ batch_size_devices: 30, delay: 2, delayUnit: 'hours' }, { batch_size_devices: 70 }];
    expect(parsePreviousPhases(stored)).toEqual({
      pattern: rolloutPatterns.custom.key,
      phases: [{ batchSize: 30, delay: 2, delayUnit: 'hours' }],
      rolloutMode: rolloutModes.device_count.key
    });
  });
  it('turns a stored uniform phase into a single repeating definition', () => {
    const stored = [{ isUniform: true, batch_size: 20, delay: 1, delayUnit: 'hours' }];
    expect(parsePreviousPhases(stored)).toEqual({
      pattern: rolloutPatterns.uniform.key,
      phases: [{ batchSize: 20, delay: 1, delayUnit: 'hours' }],
      rolloutMode: rolloutModes.percentage.key
    });
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
