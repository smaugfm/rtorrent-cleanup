# rtorrent-cleanup

Command line utility to delete rTorrent closed (that met ratio group criteria) torrents but with a time delay.

Some torrent trackers have a statistics update lag. It means that your download client may already
see that a torrent reached ratio of 1.0 but the tracker will update its statistics once in a while, i.e. every 30
minutes.
If you configure your rTorrent ratio group with an action of "Remove data" it will remove the torrent and its files
as soon as specified ratio is reached.
But your tracker will not catch up, and you may receive penalties for insufficient ratio (i.e. a hit&Run warning).

This tool is intended to be used with a ratio group action "Stop". It then queries rTorrent for torrents that
have a ratio group criteria met and are now stopped and deletes them *only if* a specified amount of
time has already passed since the torrent has been stopped.

Also, this tool is not a daemon, and you will have to run it periodically with (with a cron job, systemd timers, etc.).

Systemd [service](./systemd/rtorrent-cleanup.service) and [timer](./systemd/rtorrent-cleanup.timer) examples are included.

Installation

```shell
go install github.com/smaugfm/rtorrent-cleanup@latest
```

Usage:

```shell
rtorrent-cleanup http://localhost:8000/RPC2
```

With HTTP Basic Auth:

```shell
rtorrent-cleanup --username someUser --password somePassword http://localhost:8000/RPC2
```
