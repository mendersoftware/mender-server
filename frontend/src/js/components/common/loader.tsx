// Copyright 2024 Northern.tech AS
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
import { CSSProperties } from 'react';

interface LoaderProps {
  show: boolean;
  fade?: boolean;
  small?: boolean;
  style?: CSSProperties;
  table?: boolean;
  waiting?: boolean;
}

export const Loader = (props: LoaderProps) => {
  const { fade, show, small, style, table, waiting } = props;

  const hideClass = fade ? 'hidden' : 'loaderContainer shrunk';
  const showClass = table ? 'miniLoaderContainer' : 'loaderContainer';
  return (
    <div style={style} className={show ? showClass : hideClass}>
      <div className={`${small ? 'small' : ''} ${waiting ? 'waiting-loader' : ''} loader`}>
        <span className="dot dot_1" />
        <span className="dot dot_2" />
        <span className="dot dot_3" />
        <span className="dot dot_4" />
      </div>
    </div>
  );
};

export default Loader;
