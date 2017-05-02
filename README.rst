GoSpawn
=======

This is intended to be a cross between dumb-init, a simplified
supervisord and syslog2stdout in Go.


-----
Usage
-----

::

    gospawn 514 /dev/log -- /usb/sbin/cron -f -L 15 -- /usr/bin/uwsgi /uwsgi.ini


----
TODO
----

* Check what we do with stale sockets in syslog2stdout and do the same here.
* Parse syslogd stuff instead of printing it verbatim:
  See also: https://github.com/ossobv/syslog2stdout/blob/master/syslog2stdout.c#L61
* Auto-respawn processes which do not end with return code 0? Sounds supervisorday-y.
* Future: add cron-daemon?
* Should we have called this "minitgo" or "initgo" (mini, init in go?).
* See also: http://git.suckless.org/sinit/tree/sinit.c
* See also: https://github.com/Yelp/dumb-init/blob/master/dumb-init.c
