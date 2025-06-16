// Copyright 2025 Northern.tech AS
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

const variants = {
  email: 'support@mender.io',
  ourTeam: 'contact our team',
  support: 'contact support',
  salesTeam: 'contact our sales team'
};

export const SupportLink = ({ className = '', variant }: { className?: string; variant: keyof typeof variants | string }) => (
  <a className={className} href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
    {variants[variant] ?? variant}
  </a>
);
