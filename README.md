
# Khronos - Fault tolerant job scheduling system

```

      ___           ___           ___           ___           ___           ___           ___     
     /  /\         /  /\         /  /\         /  /\         /  /\         /  /\         /  /\    
    /  /:/        /  /:/        /  /::\       /  /::\       /  /::|       /  /::\       /  /::\   
   /  /:/        /  /:/        /  /:/\:\     /  /:/\:\     /  /:|:|      /  /:/\:\     /__/:/\:\  
  /  /::\____   /  /::\ ___   /  /::\ \:\   /  /:/  \:\   /  /:/|:|__   /  /:/  \:\   _\_ \:\ \:\ 
 /__/:/\:::::\ /__/:/\:\  /\ /__/:/\:\_\:\ /__/:/ \__\:\ /__/:/ |:| /\ /__/:/ \__\:\ /__/\ \:\ \:\
 \__\/~|:|~~~~ \__\/  \:\/:/ \__\/~|::\/:/ \  \:\ /  /:/ \__\/  |:|/:/ \  \:\ /  /:/ \  \:\ \:\_\/
    |  |:|          \__\::/     |  |:|::/   \  \:\  /:/      |  |:/:/   \  \:\  /:/   \  \:\_\:\  
    |  |:|          /  /:/      |  |:|\/     \  \:\/:/       |__|::/     \  \:\/:/     \  \:\/:/  
    |__|:|         /__/:/       |__|:|~       \  \::/        /__/:/       \  \::/       \  \::/   
     \__\|         \__\/         \__\|         \__\/         \__\/         \__\/         \__\/    


```


### CRON

Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | Yes        | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

### Predefined schedules

Entry                  | Description                                | Equivalent To
-----                  | -----------                                | -------------
@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
@monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 0 * * 0
@daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
@hourly                | Run once an hour, beginning of hour        | 0 0 * * * *


### Intervals

You may also schedule a job to execute at fixed intervals. This is supported by formatting the cron spec like this:
```
@every <duration> 
```
For example, “@every 5s” would indicate a schedule that activates 5 seconds.

### Concurrency
allow (default): Allow concurrent job executions.
forbid: If the job is already running don’t send the execution, it will skip the executions until the next schedule.

### Fault tolerance
Fault detection, Failover, Failtry.

### Load banlancing
support Random, Workload RoundRobin


### Requirements
Khronos relies on the key-value data storage, Currently only etcd is supported


### Getting Started

Supposed you have installed Khronos onto your machine, you can follow the below steps to start using it.

To Run khronos:
```bash
$ khronos agent -e local
```

### TODO
- [x] Support Distributed
- [x] REST API
