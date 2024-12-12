# Welcome to the backend integration tests

## How to run

```sh
make -C backend test-integration GOCOVERDIR=/tmp
```

by default the following tags will be used:

```
mendersoftware/iot-manager:test
```

## Test separation

 As of the moment of speaking we ported the compat test from the integration to the backend integration tests.
This was not that straightforward as it seemed, mainly because the pending devices presence disturbed
other tests. KJ suggested to run it in a separate pytest session. This also seemed straightforward.
 If you want to run your test in a separate pytest session, you need to add `+separate` interfix before `.py`.
See the below example for ``:

```sh
ls -1 backend/tests/integration/tests/
test_access.py
..
test_compat+separate.py
..
test_workflows.py
```

Further more if you need to have the additional docker-compose files included in your separated run
you have to define `COMPOSE_FILES_$yourtest_name`, see the example in the `backend/tests/integration/run`,
for `test_compat+separate.py` the variable is:

```sh
COMPOSE_FILES_test_compat
```

where the `test_compat` corresponds to thetest file name without `.py` and `+separate`.

