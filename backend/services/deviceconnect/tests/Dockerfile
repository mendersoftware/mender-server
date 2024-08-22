FROM mendersoftware/mender-test-containers:acceptance-testing

# FIXME: The old and deprecated version of msgpack is a dependency of bravado (unused)
RUN pip uninstall -y msgpack && pip install msgpack==1.0.3

# install additional packages
COPY requirements.txt /requirements.txt
RUN pip install -r /requirements.txt
