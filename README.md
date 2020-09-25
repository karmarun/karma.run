<img src="https://karma.run/build/nav-logo.svg" height="100"/>

Copyright 2018 © karma.run AG. All rights reserved.

Use of this source code is governed by an AGPL license that can be found in the LICENSE file.

## Introduction

karma.run ("karma" for short) is an integrated back-end server solution to power data-driven
client applications. It is designed to be easy to use directly from a web browser, IoT or mobile environment.

## Features

karma.run comes with batteries included and features such as:

- nested data object models.
- strong serializable consistency.
- powerful ABAC-based access control.
- a relational model with traversable reference graph.
- functional, homoiconic query interface.

## Roadmap

There are many features and tasks yet to be implemented. Some of them are:

- Write lots of unit tests and measure code coverage.
- Releasing a karma.run-compatible web-based GUI ("the web editor") as FOSS.
- Writing and releasing official client libraries for different programming languages.

## Status

The project is currently in alpha stage. Concretely, any and all APIs are subject to change.
In regards to documentation, there is some but it's mostly old and unusable, we are working
on an updated and accurate version.
Tests have been conducted by building real-life applications on top of karma.run but not
formalized as unit-tests and no coverage numbers are available yet.

## Installation

To install karma.run you need the [Go toolchain](https://golang.org):

    $ go get karma.run

Stable tags will have binary releases.

Get started in 5 minutes

[![asciicast](https://asciinema.org/a/156598.png)](https://asciinema.org/a/156598)

## Dependencies & Licenses

karma.run depends on three external software packages:

- [github.com/boltdb/bolt](https://github.com/boltdb/bolt) (MIT licensed)
- [github.com/kr/pretty](https://github.com/kr/pretty) (MIT licensed)
- [golang.org/x/crypto/bcrypt](https://github.com/golang/crypto) (BSD licensed)

Please see their respective LICENSE files for legal information.

## Karma Migration Error 0.4.0

If during a migration the error "unexpext end of JSON file" occurs and the offset is 1048576 (1MB). Use version 0.4.1 on the Docker Hub. In the newer version the MaxPayloadBytes has been increased to 10MB
