# Test Network Function

This repository contains a set of network function test cases.

## Install

0. Install packages for dependencies: `expect` and `tcllib`.
1. Ensure `bin/reel.exp` is on your PATH.

## Command Line Tools

A set of command line tools is provided, where each tool wraps a test or some
aspect of testing. The tools allow a test scenario to be manually invoked, which
is useful in development and debugging.

Where a test is implemented using a controlled subprocess, the corresponding
command line tool provides a `-d` option. Specifying this option captures the
dialog between the controlling process and controlled subprocess to the given
filename.

### ping

Using default options, the `ping` tool sends a single ICMP Echo Request to the
target host.

```bash
$ go run cmd/ping/main.go -d success.log 10.5.0.3
PING 10.5.0.3 (10.5.0.3) 56(84) bytes of data.
64 bytes from 10.5.0.3: icmp_seq=1 ttl=63 time=38.6 ms

--- 10.5.0.3 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 38.646/38.646/38.646/0.000 ms
```

The tool writes to stdout the text output by the controlled subprocess.
The exit code is 0 (zero) when the test result is success.

```bash
$ echo $?
0
```

For this example, `success.log` contains the dialog observed by `bin/reel.exp`.
This comprises each line of JSON sent by the controlling process to instruct
`bin/reel.exp` of the next step; the text output by the controlled subprocess;
each line of JSON sent by `bin/reel.exp` to notify the controlling process of
an event.

```bash
$ cat success.log
{"expect":["\\D\\d+ packets transmitted.*\\r\\n(?:rtt )?.*$"],"timeout":2}
PING 10.5.0.3 (10.5.0.3) 56(84) bytes of data.
64 bytes from 10.5.0.3: icmp_seq=1 ttl=63 time=38.6 ms

--- 10.5.0.3 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 38.646/38.646/38.646/0.000 ms
{"event":"match","idx":0,"pattern":"\\D\\d+ packets transmitted.*\\r\\n(?:rtt )?.*$","before":"PING 10.5.0.3 (10.5.0.3) 56(84) bytes of data.\r\n64 bytes from 10.5.0.3: icmp_seq=1 ttl=63 time=38.6 ms\r\n\r\n--- 10.5.0.3 ping statistics ---\r","match":"\n1 packets transmitted, 1 received, 0% packet loss, time 0ms\r\nrtt min/avg/max/mdev = 38.646/38.646/38.646/0.000 ms\r\n"}
```

The following example shows the controlling process giving up on the test due
to a timeout and killing the controlled subprocess.

```bash
$ go run cmd/ping/main.go -d failure.log 10.3.0.99
(timeout)
PING 10.3.0.99 (10.3.0.99) 56(84) bytes of data.
^C

--- 10.3.0.99 ping statistics ---
1 packets transmitted, 0 received, 100% packet loss, time 0ms

exit status 1
```

The exit code is 1 (one) when the test result is failure (the test was correctly
executed and a negative result was determined).

The dialog clearly shows the timeout and sending of ^C to kill the subprocess.

```bash
$ cat failure.log
{"expect":["\\D\\d+ packets transmitted.*\\r\\n(?:rtt )?.*$"],"timeout":2}
PING 10.3.0.99 (10.3.0.99) 56(84) bytes of data.
{"event":"timeout"}
{"execute":"\u0003","expect":["\\D\\d+ packets transmitted.*\\r\\n(?:rtt )?.*$"]}
^C

--- 10.3.0.99 ping statistics ---
1 packets transmitted, 0 received, 100% packet loss, time 0ms

{"event":"match","idx":0,"pattern":"\\D\\d+ packets transmitted.*\\r\\n(?:rtt )?.*$","before":"PING 10.3.0.99 (10.3.0.99) 56(84) bytes of data.\r\n^C\r\n\r\n--- 10.3.0.99 ping statistics ---\r","match":"\n1 packets transmitted, 0 received, 100% packet loss, time 0ms\r\n\r\n"}
```

The final example shows a test-specific error in executing the test.

```bash
$ go run cmd/ping/main.go -c 8 -t 10 10.5.0.99
PING 10.5.0.99 (10.5.0.99) 56(84) bytes of data.
From 10.3.0.1 icmp_seq=1 Destination Host Unreachable
From 10.3.0.1 icmp_seq=2 Destination Host Unreachable
From 10.3.0.1 icmp_seq=3 Destination Host Unreachable
From 10.3.0.1 icmp_seq=4 Destination Host Unreachable
From 10.3.0.1 icmp_seq=5 Destination Host Unreachable
From 10.3.0.1 icmp_seq=6 Destination Host Unreachable
From 10.3.0.1 icmp_seq=7 Destination Host Unreachable
From 10.3.0.1 icmp_seq=8 Destination Host Unreachable

--- 10.5.0.99 ping statistics ---
8 packets transmitted, 0 received, +8 errors, 100% packet loss, time 7121ms
pipe 4
exit status 2
```

