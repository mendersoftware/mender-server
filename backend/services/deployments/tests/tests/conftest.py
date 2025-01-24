import pytest
import threading

from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Any


class OptimistHandler(BaseHTTPRequestHandler):
    """
    OptimistHandler answers 200 OK to all requests.
    """

    def _handle_request(self):
        self.send_response(200)
        self.end_headers()

    def __getattribute__(self, name: str, /) -> Any:
        """
        BaseHTTPRequestHandler will call do_<http method> for every incoming
        request. This handler will respond with an empty 200 OK always.
        """
        if name.startswith("do_"):
            return self._handle_request
        return super().__getattribute__(name)


@pytest.fixture(scope="function")
def http_mock(request, handler=OptimistHandler):
    """
    http_mock uses python http.server to implement a simple HTTP mock server.
    By default it answers 200 OK to all requests.
    To customize the handler, create a HTTPRequestHandler instance and use
    @pytest.mark.parameterize(handler=[YourHandler])
    """
    server = HTTPServer(("0.0.0.0", 0), handler)
    _, port = server.socket.getsockname()
    t = threading.Thread(target=server.serve_forever)
    try:
        t.start()
        yield f"http://{request.config.getoption("http_mock_host")}:{port}"
    finally:
        server.shutdown()
