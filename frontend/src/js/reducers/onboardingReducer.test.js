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
import * as OnboardingConstants from '../constants/onboardingConstants';
import reducer, { initialState } from './onboardingReducer';

describe('organization reducer', () => {
  it('should return the initial state', async () => {
    expect(reducer(undefined, {})).toEqual(initialState);
  });
  it('should handle SET_SHOW_ONBOARDING_HELP', async () => {
    expect(reducer(undefined, { type: OnboardingConstants.SET_SHOW_ONBOARDING_HELP, show: true }).showTips).toEqual(true);
    expect(reducer(initialState, { type: OnboardingConstants.SET_SHOW_ONBOARDING_HELP, show: false }).showTips).toEqual(false);
  });
  it('should handle SET_SHOW_ONBOARDING_HELP_DIALOG', async () => {
    expect(reducer(undefined, { type: OnboardingConstants.SET_SHOW_ONBOARDING_HELP_DIALOG, show: true }).showTipsDialog).toEqual(true);
    expect(reducer(initialState, { type: OnboardingConstants.SET_SHOW_ONBOARDING_HELP_DIALOG, show: false }).showTipsDialog).toEqual(false);
  });
  it('should handle SET_ONBOARDING_COMPLETE', async () => {
    expect(reducer(undefined, { type: OnboardingConstants.SET_ONBOARDING_COMPLETE, complete: true }).complete).toEqual(true);
    expect(reducer(initialState, { type: OnboardingConstants.SET_ONBOARDING_COMPLETE, complete: false }).complete).toEqual(false);
  });
  it('should handle SET_ONBOARDING_PROGRESS', async () => {
    expect(reducer(undefined, { type: OnboardingConstants.SET_ONBOARDING_PROGRESS, value: 'test' }).progress).toEqual('test');
    expect(reducer(initialState, { type: OnboardingConstants.SET_ONBOARDING_PROGRESS, value: 'test' }).progress).toEqual('test');
  });
  it('should handle SET_ONBOARDING_DEVICE_TYPE', async () => {
    expect(reducer(undefined, { type: OnboardingConstants.SET_ONBOARDING_DEVICE_TYPE, value: 'bbb' }).deviceType).toEqual('bbb');
    expect(reducer(initialState, { type: OnboardingConstants.SET_ONBOARDING_DEVICE_TYPE, value: 'rpi4' }).deviceType).toEqual('rpi4');
  });
  it('should handle SET_ONBOARDING_APPROACH', async () => {
    expect(reducer(undefined, { type: OnboardingConstants.SET_ONBOARDING_APPROACH, value: 'physical' }).approach).toEqual('physical');
    expect(reducer(initialState, { type: OnboardingConstants.SET_ONBOARDING_APPROACH, value: 'virtual' }).approach).toEqual('virtual');
  });
});
