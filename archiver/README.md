# Demo Blueprint

References: [Golang Blueprint](https://github.org/jeks313/service-blueprint-golang).

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
