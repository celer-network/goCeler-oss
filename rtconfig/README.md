# Runtime Config
## High level design
When process receives SIGHUP signal, it'll reload runtime config file (path set as a flag), parse and update the runtime config (eg. https://golang.org/pkg/sync/atomic/#example_Value_config)

Current rtc includes
1. gas lower bound
2. gas upper bound
3. openchannel rate limit interval
4. stream send timeout


task breakdown:
1. rtc json key value and corresponding go struct (or use .proto?)
2. parse rtc.json once in server init
3. create a chan for os.Signal and if syscall.SIGHUP, reload again
4. places using the value needs to be changed to read from rtc