The exit code is 2 (two) when the test result is error (the test was not
correctly executed; no result could be determined).

### ssh

Using default options, the `ssh` tool simply establishes a SSH session to the
target host, then closes it. The controlled subprocess is the native `ssh`
client tool, with its command line options and args passed through. The
session is closed when the supplied prompt string (regex) is matched.

```bash
$ env TERM=vt220 go run cmd/ssh/main.go 'user@hhh:\S+\$ ' hhh -o 'PreferredAuthentications=publickey'
Last login: Thu Jul 23 18:10:34 2020 from 10.3.0.109
user@hhh:~$ logout
Connection to hhh.ddd closed.
```

Using `-f lines` allows interactive execution of lines of text from stdin,
mimicking the native client tool.

In this example, two commands are supplied from a here document.

```bash
$ env TERM=vt220 go run cmd/ssh/main.go -d ssh.log -f lines 'user@hhh:\S+\$ ' hhh -o 'PreferredAuthentications=publickey' <<EOF
> echo foobar
> date
> EOF
Last login: Tue Jul 14 15:58:14 2020 from 10.3.0.109
user@hhh:~$ echo foobar
foobar
user@hhh:~$ date
Tue 14 Jul 15:59:08 BST 2020
user@hhh:~$ logout

$ echo $?
0

$ cat ssh.log
{"expect":["Are you sure you want to continue connecting \\(yes/no\\)\\?","Please type 'yes' or 'no': ","user@hhh:\\S+\\$ "],"timeout":2}
Last login: Tue Jul 14 15:58:14 2020 from 10.3.0.109
user@hhh:~$ {"event":"match","idx":2,"pattern":"user@hhh:\\S+\\$ ","before":"Last login: Tue Jul 14 15:58:14 2020 from 10.3.0.109\r\r\n","match":"user@hhh:~$ "}
{"execute":"echo foobar","expect":["user@hhh:\\S+\\$ "],"timeout":2}
echo foobar
foobar
user@hhh:~$ {"event":"match","idx":0,"pattern":"user@hhh:\\S+\\$ ","before":"echo foobar\r\nfoobar\r\n","match":"user@hhh:~$ "}
{"execute":"date","expect":["user@hhh:\\S+\\$ "],"timeout":2}
date
Tue 14 Jul 15:59:08 BST 2020
user@hhh:~$ {"event":"match","idx":0,"pattern":"user@hhh:\\S+\\$ ","before":"date\r\nTue 14 Jul 15:59:08 BST 2020\r\n","match":"user@hhh:~$ "}
{"execute":"\u0004","expect":["Connection to .+ closed\\..*$"],"timeout":2}
logout
{"event":"match","idx":0,"pattern":"Connection to .+ closed\\..*$","before":"logout\r\n","match":"Connection to hhh.ddd closed.\r"}
```

Using `-f tests` allows execution of test configurations; each line from stdin
contains a JSON test configuration.

