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
import React, { useEffect } from 'react';

import { FitAddon } from '@xterm/addon-fit';
import { SearchAddon } from '@xterm/addon-search';
import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';

const searchAddon = new SearchAddon();
const fitAddon = new FitAddon();

const addons = [fitAddon, searchAddon];
const defaultOptions = { allowProposedApi: true, scrollback: 5000 };

export const Xterm = ({ className, customKeyEventHandler, options = {}, onResize, style, triggerResize, xtermRef, ...remainingProps }) => {
  /**
   * XTerm.js Terminal object.
   */
  // This is assigned in the setupTerminal() which is called from the constructor
  // terminal: Terminal
  // const terminal = ref.current.terminal.current;

  /**
   * The ref for the containing element.
   */
  // const terminalRef = ref.current.terminalRef.current;

  useEffect(() => {
    let { terminal, terminalRef } = xtermRef.current;
    // Setup the XTerm terminal.
    terminal.current = new Terminal({ ...defaultOptions, ...options });
    // Load addons
    addons.forEach(addon => terminal.current.loadAddon(addon));

    // Create Listeners
    Object.entries(remainingProps).map(([key, value]) => (value ? terminal.current[key](value) : undefined));

    // Add Custom Key Event Handler
    if (customKeyEventHandler) {
      terminal.current.attachCustomKeyEventHandler(customKeyEventHandler);
    }

    if (terminalRef.current) {
      // Creates the terminal within the container element.
      terminal.current.open(terminalRef.current);
    }
    return () => {
      // When the component unmounts dispose of the terminal and all of its listeners.
      // eslint-disable-next-line react-hooks/exhaustive-deps
      terminal.current.dispose();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [addons, customKeyEventHandler, JSON.stringify(options), Object.keys(remainingProps).join('')]);

  useEffect(() => {
    if (!xtermRef.current.terminalRef.current) {
      return;
    }
    try {
      fitAddon.fit();
      const { rows = 40, cols = 80 } = fitAddon.proposeDimensions() || {};
      if (onResize) {
        onResize({ rows, cols });
      }
    } catch {
      /* nothing really  */
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [triggerResize]);

  return <div className={className} ref={xtermRef.current.terminalRef} style={style} />;
};

export default Xterm;
