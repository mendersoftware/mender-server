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
import React, { useCallback, useEffect, useRef, useState } from 'react';

import { CloudDownload, Pause, PlayArrow, Refresh } from '@mui/icons-material';
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import XTerm from '@northern.tech/common-ui/xterm';
import { DEVICE_MESSAGE_PROTOCOLS as MessageProtocols, DEVICE_MESSAGE_TYPES as MessageTypes, TIMEOUTS, deviceConnect } from '@northern.tech/store/constants';
import { createFileDownload, toggle } from '@northern.tech/utils/helpers';
import { blobToString, byteArrayToString } from '@northern.tech/utils/sockethook';
import msgpack5 from 'msgpack5';
import Cookies from 'universal-cookie';

const cookies = new Cookies();
const MessagePack = msgpack5();

let socket = null;
let buffer = [];
let timer;

const useStyles = makeStyles()(theme => ({
  playArrow: { fontSize: '7rem', color: theme.palette.text.disabled }
}));

const generateHtml = (versions, content) => {
  const { fit, search, xterm } = Object.entries(versions).reduce((accu, [key, version]) => {
    accu[key] = version.match(/(?<version>\d.*)/).groups.version;
    return accu;
  }, {});
  return `
  <!DOCTYPE html>
  <html>
    <head>
      <link rel="stylesheet" href="https://unpkg.com/@xterm/xterm@${xterm}/css/xterm.css" />

      <script src="https://unpkg.com/@xterm/xterm@${xterm}/lib/xterm.js"></script>
      <script src="https://unpkg.com/@xterm/addon-search@${search}/lib/addon-search.js"></script>
      <script src="https://unpkg.com/@xterm/addon-fit@${fit}/lib/addon-fit.js"></script>
      <style type="text/css">
        body {
          display: grid;
          justify-items: center;
          max-width: 80vw;
          margin: 10vh auto;
          row-gap: 5vh;
          font-family: 'Segoe UI', Roboto, Ubuntu, 'Helvetica Neue', Helvetica, Arial, sans-serif;
        }
        h2 {
          color: #24444a;
        }
        button {
          background-color: #921267;
          padding: 1.3em 3.4em;
          color: #fff;
          font-weight: 700;
          text-transform: uppercase;
          border: 0;
          border-radius: 3px;
          cursor: pointer;
        }
        .disabled {
          background-color: lightgrey;
          opacity: 0.4;
        }
      </style>
    </head>
    <body>
      <img
        alt="mender-logo"
        src="data:image/png;base64,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"
      />
      <h2>Terminal playback</h2>
      <div id="terminal"></div>
      <div>
        <button id="start" onclick="handleStart()">Start</button>
        <button id="pause" class="disabled" disabled onclick="handlePause()">Pause</button>
        <button id="stop" class="disabled" disabled onclick="handleStop()">Stop</button>
      </div>
      <script>
        const byteArrayToString = body => String.fromCharCode(...body);
        const transfer = '${JSON.stringify(content.map(item => ({ delay: item.delay, content: btoa(JSON.stringify(item.content)) })))}';
        const content = JSON.parse(transfer);
        let contentIndex = 0;
        let timer;
        const term = new Terminal();
        const fitAddon = new FitAddon.FitAddon();
        const searchAddon = new SearchAddon.SearchAddon();
        term.loadAddon(searchAddon);
        term.loadAddon(fitAddon);
        term.open(document.getElementById('terminal'));
        fitAddon.fit();
        const startButton = document.getElementById('start');
        const pauseButton = document.getElementById('pause');
        const stopButton = document.getElementById('stop');

        const resetPlayer = () => {
          contentIndex = 0;
          term.reset()
          startButton.toggleAttribute('disabled');
          pauseButton.toggleAttribute('disabled');
          stopButton.toggleAttribute('disabled');
          pauseButton.classList.toggle('disabled');
          stopButton.classList.toggle('disabled');
          startButton.classList.toggle('disabled');
        }

        const processContent = () => {
          if (contentIndex === content.length) {
            return handlePause();
          }
          const item = content[contentIndex];
          contentIndex += 1;
          let delay = 1;
          if (item.delay) {
            delay = item.delay;
          } else if (item.content) {
            const buffer = JSON.parse(atob(item.content));
            term.write(byteArrayToString(buffer.data || []))
          }
          timer = setTimeout(processContent, delay)
        };

        const handleStart = () => {
          contentIndex = 0;
          startButton.toggleAttribute('disabled');
          pauseButton.toggleAttribute('disabled');
          stopButton.toggleAttribute('disabled');
          startButton.classList.toggle('disabled');
          pauseButton.classList.toggle('disabled');
          stopButton.classList.toggle('disabled');
          timer = setTimeout(processContent, 1);
        };

        const handlePause = () => {
          startButton.toggleAttribute('disabled');
          pauseButton.toggleAttribute('disabled');
          startButton.classList.toggle('disabled');
          pauseButton.classList.toggle('disabled');
          clearTimeout(timer);
        };

        const handleStop = () => {
          clearTimeout(timer);
          resetPlayer();
        };
      </script>
    </body>
  </html>`;
};

