package tnf

import (
    "encoding/json"
)

type TestConfig struct {
    Test    string  `json:"test"`
    Timeout int     `json:"timeout"`
    Count   int     `json:"count"`
    Host    string  `json:"host"`
}

type TesterFunc func (*TestConfig)(Tester)

var Tests = map[string]TesterFunc {
    "https://tnf.redhat.com/ping/one": func (tc *TestConfig)(Tester) {
        return NewPing(2, tc.Host, 1)
    },
    "https://tnf.redhat.com/ping/flexi": func (tc *TestConfig)(Tester) {
        timeout := tc.Timeout
        if timeout < 1 {
            timeout = 10
        }
        return NewPing(timeout, tc.Host, tc.Count)
    },
}

func DecodeConfig(bytes []byte) (*TestConfig, error) {
    var tc TestConfig
    err := json.Unmarshal(bytes, &tc)
    return &tc, err
}
