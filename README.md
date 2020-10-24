[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/skwair/harmony)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat-square)](LICENSE)
[![Discord](https://img.shields.io/badge/Discord-online-7289DA.svg?style=flat-square)](https://discord.gg/3sVFWQC)
[![Build Status](https://travis-ci.org/skwair/harmony.svg?branch=master)](https://travis-ci.org/skwair/harmony)


# Harmony

<img align="right" height="200" src=".github/discord-gopher.png">

Harmony is a peaceful [Go](https://go.dev) module for interacting with [Discord](https://discord.com)'s API.

Although this package is usable, it still is under active development so please don't use it for anything other than experiments, yet.

**Contents**

- [Installation](#installation)
- [Usage](#usage)
- [Testing](#testing)
- [How does it compare to DiscordGo?](#how-does-it-compare-to-discordgo-)
- [License](#license)

# Installation

Make sure you have a working Go installation, if not see [this page](https://golang.org/dl) first.

Then, install this package with the `go get` command:

```sh
go get github.com/skwair/harmony
```

> Note that `go get` will always pull the latest version from the master branch before Go 1.11. With newer versions and Go modules enabled, the latest minor or patch release will be downloaded. `go get github.com/skwair/harmony@major.minor.patch` can be used to download a specific version. See [Go modules](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies) for more information.

# Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/skwair/harmony"
)

func main() {
    client, err := harmony.NewClient("your.bot.token")
    if err != nil {
        log.Fatal(err)
    }

    // Get information about the current user (the bot itself).
    u, err := client.User("@me").Get(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(u)
}
```

For information about how to create bots and more examples on how to use this package, check out the [examples](https://github.com/skwair/harmony/blob/master/examples) directory and the [tests](https://github.com/skwair/harmony/blob/master/harmony_test.go).

# Testing

For now, only some end to end tests are provided with this module. To run them, you will need a valid bot token and a valid Discord server ID. The bot attached to the token must be in the server with administrator permissions.

1. Create a Discord test server

From a Discord client and with you main account, simply create a new server. Then, right click on the new server and get its ID.

> Note that for the UI to have the `Copy ID` option when right clicking on the server, you will need to enable developer mode. You can find this option in `User settings > Appearance > Advanced > Developer Mode`.

2. Create a bot and add it to the test Discord server

Create a bot (or use an existing one) and add it to the freshly created server.

> See the [example directory](https://github.com/skwair/harmony/blob/master/examples) for information on how to create a bot and add it to a server.

3. Set required environment variables and run the tests

Set `HARMONY_TEST_BOT_TOKEN` to the token of your bot and `HARMONY_TEST_GUILD_ID` to the ID of the server you created and simply run:

⚠️ **For the tests to be reproducible, they will start by deleting ALL channels in the provided server. Please make sure to provide a server created ONLY for those tests.** ⚠️

```bash
go test -v -race ./...
```

> Step 1 and 2 must be done only once for initial setup. Once you have your bot token and the ID of your test server, you can run the tests as many times as you want.

# How does it compare to [DiscordGo](https://github.com/bwmarrin/discordgo)?

Harmony exposes its API differently. It uses a [resource-based](https://godoc.org/github.com/skwair/harmony#hdr-Using_the_HTTP_API) approach which organizes methods by topic, greatly reducing the number of methods on the main `Client` type. The goal by doing this is to have a more friendly API which is easier to navigate.

Another key difference is in the "event handler" mechanism. Instead of having a single [method](https://github.com/bwmarrin/discordgo/blob/7ab242d361c0dd43613f8c188e4978b4d18a8c89/event.go#L120) that takes an `interface{}` as a parameter and guesses which event you registered a handler for based on its concrete type, this library provides a dedicated method for each event type, making it clear what signature your handler must have and ensuring it at compile time, not at runtime.

Each action that results in an entry in the audit log has a `...WithReason` form, allowing to set a reason for the change (see the `X-Audit-Log-Reason` [header](https://discord.com/developers/docs/resources/audit-log#audit-logs) documentation for more information).

Finally, this library has a full support of the [context](https://golang.org/pkg/context/) package, allowing the use of timeouts, deadlines and cancellation when interacting with Discord's API.

# License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/skwair/harmony/blob/master/LICENSE) file for details.

Original logo by [Renee French](https://instagram.com/reneefrench), dressed with the cool t-shirt by [@HlneChd](https://twitter.com/hlnechd).