export const TerminalPlayer = ({ className, item, sessionInitialized, token }) => {
  const xtermRef = useRef({ terminal: React.createRef(), terminalRef: React.createRef() });
  const [socketInitialized, setSocketInitialized] = useState(false);
  const [bufferIndex, setBufferIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [wasStarted, setWasStarted] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [isLoadingSession, setIsLoadingSession] = useState(true);
  const [fitTrigger, setFitTrigger] = useState(false);

  const { classes } = useStyles();

  useEffect(() => {
    if (!sessionInitialized) {
      return;
    }
    cookies.set('JWT', token, { path: '/', secure: true, sameSite: 'strict', maxAge: 5 });
    socket = new WebSocket(`wss://${window.location.host}${deviceConnect}/sessions/${item.meta.session_id[0]}/playback`);
    socket.onopen = () => setSocketInitialized(true);
  }, [item.meta.session_id, sessionInitialized, token]);

  useEffect(() => {
    if (!socketInitialized) {
      return;
    }
    buffer = [];
    socket.onmessage = event =>
      blobToString(event.data).then(data => {
        const {
          hdr: { proto, typ, props = {} },
          body
        } = MessagePack.decode(data);
        if (proto !== MessageProtocols.Shell) {
          return;
        }
        clearTimeout(timer);
        timer = setTimeout(() => setIsLoadingSession(false), TIMEOUTS.oneSecond);
        switch (typ) {
          case MessageTypes.Shell:
            return buffer.push({ content: body });
          case MessageTypes.Delay:
            return buffer.push({ delay: props.delay_value });
          default:
            break;
        }
      });
  }, [socketInitialized]);

  useEffect(() => {
    if (isPlaying && bufferIndex < buffer.length) {
      if (bufferIndex === 0) {
        xtermRef.current.terminal.current.reset();
      }
      if (buffer[bufferIndex].content) {
        xtermRef.current.terminal.current.write(byteArrayToString(buffer[bufferIndex].content));
        setTimeout(() => setBufferIndex(bufferIndex + 1), 20);
      }
      if (buffer[bufferIndex].delay) {
        setTimeout(() => {
          setBufferIndex(bufferIndex + 1);
        }, buffer[bufferIndex].delay);
      }
    } else if (!isPaused) {
      resetPlayer();
    }
  }, [bufferIndex, isPaused, isPlaying]);

  const resetPlayer = () => {
    setIsPlaying(false);
    setBufferIndex(0);
  };

  const onTogglePlayClick = useCallback(() => {
    if (!wasStarted) {
      setWasStarted(true);
      return setTimeout(() => {
        setFitTrigger(toggle);
        xtermRef.current.terminal.current.focus();
        setIsPlaying(!isPlaying);
      }, TIMEOUTS.debounceShort);
    }
    setIsPaused(isPlaying);
    setIsPlaying(!isPlaying);
  }, [isPlaying, wasStarted]);

  const onReplayClick = () => {
    resetPlayer();
    setIsPlaying(true);
  };

  const onDownloadClick = () => {
    // eslint-disable-next-line no-undef
    const text = generateHtml({ fit: XTERM_FIT_VERSION, search: XTERM_SEARCH_VERSION, xterm: XTERM_VERSION }, buffer);
    createFileDownload(text, 'terminalsession.html', token);
  };

  return (
    <div className={`${className} `}>
      <div className="relative">
        <XTerm className="xterm-min-screen" triggerResize={fitTrigger} xtermRef={xtermRef} />
        {!wasStarted && (
          <div
            className="flexbox centered clickable"
            style={{ background: 'black', width: '100%', height: '100%', position: 'absolute', top: 0, zIndex: 10 }}
            onClick={onTogglePlayClick}
          >
            <PlayArrow className={classes.playArrow} />
          </div>
        )}
      </div>
      <div className="flexbox margin-top-small margin-bottom-small">
        <Button color="primary" onClick={onTogglePlayClick} startIcon={isPlaying ? <Pause /> : <PlayArrow />}>
          {isPlaying ? 'Pause' : 'Play'}
        </Button>
        <Button color="primary" onClick={onReplayClick} disabled={isPlaying} startIcon={<Refresh />}>
          Replay
        </Button>
        <Button color="primary" onClick={onDownloadClick} startIcon={<CloudDownload />} disabled={isLoadingSession}>
          Download
        </Button>
      </div>
    </div>
  );
};

export default TerminalPlayer;
