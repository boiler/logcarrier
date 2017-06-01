# logcarrier
Logfile tailing/delivery system

Config format:

```toml
listen=":1146"
listen_debug=":40000"              # HTTP server will be started, useful to check availability
wait_timeout="1m13s"
LogFile="/var/log/logcarrier/log"  # stderr will be used if this parameter is not set

[compression]
method="zstd"                      # Can be `zstd` or `raw`
level=6

[buffers]
input="128Kb"                      # Kb, Mb, Gb can be used here (or nothing). Input buffer right from the connection.
framing="256Kb"                    # Same format. Framing buffer after compression to ensure compressed frame integrity.
zstdict="128Kb"                    # Same format. ZSTD compression dictionary size.
connections=1024                   # Depth of queue for connections
dumps=512                          # Depth of dumping tasks' queue
logrotates=512                     # Depth of logrotating tasks' queue

[workers]
route=1024                         # Amount of workers to process and route incoming connection
dumper=24                          # Amount of workers reading data from tailers
logrotater=12                      # Amount of workers what logrotate on LOGROTATE request

flusher_sleep="30s"                # Intervals for force flusher sleep

[files]
root="/var/logs/logcarrier"                     # Root directory to put logs in
root_mode=  0755                                  # Mode for directories what are creating in process
rotation="/${dir}/${name}-${ time | %Y%m%d%H }" # Masks for file name current file moves into after rotation.
                                                # Available vars:
                                                #   time:  time
                                                #   dir:   string
                                                #   name:  string
                                                #   group: string
      
[links]
root="/var/logs/logcarrier"                                    # Same as for files
name="/${dir}/${ time | %Y/%m/%d }/${name}-${ time | %H}"      # Name for the link that follows current file
rotation="/${dir}/${ time | %Y/%m/%d }/${name}-${ time | %H}"  # Same as for files

[logrotate]
method="periodic"                  # can be periodic, guided (via protocol) and both
schedule="* */1 * * *"             # start log rotation tasks at the start of each hour
```
