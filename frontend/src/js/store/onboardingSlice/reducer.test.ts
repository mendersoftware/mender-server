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
// @ts-nocheck
import reducer, { actions, initialState } from '.';

describe('organization reducer', () => {
  it('should return the initial state', async () => {
    expect(reducer(undefined, {})).toEqual(initialState);
  });
  it('should handle setOnboardingState', async () => {
    expect(reducer(undefined, { type: actions.setOnboardingState, payload: { foo: 'bar', showTips: true } }).showTips).toEqual(true);
    expect(reducer(initialState, { type: actions.setOnboardingState, payload: { foo: false } }).showTips).toEqual(null);
  });
  it('should handle SET_SHOW_ONBOARDING_HELP', async () => {
    expect(reducer(undefined, { type: actions.setShowOnboardingHelp, payload: true }).showTips).toEqual(true);
    expect(reducer(initialState, { type: actions.setShowOnboardingHelp, payload: false }).showTips).toEqual(false);
  });
  it('should handle SET_SHOW_ONBOARDING_HELP_DIALOG', async () => {
    expect(reducer(undefined, { type: actions.setShowDismissOnboardingTipsDialog, payload: true }).showTipsDialog).toEqual(true);
    expect(reducer(initialState, { type: actions.setShowDismissOnboardingTipsDialog, payload: false }).showTipsDialog).toEqual(false);
  });
  it('should handle SET_ONBOARDING_COMPLETE', async () => {
    expect(reducer(undefined, { type: actions.setOnboardingComplete, payload: true }).complete).toEqual(true);
    expect(reducer(initialState, { type: actions.setOnboardingComplete, payload: false }).complete).toEqual(false);
  });
  it('should handle SET_ONBOARDING_PROGRESS', async () => {
    expect(reducer(undefined, { type: actions.setOnboardingProgress, payload: 'test' }).progress).toEqual('test');
    expect(reducer(initialState, { type: actions.setOnboardingProgress, payload: 'test' }).progress).toEqual('test');
  });
  it('should handle SET_ONBOARDING_DEVICE_TYPE', async () => {
    expect(reducer(undefined, { type: actions.setOnboardingDeviceType, payload: 'bbb' }).deviceType).toEqual('bbb');
    expect(reducer(initialState, { type: actions.setOnboardingDeviceType, payload: 'rpi4' }).deviceType).toEqual('rpi4');
  });
  it('should handle SET_ONBOARDING_APPROACH', async () => {
    expect(reducer(undefined, { type: actions.setOnboardingApproach, payload: 'physical' }).approach).toEqual('physical');
    expect(reducer(initialState, { type: actions.setOnboardingApproach, payload: 'virtual' }).approach).toEqual('virtual');
  });
});
