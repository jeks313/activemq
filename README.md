# ActiveMQ Archiver And Other Utilities

## Archiver

Archiver listens to an ActiveMQ topic via STOMP specified by using just the topic name. This creates a queue automatically
using ActiveMQ convention of Consumer.Archive.VirtualTopic.your-topic which creates a parallel queue.

Main features:

- Files are written by hour and by the specified key in the JSON payload given. This assumes the payload is JSON, of course.
- Headers specified that on the ActiveMQ message are merged with the JSON output.
- Certain payload encoding is undone (such as base64'd zips) and the JSON is compressed to a single line in the resulting archive.
- Files written are rolled over when a certain filesize is given so that Athena can use them out of S3 as it cannot process files larger than 32MB.
- Path to write files and output path is specified for when the file is complete. This output path can then be watched and the files uploaded to S3
  by a separate process. 

Important environment variables/options to set:

`$TOPIC` - the activemq topic to archive, don't specify the VirtualTopic part ... if you didn't follow the right convention shame on you.
`$ACTIVE_MQ` - the activemq hostname to use.
`$KEY` - the key in the JSON to partition the filename on.
`ACTIVEMQ_HEADERS` - the headers you want merged to the JSON under the 'headers' key.

```
Usage:
  activemq-archiver [OPTIONS]

Application Options:
      --port=         application port (default: 8080) [$PORT]

ActiveMQ Options:
      --topic=        topic to archive [$TOPIC]
      --activemq=     activemq hostname (default: localhost:61613) [$ACTIVE_MQ]
      --archive-path= base directory to write archive files (default: /var/lib/activemq-archive) [$ARCHIVE_PATH]
      --key=          key to look for in the document to use to construct archive filename [$TOPIC_KEY]
      --max-size=     maximum archive size, defaults to 32M for athena usage in S3 (default: 33554432) [$MAX_ARCHIVE_SIZE]
      --header=       headers to include from activemq into the payload (default: accountUid, deviceIdentity, deviceUid, esn, x-Content-Type) [$ACTIVEMQ_HEADERS]

Default Service Options:
      --limit=        maximum permitted http connections (default: 1000) [$LIMIT]
      --ssl           enable SSL, default key and crt will be binary name .crt and .key [$ENABLE_SSL]

Default Application Server Options:
  -d, --debug         enable debug logging level [$DEBUG]
  -e, --env=          environment this is running in (default: dev) [$ENVIRONMENT]
  -v, --version       output version variables

Help Options:
  -h, --help          Show this help message
```

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
$ make build
```
or
```
$ docker build
```
