/* eslint-disable no-undef */
const { TextDecoder, TextEncoder } = require('node:util');
const { ReadableStream, TransformStream } = require('node:stream/web');
const { BroadcastChannel, MessagePort } = require('node:worker_threads');

Reflect.set(globalThis, 'BroadcastChannel', BroadcastChannel);
Reflect.set(globalThis, 'MessagePort', MessagePort);
Reflect.set(globalThis, 'ReadableStream', ReadableStream);
Reflect.set(globalThis, 'TextDecoder', TextDecoder);
Reflect.set(globalThis, 'TextEncoder', TextEncoder);
Reflect.set(globalThis, 'TransformStream', TransformStream);

const { Blob } = require('node:buffer');
const { fetch, FormData, Headers, Request, Response } = require('undici');

Reflect.set(globalThis, 'Blob', Blob);
Reflect.set(globalThis, 'fetch', { value: fetch, writable: true });
Reflect.set(globalThis, 'FormData', FormData);
Reflect.set(globalThis, 'Headers', Headers);
Reflect.set(globalThis, 'Request', Request);
Reflect.set(globalThis, 'Response', Response);
