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
import React from 'react';

export const BaseWidget = ({ className = '', header, innerRef, main, onClick }) => (
  <div className={`widget ${className}`} onClick={onClick} ref={ref => (innerRef ? (innerRef.current = ref) : null)}>
    {!!header && <div className="flexbox widgetHeader">{header}</div>}
    {React.isValidElement(main) ? (
      <div className="widgetMainContent">{main}</div>
    ) : (
      <div className="flexbox column widgetMainContent">
        <div className="counter">{main.counter}</div>
        {main.targetLabel && <span className="link">{main.targetLabel}</span>}
      </div>
    )}
  </div>
);

export default BaseWidget;
