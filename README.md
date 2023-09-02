<h3 align="center">
    <a href="https://github.com/b4fun/turtle">
        <img src="docs/assets/turtle-logo.png" width="220px" style="inline-block" />
    </a>
</h3>

🐢 Turtle is a toolkit for simulating and validating application layer denial-of-service attacks in both live and unit testing environments.

<p align="center">
    <a href="https://github.com/b4fun/turtle/releases"><img src="https://img.shields.io/github/release/b4fun/turtle.svg" alt="Github release"></a>
    <a href="https://pkg.go.dev/github.com/b4fun/turtle"><img src="https://pkg.go.dev/badge/github.com/b4fun/turtle.svg" alt="GoDoc" /></a>
</p>

## 🚨 Disclaimer

> **Important**: The use of this program for attacking targets without prior mutual consent is illegal. It is the end user's responsibility to comply with all applicable local, state, and federal laws. The developers assume no liability and are not responsible for any misuse or damage caused by this program.

## 🎯 Why Use Turtle?

Exposing an application to the public internet is fraught with risks due to various types of denial-of-service attacks, such as:

- [slowloris][cf_slowloris]
- [low and slow attack][cf_low_and_slow]
- [R.U.D.Y][cf_rudy]
- ... and many more

While some applications may have well-configured settings that render them invulnerable to these attacks, others, such as those built with popular languages like Golang, might be [vulnerable by default][gonuts_slowloris].
Turtle provides an easy way to validate your application against these common threats to identify risks.

Furthermore, an application that is secure today may become vulnerable due to future changes.
Therefore, integrating these attack simulations into your regular validation process is crucial.

## 🛠 Features

Turtle provides:

- A Command-Line Interface (CLI) for validating real endpoints
- A Golang library for easy integration into unit/integration tests

### Supported Scenarios

Turtle current supports the following scenarios:

- [slowloris][cf_slowloris]
- [slow body read][cf_low_and_slow]

## 🚀 Getting Started

### Turtle CLI

You can install the CLI tool via:

```bash
go install github.com/b4fun/turtle/cmd/turtle@latest
```

Or download a release binary from the [GitHub Release page][gh_release].

### Using Turtle CLI

The turtle CLI embeds supported scenarios as sub-commands. A common way to invoke a scenario test:

```
$ turtle <scenario-name> <target-url>
```

![](/docs/demo/demo.gif)

Further details can be obtained by viewing the command's help message:


```
$ turtle -h
# Scenario specified help
$ turtle slowloris -h
```

### Turtle Golang Library

For the Golang library, documentation can be found on [GoDoc][godoc].

## 📜 LICENSE

Turtle is distributed under the [MIT license][/LICENSE]

[cf_slowloris]: https://www.cloudflare.com/learning/ddos/ddos-attack-tools/slowloris/
[cf_low_and_slow]: https://www.cloudflare.com/learning/ddos/ddos-low-and-slow-attack/
[cf_rudy]: https://www.cloudflare.com/learning/ddos/ddos-attack-tools/r-u-dead-yet-rudy/
[gonuts_slowloris]: https://groups.google.com/g/golang-nuts/c/MFZd6b8zQTQ
[gh_release]: https://github.com/b4fun/turtle/releases
[godoc]: http://godoc.org/github.com/b4fun/turtle