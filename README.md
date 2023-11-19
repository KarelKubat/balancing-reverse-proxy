# balancing-reverse-proxy

<!-- toc -->
- [Invocation](#invocation)
  - [Minimal invocation](#minimal-invocation)
  - [When does <code>balancing-reverse-proxy</code> return a response to the client?](#when-does-balancing-reverse-proxy-return-a-response-to-the-client)
  - [Other options](#other-options)
- [This is work in progress](#this-is-work-in-progress)
- [A simple test](#a-simple-test)
<!-- /toc -->

`balancing-reverse-proxy` is a reverse HTTP proxy, but one which allows configuring multiple endpoints (back ends). A request that arrives at the proxy is resolved at all endpoints, and the first usable response is returned to the client.

This setup is useful in situations where different HTTP(s) servers exist (typically API servers), but which are "flakey" in their processing. To improve the chance that a "good" response is collected, a client can forward their request to `balancing-reverse-proxy` to have the request processed on one or more  HTTP servers. The client gets the first correct response.

## Invocation

### Minimal invocation

`balancing-reverse-proxy` needs at least one flag: `-endpoints` (or shortended, `-e`). This is a comma-separated list of backends to call. Example: 

- Imagine that an API that your client wants to consume is hosted on `https://one.com/apis`. There is a mirror at `https://two.com/public-apis`.
- The endpoints flag then becomes: `-e https://one.com/apis,https://two.com/public-apis`

### When does `balancing-reverse-proxy` return a response to the client?

The "fitness" of an endpoint's response is determined by the status that an end point sends. The default is that any HTTP status in the 100's, 200's, 300's or 400's range is a valid outcome and is returned to the client. An endpoint's response is only discarded when the HTTP status is in the 500's range. This can be changed using the flag `-terminal-responses` (shorthand `-t`).

Assume that some APIs are hosted on `https://one.com/apis`. There is a mirror at `https://two.com/public-apis` but it's incomplete: for some calls it will return a status 400 (not found). If that happens then `balancing-reverse-proxy` should **not** take that as a terminating response, but should wait what `one.com` returns.

Flag `-terminal-responses` is a comma-separated list of "hundreds", e.g., `100,200,300`. Each number must be a multiple of a hundred, and means that **all** HTTP statuses in the range of that number, up to and including +99, indicate that such an endpoint response is eligible for the client.

The flag then becomes: `-terminal-responses 100,200,300`. That means that once an endpoint is (a) first to return a response, (b) the HTTP status of the response is 100-199, 200-299 and 300-399 (not all of these exist in the wild), that endpoint's response will be passed to the client, and other responses will be discarded.

### Fanning out

Presence of the flag `-fanout` forwards a client's request to **all** endpoints in parallel. The first usable response is returned to the client, the other responses are discarded. Flag `-f` can be used when you care about latency and you want to hedge the calls to the endpoints.

The default is not to fan out: the request is sent to the first endpoint, then if needed to the second one, and so on. This mode is useful when e.g. the endpoints host per-tick paid APIs, and you don't want to needlessly call them.

### Other options

To see other flags, just start `balancing-reverse-proxy` without arguments. A list will be shown.

## This is work in progress

This version is a proof of concept. Future changes will include:

- Tests
- Possibly better modularization
- Better logging (this version logs all to `stdout`, maybe we need less logging)

## A simple test

1. Start in one terminal a dummy HTTP server on port 8000. This server will at random serve errors or not, and will delay the response by a random duration up to 1000ms.

    ```shell
    go run -- dummy-http-server/main.go -fanout
    ```

2. Start in another terminal the balancer on port 8080. Define the same dummy endpoint a few times:

    ```shell
    go run  --  balancing-reverse-proxy.go \
      --endpoints http://localhost:8000,http://localhost:8000,http://localhost:8000,http://localhost:8000

    2023/11/16 15:39:15 configuring endpoint 0: "http://localhost:8000"
    2023/11/16 15:39:15 configuring endpoint 1: "http://localhost:8000"
    2023/11/16 15:39:15 configuring endpoint 2: "http://localhost:8000"
    2023/11/16 15:39:15 configuring endpoint 3: "http://localhost:8000"
    2023/11/16 15:39:15 starting on: ":8080"
    ```

3. Start in yet another terminal `curl` to repeatedly hit the balancer. The following command will run until you hit ^C:

    ```shell
    while [ true ] ; do curl -v http://localhost:8080; done
    ```

The log on terminal 2 (where the balancer runs) will show what's happening. Example:

```
2023/11/16 15:43:37 serving request /
2023/11/16 15:43:37 endpoint "http://localhost:8000" sent 24 bytes (status: 500)
2023/11/16 15:43:37 endpoint "http://localhost:8000" sent 25 bytes (status: 200)
2023/11/16 15:43:37 endpoint returned a terminating status 200, discarding others
2023/11/16 15:43:37 endpoint "http://localhost:8000" sent 25 bytes (status: 500)
2023/11/16 15:43:37 endpoint "http://localhost:8000" sent 25 bytes (status: 200)
```

In this case the first endpoint that replied returned an internal server error (status 500). The second one returned OK, so that answer went to the client (in our test `curl`). The last two endpoints also sent replies, one with status 500 and the other with 200, but these were discarded.
