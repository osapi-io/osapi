# Copyright (c) 2024 John Dewey

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to
# deal in the Software without restriction, including without limitation the
# rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
# sell copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
# FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
# DEALINGS IN THE SOFTWARE.

PROGRAM="../main.go"
BATS_TEST_TIMEOUT=60
CONFIG="osapi.yaml"
export OSAPI_OSAPIFILE="${CONFIG}"

# Function to start the server
start_server() {
  # Start embedded NATS server (replaces external nats-server binary)
  go run ${PROGRAM} nats server start &
  sleep 2

  # Generate fresh admin token and update config
  TOKEN=$(go run ${PROGRAM} -j token generate \
    -r admin -u test@ci 2>/dev/null \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
  if [ -n "${TOKEN}" ]; then
    sed -i.bak "s|bearer_token:.*|bearer_token: ${TOKEN}|" "${CONFIG}"
    rm -f "${CONFIG}.bak"
  fi

  # Start API server
  go run ${PROGRAM} api server start &
  sleep 2

  # Start job worker
  go run ${PROGRAM} job worker start &
  sleep 3
}

# Function to stop the server
stop_server() {
  pkill -f "api server start" || true
  pkill -f "job worker start" || true
  pkill -f "nats server start" || true
  rm -rf .nats/
}
