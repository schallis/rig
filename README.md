![](http://cl.ly/image/3a0k2U2y2B42/ignition-header.png)

## Usage

```bash
$ go build .
$ ./ignition conf.json
# a bunch of cool stuff happens
```

## Configuration

Ignition accepts a JSON configuration file, which defines how to run your
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

