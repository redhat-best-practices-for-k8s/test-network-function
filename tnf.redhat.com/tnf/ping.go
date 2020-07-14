package tnf

import (
    "strconv"
    "regexp"
    "tnf.redhat.com/reel"
)

type Ping struct {
    result  int
    timeout int
    args    []string
}

const stat string = `(?m)^\D(\d+) packets transmitted, (\d+) received, (?:\+(\d+) errors)?.*$`
const done string = `\D\d+ packets transmitted.*\r\n(?:rtt )?.*$`

func (ping *Ping) Args() []string {
    return ping.args
}
func (ping *Ping) Timeout() (int) {
    return ping.timeout
}
func (ping *Ping) Result() (int) {
    return ping.result
}
func (ping *Ping) ReelFirst() (*reel.Step) {
    return &reel.Step{
        Expect: []string{done},
        Timeout: ping.timeout,
    }
}
func (ping *Ping) ReelMatch(pattern string, before string, match string) (*reel.Step) {
    re := regexp.MustCompile(stat)
    matched := re.FindStringSubmatch(match)
    if matched != nil {
        var txd, rxd, ers int
        txd, _ = strconv.Atoi(matched[1])
        rxd, _ = strconv.Atoi(matched[2])
        ers, _ = strconv.Atoi(matched[3])
        switch {
        case txd == 0 || ers > 0:
            ping.result = ERROR
        case txd == rxd:
            ping.result = SUCCESS
        case rxd > 0 && txd - rxd <= 1:
            ping.result = SUCCESS
        default:
            ping.result = FAILURE
        }
    }
    return nil
}
func (ping *Ping) ReelTimeout() (*reel.Step) {
    return &reel.Step{
        Execute: "\003", // ^C
        Expect: []string{done},
    }
}
func (ping *Ping) ReelEof() (*reel.Step) {
    return nil
}

func PingCmd(host string, count int) []string {
    if count > 0 {
        return []string{"ping", "-c", strconv.Itoa(count), host}
    } else {
        return []string{"ping", host}
    }
}

func NewPing(timeout int, host string, count int) (*Ping) {
    return &Ping{
        result:  ERROR,
        timeout: timeout,
        args:    PingCmd(host, count),
    }
}
