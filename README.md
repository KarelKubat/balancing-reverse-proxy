# balancing-reverse-proxy

<!-- toc -->
- [Invocation](#invocation)
  - [Minimal invocation](#minimal-invocation)
  - [When does <code>balancing-reverse-proxy</code> return a response to the client?](#when-does-balancing-reverse-proxy-return-a-response-to-the-client)
  - [Fanning out](#fanning-out)
  - [Other options](#other-options)
- [A simple test](#a-simple-test)
  - [DIY](#diy)
  - [Provided test shellscript](#provided-test-shellscript)
<!-- /toc -->

`balancing-reverse-proxy` is a reverse HTTP proxy, but one which allows configuring multiple endpoints (back ends). A request that arrives at the proxy is resolved at these endpoints, and the first usable response is returned to the client.

This setup is useful in situations where different HTTP(s) servers exist (typically API servers), but which are "flakey" in their processing. To improve the chance that a "good" response is collected, a client can forward their request to `balancing-reverse-proxy` to have the request processed on one or more  HTTP servers. The client gets the first correct response.

## Invocation

### Minimal invocation

`balancing-reverse-proxy` needs at least one flag: `-endpoints` (or shortended, `-e`). This is a comma-separated list of backends to call. Example: 

- Imagine that an API that your client wants to consume is hosted on `https://one.com/apis`. There is a mirror at `https://two.com/public-apis`.
- The endpoints flag then becomes: `-e https://one.com/apis,https://two.com/public-apis`

### When does `balancing-reverse-proxy` return a response to the client?

The "fitness" of an endpoint's response is determined by the status that an end point sends. The default is that any HTTP status in the 100's, 200's, 300's or 400's range is a valid outcome and is returned to the client. An endpoint's response is only discarded when the HTTP status is in the 500's range. This can be changed using the flag `-terminal-responses` (shorthand `-t`).

Flag `-terminal-responses` is a comma-separated list of "hundreds" and defaults to `100,200,300,400`. Each number must be a multiple of a hundred, and means that **all** HTTP statuses in the range of that number, up to and including +99, indicate that such an endpoint response is eligible for the client.

For example:

- Assume that some APIs are hosted on `https://one.com/apis`. There is a mirror at `https://two.com/public-apis` but it's incomplete: for some calls it will return a status 400 (not found). If that happens then `balancing-reverse-proxy` should **not** take that as a terminating response, but should use what `one.com` returns.
- The flag then becomes: `-terminal-responses 100,200,300`. That means that once an endpoint returns 100-199, 200-299 and 300-399 (not all of these exist in the wild), that endpoint's response will be passed to the client, and other responses will be discarded.

### Fanning out

Presence of the flag `-fanout` forwards a client's request to **all** endpoints in parallel. The first usable response is returned to the client, the other responses are discarded. Flag `-f` can be used when you care about latency and you want to hedge the calls to the endpoints.

The default is not to fan out: the request is sent to the first endpoint, then if needed to the second one, and so on. This mode is useful when e.g. the endpoints host per-tick paid APIs, and you don't want to needlessly call them.

### Other options

By default actions are logged to `stdout` with a date and time stamp. More logging options can be set using the flags `-log-*`, e.g., `-log-file` to append to a file, `-log-msec` to include a microsecond stamp. Run `balancing-reverse-proxy` without arguments to see the flags.

To send the output to the system logs, the utility `logger` can be used. Example:

```shell
# Logger supplies the date and time, so that balancing-reverse-proxy should not also
# send it. The invocation `logger -t $TAG generates a string tag. The output will probably
# go to /var/log/user.log (depending on your configuration).
# The placeholder $OTHER_FLAGS must at a minimum define the endpoints.
balancing-reverse-proxy -log-time=false -log-date=false \
  $OTHER_FLAGS | logger -t balancing-reverse-proxy
```

To see other flags, just start `balancing-reverse-proxy` without arguments. A list will be shown.

## A simple test

### DIY

1. Start in one terminal a dummy HTTP server on port 8000. This server will at random serve errors or not, and will delay the response by a random duration up to 1 second.

    ```shell
    go run -- dummy-http-server/main.go
    ```

2. Start in another terminal the balancer on port 8080. Define the same dummy endpoint a few times:

    ```shell
    go run  --  balancing-reverse-proxy.go -fanout \
      --endpoints http://localhost:8000,http://localhost:8000,http://localhost:8000,http://localhost:8000

    configuring endpoint 0: "http://localhost:8000"
    configuring endpoint 1: "http://localhost:8000"
    configuring endpoint 2: "http://localhost:8000"
    configuring endpoint 3: "http://localhost:8000"
    starting on: ":8080"
    ```

3. Start in yet another terminal `curl` to repeatedly hit the balancer. The following Bash command will run until you hit ^C:

    ```shell
    while [ true ] ; do curl -v http://localhost:8080; done
    ```

The log on terminal 2 (where the balancer runs) will show what's happening. Example:

```plain
request / served in 983.247916ms by endpoint 1
request / served in 908.07625ms by endpoint 1
request / served in 336.30325ms by endpoint 0
failed to return a valid answer, returning 500
```

The first three requests were served by endpoints 0 and 1. The last request indicates that no endpoint successfully fulfilled the request and `balancing-reverse-proxy` replied to the client with a 500 Server Error.

### Provided test shellscript

For a more elaborate test start `dummy-http-server/multi-test.sh`. This is is a test that stops after a few seconds and has the following parts:

- The test starts three local HTTP servers on the ports 8000, 8001 and 8002. These send a oneline response with respectively `Hello, world from localhost:8000`, or `from localhost:8001` or `from localhost:8002`. The web servers stop after 5 seconds.
- The test starts a `balancing-http-proxy` on port 8080 with the above servers as endpoints. The proxy also stops after 5 seconds.
- Next, `curl` repeatedly calls the proxy on `http://localhost:8080`, or stops when no valid response is received.

Below is sample output. New runs will produce other outputs given that parallel fanout to endpoints is not deterministic.

```plain
Hello, world from localhost:8000/
Hello, world from localhost:8001/
Hello, world from localhost:8000/
Hello, world from localhost:8000/
Hello, world from localhost:8002/
Hello, world from localhost:8001/
stopping server on "localhost:8000" after 5s
Hello, world from localhost:8001/
Hello, world from localhost:8002/
Hello, world from localhost:8001/
Hello, world from localhost:8002/
stopping server on "localhost:8002" after 5s
Hello, world from localhost:8001/
Hello, world from localhost:8001/
Hello, world from localhost:8001/
stopping server on "localhost:8001" after 5s
stopping balancing-reverse-proxy after 5s
curl: (7) Failed to connect to localhost port 8080 after 0 ms: Couldn't connect to server
```

- At first, the proxy serves to `curl` whatever appears first: the response from the web server on port 8000, or 8001, or 8002.
- On line 7 the webserver on port 8000 stops (this is at the 5 second mark). After that, `curl` only receives respnses from the servers on port 8001 and 8002.
- Line 12 shows that the web server on port 8002 also stops. After that, `curl` only receives responses from the server on port 8001.
- Near the end both the last web server on port 8001 and the proxy stop.

The test also creates a log file `/tmp/balancing-reverse-proxy-$$.log` (`$$` being a process ID). This log shows the corresponding view from the proxy. For example, at the time when only the webserver on port 8001 is up (last three `Hello World` lines above), the log shows that two endpoints are down, and that the request is successfully served from the remaining one.

```plain
http: proxy error: dial tcp [::1]:8000: connect: connection refused
request / served in 807.125µs by endpoint 1
http: proxy error: dial tcp [::1]:8002: connect: connection refused
```
