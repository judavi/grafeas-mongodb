# Grafeas - MongoDb

[![Build Status](https://github.com/judavi/grafeas-mongodb/workflows/GitHub%20Actions/badge.svg)](https://github.com/judavi/grafeas-modngodb/actions)

This project provides a [Grafeas](https://github.com/grafeas/grafeas) implementation that supports using MongoDb as a storage mechanism.

## Building

Build using the provided Makefile or via Docker.

```shell
# Either build via make
make build

# or docker
docker build --rm .
```

## Unit tests

Testing is performed against a MongoDb instance.  The Makefile offers the ability to start and stop a locally installed MongoDb instance running via Java.  This requires that port 8000 be free.


## Contributing

Pull requests welcome.

## License

Grafeas-mongodb is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
