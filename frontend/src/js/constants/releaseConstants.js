// Copyright 2019 Northern.tech AS
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
export const ARTIFACT_GENERATION_TYPE = { SINGLE_FILE: 'single_file' };
export const ARTIFACTS_REMOVED_ARTIFACT = 'ARTIFACTS_REMOVED_ARTIFACT';
export const ARTIFACTS_SET_ARTIFACT_URL = 'ARTIFACTS_SET_ARTIFACT_URL';
export const UPDATED_ARTIFACT = 'UPDATED_ARTIFACT';
export const RECEIVE_ARTIFACTS = 'RECEIVE_ARTIFACTS';
export const RECEIVE_RELEASE = 'RECEIVE_RELEASE';
export const RECEIVE_RELEASE_TAGS = 'RECEIVE_RELEASE_TAGS';
export const RECEIVE_RELEASE_TYPES = 'RECEIVE_RELEASE_TYPES';
export const RECEIVE_RELEASES = 'RECEIVE_RELEASES';
export const RELEASE_REMOVED = 'RELEASE_REMOVED';
export const SELECTED_RELEASE = 'SELECTED_RELEASE';
export const SET_RELEASES_LIST_STATE = 'SET_RELEASES_LIST_STATE';
export const currentArtifact = 'artifact_name';
export const rootfsImageVersion = 'rootfs-image.version';

export const softwareTitleMap = { [rootfsImageVersion]: { title: 'Root filesystem', priority: 0, key: rootfsImageVersion } };

export const ALL_RELEASES = 'All releases';
