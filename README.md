# limits-probe
Determine the resource usage limits of a process by trying them out!

## Maxing out file descriptors by opening files and making TCP connections
### See ulimit in live action
    > ulimit -n 123 -S
    > go run github.com/HouzuoGuo/limits-probe/cmd/limits-probe -ex file -ex sock
    2020/11/01 14:54:29 Max number of open files: soft limit 123, hard limit 1048576, kernel limit 1048576
    2020/11/01 14:54:29 Successfully opened 116 FDs and then encountered failure: open /tmp/github-HouzuoGuo-limits-probe155574448: too many open files
    2020/11/01 14:54:29 Sleeping 10 seconds in between experiments
    2020/11/01 14:54:39 server is quitting: accept tcp 127.0.0.1:19441: accept4: too many open files
    2020/11/01 14:54:39 Successfully made 58 TCP connections and then encountered failure: dial tcp 127.0.0.1:19441: socket: too many open files

### Setting ulimit via systemd
    > systemd-run --user --pty -p LimitNOFILE=123 --working-directory="$(readlink -f .)" go run github.com/HouzuoGuo/limits-probe/cmd/limits-probe -ex file -ex sock
    (Same output as above)

## Maxing out memory usage by allocating and using them
### See cgroup memory limit in live action
    > sudo cgcreate -a howard -t howard -g memory:/limits-probe-very-own
    > cgset -r memory.limit_in_bytes=536870912 limits-probe-very-own
    > cgset -r memory.swappiness=0 limits-probe-very-own
    > cgget -g memory /limits-probe-very-own
    > cgexec -g memory:/limits-probe-very-own go run github.com/HouzuoGuo/limits-probe/cmd/limits-probe -ex mem
    ...
    2020/11/01 15:07:37 Allocated 400 MB of memory
    signal: killed

    > sudo journalctl -r
    ...
    Nov 01 15:07:37 ip-10-0-78-238 kernel: Memory cgroup out of memory: Killed process 313675 (limits-probe) total-vm:1544140kB, anon-rss:511316kB, file-rss:936kB, shmem-rss:0kB, UID:1001 pgtables:1132kB oom_score_adj:0

    > cgget -r memory.oom_control limits-probe-very-own
    limits-probe-very-own:
    memory.oom_control: oom_kill_disable 0
            under_oom 0
            oom_kill 1

### Setting memory limit via systemd
Note that, if the Linux system uses the legacy control group hierarchy (cgroup v1), then the systemd user instance will not support resource control, as the legacy cgroup v1 does not allow safe delegation of controllers to unprivileged processes.

This exercise assumes that system is using cgroup v1.

    > sudo systemd-run --pty  -p MemoryMax=536870912 --working-directory=(readlink -f .) -E HOME=/root/ bash 
    > cgset -r memory.swappiness=0 "$(grep memory /proc/self/cgroup | cut -d ':' -f 3)"
    > go run github.com/HouzuoGuo/limits-probe/cmd/limits-probe -ex mem
    signal: killed

## Maxing out external processes by spawning "/usr/bin/sleep"
### See ulimit in live action
Note that the NPROC resource limit counts the number of processes belong to that real user ID across all sessions.

    > ulimit -u 123 -S
    > go run github.com/HouzuoGuo/limits-probe/cmd/limits-probe -ex exec
    2020/11/01 16:57:11 Spawned 1 external processes
    2020/11/01 16:57:11 Spawned 11 external processes
    ...
    2020/11/01 16:57:11 Successfully made 45 external processes and then encountered failure: fork/exec /usr/bin/sleep: resource temporarily unavailable

### Setting ulimit via systemd
    > systemd-run --user --pty -p LimitNPROC=123 --working-directory="$(readlink -f .)" go run github.com/HouzuoGuo/limits-probe/cmd/limits-probe -ex exec
    (Same output as above)

### See cgroup task limit in action
The PIDs controller only counts the number of processes belonging to that PIDs controller itself.

    > ulimit -u 9999 -S
    > sudo cgcreate -a howard -t howard -g pids:/limits-probe-very-own
    > cgset -r pids.max=123 limits-probe-very-own
    > cgget -g pids /limits-probe-very-own
    > cgexec -g pids:/limits-probe-very-own go run github.com/HouzuoGuo/limits-probe/cmd/limits-probe -ex exec
    2020/11/01 17:06:42 Spawned 1 external processes
    2020/11/01 17:06:42 Spawned 11 external processes
    ...
    2020/11/01 17:06:42 Successfully made 104 external processes and then encountered failure: fork/exec /usr/bin/sleep: resource temporarily unavailable

### Setting task limit via systemd
Similar to the case with setting memory limit via systemd, the legacy control group hierarchy (cgroup v1) does not seem to allow delegation of PIDs controller to unprivileged processes.

    > sudo systemd-run --pty -p TasksMax=123 --working-directory=(readlink -f .) -E HOME=/root/ go run github.com/HouzuoGuo/limits-probe/cmd/limits-probe -ex exec
    ...
    2020/11/01 17:11:34 Successfully made 109 external processes and then encountered failure: fork/exec /usr/bin/sleep: resource temporarily unavailable
