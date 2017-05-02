GoSpawn
=======

GoSpawn is a simplified hybrid between dumb-init_, a simplified
supervisord_ and syslog2stdout_, implemented in Go.

Because it's implemented in Go, you don't need the prerequisites for
*supervisord* which you would have to install otherwise. One statically
compiled binary is enough.

.. _dumb-init: https://github.com/Yelp/dumb-init
.. _supervisord: http://supervisord.org/
.. _syslog2stdout: https://github.com/ossobv/syslog2stdout


-----
Usage
-----

Syntax::

    gospawn [SYSLOGD_PORTS_AND_PATHS...] -- [COMMANDS...]

To spawn a syslog daemon on UDP port 514 and in /dev/log, and to spawn
*cron* and *uwsgi*, you use this:

.. code-block:: console

    $ gospawn 514 /dev/log -- /usr/sbin/cron -f -L 15 -- /usr/bin/uwsgi /uwsgi.ini

When processes succeed (return with code 0), they are not respawned. If
they fail, they are respawned:

.. code-block:: console

    $ gospawn -- /bin/sleep 1 -- /bin/false
    Spawned process 29087: [/bin/sleep 1]
    Spawned process 29089: [/bin/false]
    Reaped process 29089: [/bin/false], status 1
    Reaped process 29087: [/bin/sleep 1], status 0

    (after 10 seconds)

    Spawned process 29095: [/bin/false]
    Reaped process 29095: [/bin/false], status 1

If there is at least one process and all processes have completed
successfully, gospawn ends as well. If no commands are supplied, but
syslog ports are, it will run forever as a syslog daemon.


----
TODO
----

* Check what we do with stale sockets in syslog2stdout and do the same here.
* Also fix the socket ownership.
* Parse syslogd stuff instead of printing it verbatim:
  See also: https://github.com/ossobv/syslog2stdout/blob/master/syslog2stdout.c#L61
* Future: add cron-daemon?
* Should we have called this "minitgo" or "initgo" (mini, init in go?).
* See also: http://git.suckless.org/sinit/tree/sinit.c
* See also: https://github.com/Yelp/dumb-init/blob/master/dumb-init.c
