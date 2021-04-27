{
  "this field is not": "allowed by any means",
  "identifier": {
    "url": "http://test-network-function.com/tests/unit/ping",
    "version": "v1.0.0"
  },
  "description": "ping test",
  "reelFirstStep": {
    "execute": "ping -c 5 {{.HOST}}",
    "expect": [
      "(?m)(\\d+) packets transmitted, (\\d+)( packets){0,1} received, (?:\\+(\\d+) errors)?.*$"
    ],
    "timeout": 2000000000
  },
  "resultContexts": [
    {
      "pattern": "(?m)(\\d+) packets transmitted, (\\d+)( packets){0,1} received, (?:\\+(\\d+) errors)?.*$",
      "defaultResult": 0
    }
  ],
  "testResult": 2,
  "testTimeout": 2000000000
}
