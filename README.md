# gohippo

![CI](https://github.com/kentik/gohippo/workflows/CI/badge.svg)
[![GitHub Release](https://img.shields.io/github/release/kentik/gohippo.svg?style=flat)](https://github.com/kentik/gohippo/releases/latest)
[![Coverage Status](https://coveralls.io/repos/github/kentik/gohippo/badge.svg?branch=main)](https://coveralls.io/github/kentik/gohippo?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/kentik/gohippo)](https://goreportcard.com/report/github.com/kentik/gohippo) 

## Build process

To build hippo you should use [Eathly](https://earthly.dev/)
```
earthly +all
```

That will regenerate protobuf files and run build + test process.