```bash
$ cat tests.json
{"test": "https://tnf.redhat.com/ping/one", "host": "ggg"}
{"test": "https://tnf.redhat.com/ping/flexi", "count": 77, "host": "10.3.0.99"}

$ env TERM=vt220 go run cmd/ssh/main.go -d ssh.log -f tests 'user@hhh:\S+\$ ' hhh -o 'PreferredAuthentications=publickey' <tests.json
Last login: Tue Jul 14 15:59:07 2020 from 10.3.0.109
user@hhh:~$ ping -c 1 ggg
PING ggg.ddd (10.5.0.5) 56(84) bytes of data.
64 bytes from ggg.ddd (10.5.0.5): icmp_seq=1 ttl=64 time=0.344 ms

--- ggg.ddd ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 0.344/0.344/0.344/0.000 ms
user@hhh:~$ (timeout)
ping -c 77 10.3.0.99
PING 10.3.0.99 (10.3.0.99) 56(84) bytes of data.
From 10.5.0.1 icmp_seq=1 Destination Host Unreachable
From 10.5.0.1 icmp_seq=2 Destination Host Unreachable
From 10.5.0.1 icmp_seq=3 Destination Host Unreachable
From 10.5.0.1 icmp_seq=4 Destination Host Unreachable
From 10.5.0.1 icmp_seq=5 Destination Host Unreachable
From 10.5.0.1 icmp_seq=6 Destination Host Unreachable
From 10.5.0.1 icmp_seq=7 Destination Host Unreachable
From 10.5.0.1 icmp_seq=8 Destination Host Unreachable
^C

--- 10.3.0.99 ping statistics ---
11 packets transmitted, 0 received, +8 errors, 100% packet loss, time 10183ms
pipe 4
user@hhh:~$ 
user@hhh:~$ logout

$ cat ssh.log
{"expect":["Are you sure you want to continue connecting \\(yes/no\\)\\?","Please type 'yes' or 'no': ","user@hhh:\\S+\\$ "],"timeout":2}
Last login: Tue Jul 14 15:59:07 2020 from 10.3.0.109
user@hhh:~$ {"event":"match","idx":2,"pattern":"user@hhh:\\S+\\$ ","before":"Last login: Tue Jul 14 15:59:07 2020 from 10.3.0.109\r\r\n","match":"user@hhh:~$ "}
{"execute":"ping -c 1 ggg","expect":["user@hhh:\\S+\\$ "],"timeout":2}
ping -c 1 ggg
PING ggg.ddd (10.5.0.5) 56(84) bytes of data.
64 bytes from ggg.ddd (10.5.0.5): icmp_seq=1 ttl=64 time=0.344 ms

--- ggg.ddd ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 0.344/0.344/0.344/0.000 ms
user@hhh:~$ {"event":"match","idx":0,"pattern":"user@hhh:\\S+\\$ ","before":"ping -c 1 ggg\r\nPING ggg.ddd (10.5.0.5) 56(84) bytes of data.\r\n64 bytes from ggg.ddd (10.5.0.5): icmp_seq=1 ttl=64 time=0.344 ms\r\n\r\n--- ggg.ddd ping statistics ---\r\n1 packets transmitted, 1 received, 0% packet loss, time 0ms\r\nrtt min/avg/max/mdev = 0.344/0.344/0.344/0.000 ms\r\n","match":"user@hhh:~$ "}
{"execute":"ping -c 77 10.3.0.99","expect":["user@hhh:\\S+\\$ "],"timeout":10}
ping -c 77 10.3.0.99
PING 10.3.0.99 (10.3.0.99) 56(84) bytes of data.
From 10.5.0.1 icmp_seq=1 Destination Host Unreachable
From 10.5.0.1 icmp_seq=2 Destination Host Unreachable
From 10.5.0.1 icmp_seq=3 Destination Host Unreachable
From 10.5.0.1 icmp_seq=4 Destination Host Unreachable
From 10.5.0.1 icmp_seq=5 Destination Host Unreachable
From 10.5.0.1 icmp_seq=6 Destination Host Unreachable
From 10.5.0.1 icmp_seq=7 Destination Host Unreachable
From 10.5.0.1 icmp_seq=8 Destination Host Unreachable
{"event":"timeout"}
{"execute":"\u0003","expect":["user@hhh:\\S+\\$ "],"timeout":2}
^C

--- 10.3.0.99 ping statistics ---
11 packets transmitted, 0 received, +8 errors, 100% packet loss, time 10183ms
pipe 4
user@hhh:~$ 
user@hhh:~$ {"event":"match","idx":0,"pattern":"user@hhh:\\S+\\$ ","before":"ping -c 77 10.3.0.99\r\nPING 10.3.0.99 (10.3.0.99) 56(84) bytes of data.\r\nFrom 10.5.0.1 icmp_seq=1 Destination Host Unreachable\r\nFrom 10.5.0.1 icmp_seq=2 Destination Host Unreachable\r\nFrom 10.5.0.1 icmp_seq=3 Destination Host Unreachable\r\nFrom 10.5.0.1 icmp_seq=4 Destination Host Unreachable\r\nFrom 10.5.0.1 icmp_seq=5 Destination Host Unreachable\r\nFrom 10.5.0.1 icmp_seq=6 Destination Host Unreachable\r\nFrom 10.5.0.1 icmp_seq=7 Destination Host Unreachable\r\nFrom 10.5.0.1 icmp_seq=8 Destination Host Unreachable\r\n^C\r\n\r\n--- 10.3.0.99 ping statistics ---\r\n11 packets transmitted, 0 received, +8 errors, 100% packet loss, time 10183ms\r\npipe 4\r\nuser@hhh:~$ \r\n","match":"user@hhh:~$ "}
{"execute":"\u0004","expect":["Connection to .+ closed\\..*$"],"timeout":2}
logout
Connection to hhh.ddd closed.
{"event":"match","idx":0,"pattern":"Connection to .+ closed\\..*$","before":"logout\r\n","match":"Connection to hhh.ddd closed.\r"}
```
