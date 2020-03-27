# ActiveMQ Archiver and other Utilities

## Archiver

Fields wanted:

- `Ctes-Platform`
- `esn`
- `deviceIdentity`
- `deviceUid`
- `accountUid`

Formats:

- `X-Content-Type`
 - `zip;json`
 - `b64;zip;json`

# API

## System:

* [Health](/health)
* [Metrics](/metrics)
* [Version](/version)

## Logging and Debugging:

* [Text Log](/log)
* [JSON Log](/log?format=json)
* [Debugging](/debug/pprof/)

# Options

Invoke help for options:

```
{{options}}
```

# Deployment

Build for production:

```
$ mage build
```

References: [Golang Blueprint](https://github.org/jeks313/service-blueprint-golang).
