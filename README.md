## WARNING: This project is in development and not ready for use.

<img src="http://f.cl.ly/items/0A362Z0H2C272k1O1832/rig-logo.jpg" width="200"/>

Run and manage services on your development machine.

## Installation

Assuming you're on a Mac:

```shell-session
[me@host ~]$ brew tap gocardless/rig
[me@host ~]$ brew install rig
```

## Overview

There are three terms you need to understand: processes, services, and stacks.

A service is a directory that contains a Procfile. Typically, this will be a
component of a web application, for example, a web service written in Rails.
The Procfile defines the service's processes.

Processes are named commands, which will be run by rig. So a service may have a
process for running an application server, and a process for processing
background jobs. Here's an example of what the Procfile may look like:

```
web: bundle exec rails server -p $PORT
worker: bundle exec rake resque:work
```

A stack is a collection of services. A web application that consists of several
web services would be represented as a stack. A stack serves two purposes:
namespacing services within an application, and providing an easy way to
perform actions on a set of related services.

## Configuration

Rig is stored in a JSON file by default at `$HOME/.config/rig/config.json`
directory. Rig will create one default for you when it launches for the first time.

Here is an example configuration :

```json
{
  "stacks": {
    "default": {},
    "acme": {
      "acme-api": {
        "dir": "/Users/steve/src/acme-api"
      },
      "acme-website": {
        "dir": "/Users/steve/src/acme-website"
      }
    }
  }
}
```

If you have a simpler application that only has one service (e.g. a basic Rails
app), you can add it to the 'default' stack:

```json
{
  "stacks": {
    "default": {
      "blog": {
        "dir": "/Users/steve/src/my-rails-blog"
      }
    },
    "acme": {
      "acme-api": {
        "dir": "/Users/steve/src/acme-api"
      },
      "acme-website": {
        "dir": "/Users/steve/src/acme-website"
      }
    }
  }
}
```

## Usage

The typical usage for the Rig command line client is
`rig <command> <stack>:<service>:<process>`. However, most of the time, you
won't specify all three of the stack, service, and process. To refer to every
service and process within a stack, just provide the name of the stack.
Similarly, you can leave off the name of the process, and the command will be
applied to all processes within a given service.


```shell-session
[me@host ~]$ rig list
== Stack acme ==
Service api
  => Process web
  => Process worker
Service website
  => Process web
  => Process compass

# Start a specific process
[me@host ~]$ rig start acme:api:web
[12:01:02 rig] starting process 'acme:api:web'

# Start all the processes for a specific service
[me@host ~]$ rig start acme:api
[12:01:04 rig] starting process 'acme:api:web'
[12:01:04 rig] starting process 'acme:api:worker'

# Start all processes in every service for a specific stack
[me@host ~]$ rig start acme
[12:01:06 rig] starting process 'acme:api:web'
[12:01:06 rig] starting process 'acme:api:worker'
[12:01:06 rig] starting process 'acme:website:web'
[12:01:06 rig] starting process 'acme:website:compass'
```

If your current working directory is a service directory, you can leave off
both the stack and the service when running commands:

```shell-session
[me@host ~]$ cd ~/projects/acme-api

# Start a specific process
[me@host acme-api]$ rig start web
[12:01:02 rig] starting process 'acme:api:web'

# Start all the processes for the current service
[me@host acme-api]$ rig start
[12:01:04 rig] starting process 'acme:api:web'
[12:01:04 rig] starting process 'acme:api:worker'
```

While in a service directory, you can also refer to other services in the same
stack, by prepending a colon to the service names:

```shell-session
[me@host ~]$ cd ~/projects/acme-api

# Start a specific process
[me@host acme-api]$ rig start :website:web
[12:01:02 rig] starting process 'acme:website:web'

# Start all the processes for a related service
[me@host acme-api]$ rig start :website
[12:01:04 rig] starting process 'acme:website:web'
[12:01:04 rig] starting process 'acme:website:compass'
```

Starting a top-level service (a service in the `~/.rig` directory, not inside
a stack directory) is the same as regular services, except you leave off the
stack name:

```shell-session
# Start a specific process
[me@host ~]$ rig start blog:web
[12:01:02 rig] starting process 'blog:web'

# Start all the processes for a top-level service
[me@host ~]$ rig start blog
[12:01:04 rig] starting process 'blog:web'
[12:01:04 rig] starting process 'blog:worker'
```
