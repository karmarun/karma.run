<img src="https://karma.run/build/nav-logo.svg" height="100"/>

Copyright 2018 © karma.run AG. All rights reserved.

Use of this source code is governed by an AGPL license that can be found in the LICENSE file.

## Introduction

karma.run ("karma" for short) is an integrated back-end server solution to power data-driven
client applications. It is designed to be easy to use directly from a web browser, IoT or mobile environment.

## Features

karma.run comes with batteries included and features such as:
 * nested data object models.
 * strong serializable consistency.
 * powerful ABAC-based access control.
 * a relational model with traversable reference graph.
 * functional, homoiconic query interface.

## Run Karma (Docker)

**create a secret key:**  
The secret is the auth key of the servers "admin" user. In order to run karma you need to generate a secret:

**Mac:**
```bash
head -c 1024 /dev/urandom | base64
```
**Linux:**
If you are on Linux use the base64 -w0 flag to strip the newlines out
```bash
head -c 1024 /dev/urandom | base64 -w0
```

pull the image 
```bash
docker pull karmarun/karma.run
```

run karma
```bash
docker run -d \
           -e "KARMA_DATA_FILE=/karma-run-db.data" \
           -e "KARMA_INSTANCE_SECRET=base 64 encoded secret key" \
           -p 8080:80 \
           karmarun/karma.run
```

## Documentation

[docs.karma.run](https://docs.karma.run)

## Installation (Go)

To install karma.run you need the [Go toolchain](https://golang.org):

    $ go get karma.run

Stable tags will have binary releases.

Get started in 5 minutes

[![asciicast](https://asciinema.org/a/156598.png)](https://asciinema.org/a/156598)

## Dependencies & Licenses

karma.run depends on three external software packages:
 * [github.com/coreos/bbolt](https://github.com/coreos/bbolt) (MIT licensed)
 * [github.com/kr/pretty](https://github.com/kr/pretty) (MIT licensed)
 * [golang.org/x/crypto/bcrypt](https://github.com/golang/crypto) (BSD licensed)

 Please see their respective LICENSE files for legal information.
