![](http://cl.ly/image/2v3O3s180l36/rig.png)

Run and manage services on your development machine.

## Overview

There are three terms you need to understand: tasks, services, and stacks.

A service is a directory that contains a Procfile. Typically, this will be a
component of a web application, for example, a web service written in Rails.
The Procfile defines the service's tasks.

Tasks are named processes, which will be run by rig. So a service may have a
task for running an application server, and a task for processing background
jobs. Here's an example of what the Procfile may look like:

```
web: bundle exec rails server -p $PORT
worker: bundle exec rake resque:work
```

A stack is a collection of services. A web application that consists of several
web services would be represented as a stack. A stack serves two purposes:
namespacing services within an application, and providing an easy way to
perform actions on a set of related services.

## Configuration

Rig is configured by placing directories and symlinks in the `$HOME/.rig/`
directory. Typically, there will be a directory for each stack in the `.rig`
directory. Within each of these stack directories, there will be symlinks that
point to the stack's services.

In the following example, we create a stack named "acme", with two services:
"api" and "website":

```shell-session
# Create a stack
[me@host ~]$ mkdir -p ~/.rig/acme

# Add some services to the stack
[me@host ~]$ ln -s ~/projects/acme-api ~/.rig/acme/api
[me@host ~]$ ln -s ~/projects/acme-website ~/.rig/acme/website
```

If you have a simpler application that only has one service (e.g. a basic Rails
app), you can create symlink in the `.rig` directory that points directly to
that service:

```shell-session
# Create a top-level service
[me@host ~]$ ln -s ~/projects/my-rails-blog ~/.rig/blog
```

## Usage

The typical usage for the Rig command line client is
`rig <command> <stack>:<service>:<task>`. However, most of the time, you won't
specify all three of the stack, service, and task. To refer to every service
and task within a stack, just provide the name of the stack. Similarly, you can
leave off the name of the task, and the command will be applied to all tasks
within a given service.


```shell-session
[me@host ~]$ rig list
== Stack acme ==
Service api
  => Task web
  => Task worker
Service website
  => Task web
  => Task compass

# Start a specific task
[me@host ~]$ rig start acme:api:web
[12:01:02 rig] starting task 'acme:api:web'

# Start all the tasks for a specific service
[me@host ~]$ rig start acme:api
[12:01:04 rig] starting task 'acme:api:web'
[12:01:04 rig] starting task 'acme:api:worker'

# Start all tasks in every service for a specific stack
[me@host ~]$ rig start acme
[12:01:06 rig] starting task 'acme:api:web'
[12:01:06 rig] starting task 'acme:api:worker'
[12:01:06 rig] starting task 'acme:website:web'
[12:01:06 rig] starting task 'acme:website:compass'
```

If your current working directory is a service directory, you can leave off both
the stack and the service when running commands:

```shell-session
[me@host ~]$ cd ~/projects/acme-api

# Start a specific task
[me@host acme-api]$ rig start web
[12:01:02 rig] starting task 'acme:api:web'

# Start all the tasks for the current service
[me@host acme-api]$ rig start
[12:01:04 rig] starting task 'acme:api:web'
[12:01:04 rig] starting task 'acme:api:worker'
```

While in a service directory, you can also refer to other services in the same
stack, by prepending a colon to the service names:

```shell-session
[me@host ~]$ cd ~/projects/acme-api

# Start a specific task
[me@host acme-api]$ rig start :website:web
[12:01:02 rig] starting task 'acme:website:web'

# Start all the tasks for a related service
[me@host acme-api]$ rig start :website
[12:01:04 rig] starting task 'acme:website:web'
[12:01:04 rig] starting task 'acme:website:compass'
```

Starting a top-level service (a service in the `~/.rig` directory, not inside
a stack directory) is the same as regular services, except you leave off the
stack name:

```shell-session
# Start a specific task
[me@host ~]$ rig start blog:web
[12:01:02 rig] starting task 'blog:web'

# Start all the tasks for a top-level service
[me@host ~]$ rig start blog
[12:01:04 rig] starting task 'blog:web'
[12:01:04 rig] starting task 'blog:worker'
```
