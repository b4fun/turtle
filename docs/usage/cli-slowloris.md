# Is my server Slowloris-proofed?

[Slowloris attack][cf_slowloris] attempts to break an HTTP server by sending partial HTTP request,
which contains never finish HTTP header lines:

```
GET / HTTP 1.1  # this is the only line required to start an HTTP request
HOST example.com
User-Agent my-user-agent
Header-Name Header-Value
# ... keep sending gibberish header name & value lines
```

Since the HTTP request is never ended, vulnerable server keeps the connection open. As a result, server side resources like memory, file descriptor will be consumed.

Invulnerable server should close the connection after a specified time if the request is unable to read completely.

We can use turtle to validate if an HTTP endpoint is Slowloris-proofed:

- if the server is immune to the attack, we should see closed events, and the total number of requests should be more than the test connections;
- if the server is vulnerable to the attack, the connections will be kept opened until test finished.

## Validating via CLI

> **NOTE** We can start a test server with `turtle-proof`. For setup guide, please see [turtle proof server][turtle-proof-server].

1. Start a vulnerable server:

```
$ turtle-proof
2023/09/04 11:33:04 INFO server started turtle-proof.addr=127.0.0.1:8889
```

2. Launch the test with the `slowloris` sub-command:

```
$ turtle slowloris http://127.0.0.1:8889 --http-send-gibberish
```

We should see output similar to below, where the number of connections stays at 100 without closing.

![](/docs/demo/slowloris-vulnerable.gif)

3. Start a invulnerable server:

```
$ turtle-proof --scenario=proof
2023/09/04 11:36:49 INFO server started turtle-proof.addr=127.0.0.1:8889
```

4. Launch the test again

```
$ turtle slowloris http://127.0.0.1:8889 --http-send-gibberish
```

Now, we should see many closing / reopening events like this:

![](/docs/demo/slowloris-invulnerable.gif)

[cf_slowloris]: https://www.cloudflare.com/learning/ddos/ddos-attack-tools/slowloris/
[turtle-proof-server]: ./turtle-proof-server.md