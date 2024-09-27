// Copyright 2022 Northern.tech AS
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
import { useCallback, useEffect, useRef, useState } from 'react';

import { DEVICE_MESSAGE_PROTOCOLS as MessageProtocols, DEVICE_MESSAGE_TYPES as MessageTypes, TIMEOUTS, apiUrl } from '@northern.tech/store/constants';
import msgpack5 from 'msgpack5';
import Cookies from 'universal-cookie';

const cookies = new Cookies();

const MessagePack = msgpack5();

export const byteArrayToString = body => String.fromCharCode(...body);

export const blobToString = blob =>
  new Promise(resolve => {
    let fr = new FileReader();
    fr.onload = () => {
      resolve(fr.result);
    };
    fr.readAsArrayBuffer(blob);
  });

export const useSession = ({ onClose, onHealthCheckFailed, onMessageReceived, onNotify, onOpen, token }) => {
  const [sessionId, setSessionId] = useState();
  const healthcheckTimeout = useRef();
  const socketRef = useRef();
  const sendMessage = useCallback(({ typ, body, props }) => {
    if (!socketRef.current) {
      return;
    }
    const proto_header = { proto: MessageProtocols.Shell, typ, sid: socketRef.current.sessionId, props };
    const encodedData = MessagePack.encode({ hdr: proto_header, body });
    socketRef.current.send(encodedData);
  }, []);

  const close = useCallback(() => {
    if (!socketRef.current || socketRef.current?.readyState !== WebSocket.OPEN) {
      return;
    }
    sendMessage({ typ: MessageTypes.Stop, body: {}, props: {} });
    socketRef.current.close();
    setSessionId();
  }, [sendMessage]);

  const healthcheckFailed = useCallback(() => {
    onHealthCheckFailed();
    close();
  }, [close, onHealthCheckFailed]);

  const onSocketMessage = useCallback(
    event =>
      blobToString(event.data).then(data => {
        const {
          hdr: { props = {}, proto, sid, typ },
          body
        } = MessagePack.decode(data);
        if (proto !== MessageProtocols.Shell) {
          return;
        }
        switch (typ) {
          case MessageTypes.New: {
            if (props.status == WebSocket.CLOSING) {
              onNotify(`Error: ${byteArrayToString(body)}`);
              setSessionId();
              return close();
            } else {
              clearTimeout(healthcheckTimeout.current);
              healthcheckTimeout.current = setTimeout(healthcheckFailed, 65 * TIMEOUTS.oneSecond);
              socketRef.current.sessionId = sid;
              return setSessionId(sid);
            }
          }
          case MessageTypes.Shell:
            return onMessageReceived(body);
          case MessageTypes.Stop:
            return close();
          case MessageTypes.Ping: {
            if (healthcheckTimeout.current) {
              clearTimeout(healthcheckTimeout.current);
            }
            sendMessage({ typ: MessageTypes.Pong });
            //
            const timeout = parseInt(props.timeout);
            if (timeout > 0) {
              healthcheckTimeout.current = setTimeout(healthcheckFailed, timeout * TIMEOUTS.oneSecond);
            }
            return;
          }
          default:
            break;
        }
      }),
    [close, healthcheckFailed, onMessageReceived, onNotify, sendMessage]
  );

  const onSocketError = useCallback(
    error => {
      onNotify(`WebSocket error: ${error.message}`);
      close();
      clearTimeout(healthcheckTimeout.current);
    },
    [close, onNotify]
  );

  const onSocketOpen = useCallback(() => {
    sendMessage({ typ: MessageTypes.New, props: {} });
    onOpen(true);
  }, [onOpen, sendMessage]);

  const onSocketClose = useCallback(
    e => {
      console.log('closing');
      onClose(e);
      close();
    },
    [close, onClose]
  );

  const connect = useCallback(
    deviceId => {
      const uri = `wss://${window.location.host}${apiUrl.v1}/deviceconnect/devices/${deviceId}/connect`;
      setSessionId();
      cookies.set('JWT', token, { path: '/', secure: true, sameSite: 'strict', maxAge: 5 });
      try {
        socketRef.current = new WebSocket(uri);
      } catch (error) {
        console.log(error);
      }
    },
    [token]
  );

  useEffect(() => {
    if (!socketRef.current) {
      return;
    }

    socketRef.current.addEventListener('close', onSocketClose);
    socketRef.current.addEventListener('error', onSocketError);
    socketRef.current.addEventListener('message', onSocketMessage);
    socketRef.current.addEventListener('open', onSocketOpen);

    return () => {
      socketRef.current.removeEventListener('close', onSocketClose);
      socketRef.current.removeEventListener('error', onSocketError);
      socketRef.current.removeEventListener('message', onSocketMessage);
      socketRef.current.removeEventListener('open', onSocketOpen);
    };
  }, [onSocketClose, onSocketError, onSocketMessage, onSocketOpen, socketRef.current?.readyState]);

  return [connect, sendMessage, close, socketRef.current?.readyState ?? WebSocket.CLOSED, sessionId];
};
