GoSpawn
=======

This is intended to be a cross between dumb-init, a simplified
supervisord and syslog2stdout in Go.


----
TODO
----

* Check what we do with stale sockets in syslog2stdout and do the same here.
* Auto-respawn processes which do not end with return code 0? Sounds supervisorday-y.
* Future: add cron-daemon?
* Should we have called this "minitgo" or "initgo" (mini, init in go?).
