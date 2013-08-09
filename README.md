# rig

Runs services.

## Usage

```bash
$ go build .
$ ./rigd conf.json
# a bunch of cool stuff happens
```

## Configuration

Rig accepts a JSON configuration file, which defines how to run your
services:

```json
{
  "services": {
    "ping-google": {
      "command": "ping",
      "args": ["-c", "3", "google.com"]
    },
    "ping-bing": {
      "command": "ping",
      "args": ["-c", "3", "www.bing.com"]
    }
  }
}
```

