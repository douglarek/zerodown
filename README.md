zerodown [![Build Status](https://secure.travis-ci.org/douglarek/zerodown.png)](https://travis-ci.org/douglarek/zerodown)
=====

Package zerodown provides a library that makes it easy to build socket
based servers that can be gracefully terminated & restarted (that is,
without dropping any connections).

Usage
-----

Demo HTTP Server with graceful termination and restart:
https://github.com/douglarek/zerodown/blob/master/zerodowndemo/server.go

1. Install the demo application

        go get github.com/douglarek/zerodowndemo

2. Start it in the first terminal

        zerodowndemo

3. In a second terminal start a slow HTTP request

        curl 'http://localhost:8080/?duration=20s'

4. In a third terminal trigger a graceful server restart:

        kill -USR2 [zerodowndemo pid]

5. Trigger another shorter request that finishes before the earlier request:

        curl 'http://localhost:8080/?duration=0s'


If done quickly enough, this shows the second quick request will be served by
the new process while the slow first request will be
served by the first server. It shows how the active connection was gracefully
served before the server was shutdown. It is also showing that at one point
both the new as well as the old server was running at the same time.
