# Discovering Prometheus

## Installation

```sh
podman-compose up -d
go run main.go scrape
```

## Usage

```sh
$ curl -s -X POST http://localhost:9090/api/v1/query_range -d query='myapp_processed_ops_total' --data-urlencode start="$(date -Iseconds -u --date='1 minute ago')" --data-urlencode end="$(date -Iseconds -u)" -d step=5s | jq .

{
  "status": "success",
  "data": {
    "resultType": "matrix",
    "result": [
      {
        "metric": {
          "__name__": "myapp_processed_ops_total",
          "instance": "192.168.122.1:2112",
          "job": "test-exporter"
        },
        "values": [
          [
            1656171931,
            "359"
          ],
          [
            1656171936,
            "361"
          ],
          [
            1656171941,
            "364"
          ],
          [
            1656171946,
            "367"
          ],
          [
            1656171951,
            "369"
          ],
          [
            1656171956,
            "371"
          ],
          [
            1656171961,
            "374"
          ],
          [
            1656171966,
            "376"
          ],
          [
            1656171971,
            "379"
          ],
          [
            1656171976,
            "381"
          ],
          [
            1656171981,
            "384"
          ],
          [
            1656171986,
            "386"
          ],
          [
            1656171991,
            "389"
          ]
        ]
      }
    ]
  }
}
```

