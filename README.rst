GoSpawn
=======

.. image:: https://bettercodehub.com/edge/badge/ossobv/gospawn?branch=master
    :target: https://bettercodehub.com/

.. image:: https://goreportcard.com/badge/github.com/ossobv/gospawn
    :target: https://goreportcard.com/report/github.com/ossobv/gospawn

GoSpawn is a cross between dumb-init_, a simplified supervisord_ and
syslog2stdout_, implemented in Go.

Because it's implemented in Go, you don't need the prerequisites for
*supervisord* which you would have to install otherwise. One statically
compiled binary is enough.

.. _dumb-init: https://github.com/Yelp/dumb-init
.. _supervisord: http://supervisord.org/
.. _syslog2stdout: https://github.com/ossobv/syslog2stdout


Usage
-----

Syntax::

    gospawn [SYSLOGD_PORTS_AND_PATHS...] -- [COMMANDS...]

To spawn a syslog daemon on UDP port 514 *and* in ``/dev/log``, and to
spawn *cron* and *uwsgi*, you use this::

    gospawn 514 /dev/log -- /usr/sbin/cron -f -L 15 -- /usr/bin/uwsgi /uwsgi.ini

When processes succeed (return with code 0), they are not respawned. If
they fail, they are respawned:

.. code-block:: console

    $ gospawn -- sleep 1 -- false
    Spawned process 29087 [sleep 1], running
    Spawned process 29089 [false], running
    Reaped process 29089 [false], status 1
    Reaped process 29087 [sleep 1], status 0

    (after 10 seconds)

    Spawned process 29095 [false], running
    Reaped process 29095 [false], status 1

If there is at least one process and all processes have completed
successfully, gospawn ends as well. If no commands are supplied, but
syslog ports are, it will run forever as a syslog daemon.

Other examples:

* The environment is preserved, and so are stdin/stdout.

    .. code-block:: console

        $ docker run -e VALUE=val -it IMG /gospawn -- sh -c 'echo $VALUE'

        Spawned process 13: [sh -c echo $VALUE]
        val
        Reaped process 13: [sh -c echo $VALUE], status 0

* Zombies are reaped immediately; observe how the ``sleep 5`` background
  job is reaped in a timely fashion (see the process list) and that
  syslog messages are captured.

    .. code-block:: console

        $ docker run -it IMG /gospawn /dev/log -- python -c '\
                import subprocess,time,syslog;\
                p=subprocess.Popen("setsid sleep 5 &", shell=True);p.wait();\
                syslog.syslog("subprocess done");time.sleep(10);\
                syslog.syslog("sleep done")'
        Spawned syslogd at UNIX(/dev/log)
        Spawned process 12 [python -c ...], running
        <14>May  4 09:11:22 -c: subprocess done
        Reaped process 14, status 0
        <14>May  4 09:11:32 -c: sleep done
        Reaped process 12 [python -c ...], status 0


TODO
----

* Check what we do with stale sockets in syslog2stdout and do the same here.
* Bugs: the Close() during shutdown may send syslogd Read errors to the console.
* Also fix the socket ownership.
* Parse syslogd stuff instead of printing it verbatim:
  See also: https://github.com/ossobv/syslog2stdout/blob/master/syslog2stdout.c#L61
* Future: add cron-daemon?
* Should we have called this "minitgo" or "initgo" (mini, init in go?).
* See also: http://git.suckless.org/sinit/tree/sinit.c
* See also: https://github.com/Yelp/dumb-init/blob/master/dumb-init.c
