#!/usr/bin/python3
# Copyright 2021 Northern.tech AS
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.
import os
import threading
import logging
from contextlib import contextmanager
from concurrent import futures

import tornado.web
import tornado.ioloop


class MockRequestHandler(tornado.web.RequestHandler):
    log = logging.getLogger("MockRequestHandler")

    def initialize(self, method="GET", cb=None):
        self.method = method
        self.callback = cb

    def post(self, *args, **kwargs):
        self._run_handlers(*args, **kwargs)

    def get(self, *args, **kwargs):
        self._run_handlers(*args, **kwargs)

    def delete(self, *args, **kwargs):
        self._run_handlers(*args, **kwargs)

    def put(self, *args, **kwargs):
        self._run_handlers(*args, **kwargs)

    def patch(self, *args, **kwargs):
        self._run_handlers(*args, **kwargs)

    def _run_handlers(self, *args, **kwargs):
        if self.request.method != self.method or not self.callback:
            self.log.error("no handler for request %s", self.request)
            raise tornado.web.HTTPError(405)

        status, headers, body = self.callback(self.request, *args, **kwargs)
        # set status
        self.set_status(status)

        # set headers
        for hdr, val in headers.items():
            self.add_header(hdr, val)

        # push body
        if body:
            self.write(body)

        self.finish()


class MockServer:
    log = logging.getLogger("MockServer")

    def __init__(self, listen=("0.0.0.0", 9999), handlers=[]):
        th = []
        for method, path, cb in handlers:
            th.append((path, MockRequestHandler, {"method": method, "cb": cb}))

        self.app = tornado.web.Application(th)

        self.listen = listen
        self.thread = None
        self.loop = None

    def _start_loop_thread(self, sync):
        """Thread.run() callback that starts the loop and signals that it has started
        by setting `sync` to True"""
        self.log.info("starting test server in thread")

        host, port = self.listen[0], self.listen[1]

        self.loop = tornado.ioloop.IOLoop(make_current=True)
        self.log.info("listen on %s:%s", host, port)
        self.app.listen(port, address=host)

        self.log.info("starting loop")

        # sync with our caller
        self.loop.add_callback(lambda: sync.set_result(True))

        self.loop.start()

    def _stop_loop_thread(self):
        """In IO loop callback to stop the loop"""
        self.log.info("stopping test server")
        self.loop.stop()

    def run_thread_and_sync(self):
        """Start the server in a separate thread and wait for notification that it has
        successfuly started"""

        # create future so that we can sync with the IO loop running inside the
        # thread we started
        sync = futures.Future()

        self.thread = threading.Thread(target=self._start_loop_thread, args=(sync,))
        self.thread.start()

        # wait for the loop to be stated
        futures.wait([sync])

        self.log.info("test server thread started")

        return self

    def stop_thread_and_wait(self):
        """Stop the server running in separate thread and wait its IO loop to finish"""

        self.log.info("test server stop..")
        if not self.thread:
            return

        # run stop in loop context
        self.loop.add_callback(self._stop_loop_thread)
        # wait for thread
        self.thread.join()

        self.thread = None

        # cleanup the loop & close all file descriptors (NOTE: this will make
        # our server stop listening for connections)
        self.loop.close(all_fds=True)
        self.loop = None

    def wait(self):
        if not self.thread:
            return

        self.thread.join()


@contextmanager
def run_fake(listen_addr, handlers=[]):
    """run_fake acts as a context manager and can be used to create a mock
    server listening on `listen_addr`.

    `handlers` is a list of tuples: (<http-method>, <uri>, <callable>). Each
    callable takes a request argument (compatible with
    `tornado.httputil.HTTPServerRequest`) as its first argument and returns a
    tuple (<status-code>, <headers-dict>, <body>). You can use Tornado's
    routing facilities to have path elements become arguments to callbacks, see
    http://www.tornadoweb.org/en/stable/web.html for more details.

    """
    sp = listen_addr.split(":")
    host, port = sp[0], int(sp[1]) if len(sp) > 1 else 9999

    try:
        mock_server = MockServer((host, port), handlers=handlers)
        yield mock_server.run_thread_and_sync()
    finally:
        mock_server.stop_thread_and_wait()


if __name__ == "__main__":
    logging.basicConfig(level=logging.DEBUG)

    def verify_maybe(request):
        status = 400
        body = "Set `X-Please-Verify: true` to get 200 OK back"
        if request.headers.get("X-Please-Verify", None) == "true":
            status = 200
            body = "verified OK"
        return (status, {}, body)

    def get_tenant(request, tenant_id):
        """Example handler for GET operation on tenants. The method
        receives tenant ID as path argument. Further query arguments
        are accessed through `request.query_arguments`.
        """
        logging.info("arguments: %s", request.query_arguments)
        if not tenant_id:
            return (404, {}, "no tenant")
        return (200, {}, "tenant {:d}".format(int(tenant_id)))

    handlers = [
        (
            "POST",
            "/api/internal/v1/tenantadm/tenants/verify",
            lambda _: (200, {"Foo": "bar"}, ""),
        ),
        (
            "POST",
            "/api/internal/v1/tenantadm/tenants/verify/bad",
            lambda _: (401, {}, ""),
        ),
        ("POST", "/api/internal/v1/tenantadm/tenants/verify/maybe", verify_maybe),
        (
            "POST",
            "/api/internal/v1/tenantadm/tenants/login",
            lambda _: (200, {}, "token-token"),
        ),
        ("GET", "/api/internal/v1/tenantadm/tenant/(.*)", get_tenant),
    ]
    with run_fake("0.0.0.0:9999", handlers=handlers) as server:
        server.wait()
