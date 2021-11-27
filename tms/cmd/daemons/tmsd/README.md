# Introduction
TMSD Readme

## Summary
 
A daemon to start and stop tms processes according to a configuration file. It 
also works to monitor the running tms processes. It is configured as an upstart
job.

## Configuration file

An sample tmsd.conf file is shown below:

```
control=/var/run/trident/tmsd.sock
pid=/var/run/trident/tmsd.pid
users=tmac,tsiadmin

{global -log=INFO}
{tgwad --num 10 --name batam}
{tnsd.py /var/lib/trident/sensorData/css-btm-24hour 3452}
{tnoid --address 127.0.0.1:3452}
{tclientd }
{tdatabased -mongo-data /var/trident/db }
{twebd }
```

Besides setting flags of tmsd within the command to run tmsd, they can also be 
set in the configuration file in the format of "variable=value" (as control, 
pid, users in the example). This is necessary when running tmsd using upstart.

The commands to run sub-processes are included in braces ({}). One exception is 
if the command within a pair of braces starts with "global", the following flag 
settings will be applied to all sub-processes. In the example above, all 
sub-processes will be run with -log=INFO.

## Start tmsd

### Start tmsd manually

```bash
tmsd -start
```

### Start tmsd with upstart (recommended)

```bash
sudo service tmsd start
```
or
```bash
sudo start tmsd
```

## Monitor tmsd

If there is a running tmsd, it can be monitored by "tmsd -status" or 
"tmsd -info". They prints status or information of the running sub-processes.

### tmsd status
```bash
tmsd -status
```
tmsd status lists all the running sub-processes and their status (not launch, 
running, crashed, stopped, unknown)


### tmsd info
```bash
tmsd -info
```

tmsd info lists more information of the running sub-processes than tmsd status.
It prints these fields of each sub-process.
* pid
* name
* status
* number of times the tmsd tried to start it
* the last time that tmsd started it successfully
* command line arguments of this sub-process


## Stop tmsd

### Stop tmsd manually

```bash
tmsd -stop
```

### Stop tmsd with upstart (recommended)

```bash
sudo service tmsd stop
```
or
```bash
sudo stop tmsd
```

If the running tmsd is invoked by upstart, it has to be stopped by upstart. 
Manually stopping it will not work. Upstart will start it again automatically.

tmsd -stop is supposed to send a second SIGINT to sub-processes which are still
running (after 5 seconds in default). So they will exit gracefully. If some 
sub-processes are still running after that, tmsd will send a SIGKILL in a 
longer time (60 seconds in default). If you want to stop all sub-processes in 
a shorter time, you can add a flag -kill. Then tmsd will not send a second 
SIGINT, but send a SIGKILL directly instead to shut down running sub-processes
immediately.

To run tmsd -stop with a flag -kill using upstart:

```bash
sudo service tmsd stop shutdown=KILL
```
or
```bash
sudo stop tmsd shutdown=KILL
```

## Cleanu up running tmsd sub-processes and temporary files

If a running tmsd shut down accidently, its sub-processes will probably keep 
running and are out of control as a result. tmsd -cleanup will shut down all 
running sub-processes invoked by the last tmsd according to the tmsd.pid file.
Then tmsd.pid file is refreshed and tmsd.sock will be deleted. In fact, 
tmsd -start will do tmsd -cleanup firstly and then start to invoke all 
sub-processes.

```bash
tmsd -cleanup
```



