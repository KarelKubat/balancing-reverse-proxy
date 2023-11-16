# balancing-reverse-proxy

<!-- toc -->
- [This is work in progress.](#this-is-work-in-progress)
- [A simple test](#a-simple-test)
<!-- /toc -->

`balancing-reverse-proxy` is a reverse HTTP proxy, but one which allows to configure multiple endpoints (back ends). A request that arrives at the proxy is forwarded to all endpoints, and the first usable response is returned to the client.

This setup is useful in situations where different HTTP servers exist (typically API servers), but which are "flakey" in their processing. To improve the chance that a "good" response is collected, a client can forward their request to `balancing-reverse-proxy`, which fans out to these HTTP servers. The client gets the first correct response.

The "goodness" of an endpoint's response is determined by the status that an end point sends. The default is that any HTTP status in the 100's, 200's, 300's or 400's range is a valid outcome and is returned to the client. An endpoint's response is only discarded when the HTTP status is in the 500's range. This can be changed using the flag `-terminal-responses`. Run `balancing-reverse-proxy` without any arguments to see the supported flags.

## This is work in progress.

This version is a proof of concept. Future changes will include:

- Tests
- Possibly better modularization
- Better logging (this version logs all to `stdout`, maybe we need less logging)

## A simple test

1. Start in one terminal a dummy HTTP server on port 8000. This server will at random serve errors or not, and will delay the response by a random duration up to 1000ms.

    ```shell
    go run dummy-http-server/main.go
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

In this case the first endpoint that was asked returned an internal server error (status 500). The second one returned OK, so that answer went to the client (in our test `curl`). The last two endpoints also sent replies, one with status 500 and the other with 200, but these were discarded.
