#!/bin/sh

# Test duration
duration=5s

# Balancer's log
logfile=/tmp/balancing-reverse-proxy-$$.log

# Go to the top of this repo.
cd $(git rev-parse --show-toplevel) || exit 1

# Start 3 dummy servers.
go run -- dummy-http-server/main.go -address localhost:8000 -stop-after "$duration" \
  -delay-responding=false &
go run -- dummy-http-server/main.go -address localhost:8001 -stop-after "$duration" \
  -delay-responding=false &
go run -- dummy-http-server/main.go -address localhost:8002 -stop-after "$duration" \
  -delay-responding=false &

# Start the balancer to use the above dummies as endpoints.
go run -- balancing-reverse-proxy.go -f \
  -e http://localhost:8000/first,http://localhost:8001/second,http://localhost:8002/third \
  -s "$duration" \
  -log-file "$logfile" &
echo "NOTE: The balancer log is $logfile"

# Allow reading the above note, and allow background processes initialization.
sleep 2

# Consume from the balancer.
while [ true ]; do
    curl http://localhost:8080
    if [ $? -ne 0 ] ; then
        # allow started background processes to exit
        sleep 1
        exit
    fi
done
