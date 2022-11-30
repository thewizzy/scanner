# Scanner

Provides a consistent API around some existing scanning tools to integrate them
with the rest of the tool kit.

This scanner is meant to be ran as a single instance listening on stdin
for easy scripting. There will be a server to wrap the scanner to be able to
turn it into a service

The scanner leverages
[gitleaks](https://github.com/zricethezav/gitleaks)
internally because gitleaks is an awesome tool, and we already have quite a few
[patterns built up](https://github.com/leaktk/patterns)
for it.

## Status

Just getting started.

## Usage (Pending Implementation)

To start the scanner and have it listen for requests run:

```sh
leaktk-scanner --config /path/to/config.toml
```

The scanner listens on stdin, responds on stdout, and logs to stderr or a file
if it is defined in the config.

The scanner reads one request per line and sends a response per line in jsonl.

The "Scan Request Format" and "Scan Results Format" sections describe the
format for requests and responses.

## Config File Format (WIP)

```toml
[logger]
level = "INFO"
filepath = "/var/log/leaktk/leaktk.log"

[scanner]
# This is the directory where the scanner will store files
workdir = "/tmp/leaktk"

[scanner.patterns]
# TODO: define auth settings for servers that require auth
server_url = "https://raw.githubusercontent.com/leaktk/patterns/main/target"
refresh_interval = 43200
```

## Scan Request Format

The scan request format is in jsonl here are formatted examples of a single
line by type

### Git

(TODO: support `file://` for local scans that skip the clone)

```json
{
  "id": "<uuid>",
  "type": "git",
  "url": "https://github.com/leaktk/fake-leaks.git",
  "options": {
    "clone_depth": 5,
  }
}
```

## Scan Results Format

The scan result format is in jsonl here are formatted examples of a single
line by type

### Git

```json
{
  "id": "<uuid>",
  "request": {
    "id": "<uuid>"
  },
  "results": [
    {
      "context": "<the match>",
      "target": "<the part of the match the rule was trying to find>",
      "entropy": 3.2002888,
      "rule": {
        "id": "some-rule-id",
        "description": "A human readable description of the rule",
        "tags": [
          "alert:repo-owner",
          "type:secret",
        ]
      },
      "source": {
        "type": "git",
        "url": "https://github.com/leaktk/fake-leaks.git",
        "path": "relative/path/to/the/file",
        "lines": {
          "start": 1,
          "end": 1
        },
        "commit": {
          "id": "<commit-sha>",
          "author": {
            "name": "John Smith",
            "email": "jsmith@example.com"
          },
          "date": "2022-08-29T15:32:48Z",
          "message": "<commit message>"
        }
      }
    }
  ]
}
```

## TODO

* Fix/cleanup the error handling
* Proper logging
* Unittest and refactor what's currently there
* Group gitleaks code into a single object as the source of truth
* Create a Workspace object to manage the workspace folders (creating, clearing, etc)
* Encapsulate some of the linux specific bits
