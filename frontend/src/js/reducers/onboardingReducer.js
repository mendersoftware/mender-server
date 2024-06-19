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

export const initialState = {
  approach: null,
  complete: false,
  deviceType: null,
  demoArtifactPort: 85,
  progress: null,
  showTips: null,
  showTipsDialog: false
};

const userReducer = (state = initialState, action) => {
  switch (action.type) {
    case OnboardingConstants.SET_ONBOARDING_STATE:
      return {
        ...state,
        ...action.value
      };
    case OnboardingConstants.SET_DEMO_ARTIFACT_PORT:
      return {
        ...state,
        demoArtifactPort: action.value
      };
    case OnboardingConstants.SET_SHOW_ONBOARDING_HELP:
      return {
        ...state,
        showTips: action.show
      };
    case OnboardingConstants.SET_SHOW_ONBOARDING_HELP_DIALOG:
      return {
        ...state,
        showTipsDialog: action.show
      };
    case OnboardingConstants.SET_ONBOARDING_COMPLETE:
      return {
        ...state,
        complete: action.complete
      };
    case OnboardingConstants.SET_ONBOARDING_PROGRESS:
      return {
        ...state,
        progress: action.value
      };
    case OnboardingConstants.SET_ONBOARDING_DEVICE_TYPE:
      return {
        ...state,
        deviceType: action.value
      };
    case OnboardingConstants.SET_ONBOARDING_APPROACH:
      return {
        ...state,
        approach: action.value
      };
    default:
      return state;
  }
};

export default userReducer;
