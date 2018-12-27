# Configuration

The prom_rest_exporter configuration is a YAML file. It defines one or more `/metrics` endpoints that transform REST responses into metrics for Prometheus.

1. [Options](#options)  
  1.1. [Global options](#global-options)  
  1.2. [Endpoint options](#endpoint-options)  
  1.3. [Target options](#target-options)  
  1.4. [Metric options](#metric-options)  
  1.5. [Label options](#label-options)  
2. [Jq programs](#jq-programs)  
3. [Examples](#examples)  
  3.1. [Simple example](#simple-example)  
  3.2. [Multi-value metric example](#multi-value-metric-example)  
  3.3. [Full example](#full-example)

## Options
### Global options

| Option        | Required | Description                                                  |
| ------------- | -------- | ------------------------------------------------------------ |
| **endpoints** | Yes      | List of Endpoint options                                     |
| cache_time    | No       | Number of seconds to cache last result per `/metrics` endpoint |

### Endpoint options

| Option       | Required | Description                                          |
| ------------ | -------- | ---------------------------------------------------- |
| **port**     | Yes      | Port to listen for `/metrics` requests from Prometheus |
| **targets**  | Yes      | List of Target options                               |
| host         | No       | Host to listen for `/metrics` requests from Prometheus. Default: `localhost` |
| meta_metrics | No       | If true, includes additional meta metrics like REST response times and number of collected metrics. |
| cache_time   | No       | Number of seconds to cache last result for this `/metrics` endpoint. Overrides global cache time.            |

### Target options

| Option      | Required | Description                               |
| ----------- | -------- | ----------------------------------------- |
| **url**     | Yes      | REST URL from which to fetch data         |
| **metrics** | Yes      | List of Metric options                    |
| user        | No       | Username for basic authentication         |
| password    | No       | Password for basic authentication         |
| headers     | No       | Additional headers to add to REST request |

### Metric options

| Option       | Required | Description                                       |
| ------------ | -------- | ------------------------------------------------- |
| **name**     | Yes      | Name of the metric                                |
| **selector** | Yes      | jq program to extract value(s) from REST response |
| val_selector | No       | jq program applied to each extracted value to get numeric value. Default: `.` |
| description  | No       | Metric description added as HELP comment to `/metrics` response |
| type         | No       | Metric type added as TYPE comment to `/metrics` response |
| labels       | No       | List of Label options                             |

### Label options

| Option       | Required                | Description               |
| ------------ | ----------------------- | ------------------------- |
| **name**     | Yes                     | Name of the label         |
| selector     | selector or fixed_value | jq program applied to each extracted value to get label value |
| fixed_value  | selector or fixed_value | Fixed value for the label |

## Jq programs

`selector` fields use [jq](https://github.com/stedolan/jq) program syntax to extract data from JSON.

See the [jq manual](http://stedolan.github.io/jq/manual/) for more information.

Hint: one of the reasons for using jq in prom_rest_exporter is that it is easy to try out the selectors using the standard jq command-line tool.
E.g.
```bash
curl -s https://reqres.in/api/users | jq '.total'
```

## Examples

### Simple example

```yaml
endpoints:
  - port: 9011
    targets:
      - url: https://reqres.in/api/users
        metrics:
          - name: user_count
            description: Number of users
            type: gauge
            selector: ".total"
```

This configuration exposes a `http://localhost:9011/metrics` endpoint.  
When this is called, a REST call is made to `https://reqres.in/api/users`.  
One metric `user_count` is extracted from the REST response using the jq program
to get the user count from the `total` field.

Let's look at the REST input and metrics output of this:

**REST response**
```json
{
  "page": 1,
  "per_page": 3,
  "total": 12,
  "total_pages": 4,
  "data": [
    {
      "id": 11,
      "first_name": "George",
      "last_name": "Bluth",
      "avatar": "https://s3.amazonaws.com/uifaces/faces/twitter/calebogden/128.jpg"
    },
    {
      "id": 22,
      "first_name": "Janet",
      "last_name": "Weaver",
      "avatar": "https://s3.amazonaws.com/uifaces/faces/twitter/josephstein/128.jpg"
    },
    {
      "id": 33,
      "first_name": "Emma",
      "last_name": "Wong",
      "avatar": "https://s3.amazonaws.com/uifaces/faces/twitter/olegpogodaev/128.jpg"
    }
  ]
}
```

**Metrics output**
```
# HELP user_count Number of users
# TYPE user_count gauge
user_count 12
```

### Multi-value metric example

It's possible to extract metrics with multiple values, distinguished by labels.

```yaml
endpoints:
  - port: 9011
    targets:
      - url: https://reqres.in/api/users
        metrics:
          - name: user_id
            description: User ids
            type: gauge
            selector: ".data[]"
            val_selector: ".id"
            labels:
              - name: last_name
                selector: ".last_name"
```

This configuration extracts multiple values from the REST response using `selector`.  
These are not numeric values yet, but rather JSON objects in this case. We want to
use these to extract actual values *and* label values.  
`val_selector` is used to extract a numeric value from each base value.  
In the labels section, `selector` is used to extract a label from each base value.

Using the same REST input from the simple example, here is how the metrics output would look:
```
# HELP user_id User ids
# TYPE user_id gauge
user_id{last_name="Bluth"} 11
user_id{last_name="Weaver"} 22
user_id{last_name="Wong"} 33
```

If you do not define labels to distinguish the metric's values, a `val_index` label is used by default:
```
# HELP user_id User ids
# TYPE user_id gauge
user_id{val_index="1"} 11
user_id{val_index="2"} 22
user_id{val_index="3"} 33
```

### Full example

Here's a configuration using all possible options.

```yaml

# Global cache time in seconds
cache_time: 60
endpoints:
  # Port to run /metrics endpoint on
  - port: 9011
    # Host on 0.0.0.0 instead of localhost
    host: 0.0.0.0
    # Cache time for just this endpoint in seconds
    cache_time: 30
    # Include meta metrics like response times
    meta_metrics: yes
    targets:
      # REST endpoint to get data from
      - url: https://reqres.in/api/users
        # Basic auth credentials for the REST request
        user: user123
        password: pass123
        # Additional HTTP headers for the REST request
        headers:
          My-Header: my-value
        # Metrics to create from the REST data
        metrics:
          - name: user_count
            description: Number of users
            # Metric type: gauge or counter
            type: gauge
            # jq program to extract a numeric value from REST response
            selector: "[.data[].last_name] | length"
            # Labels to add to metric
            labels:            
              - name: env
                fixed_value: prod
          # Metric with multiple values:
          - name: user_id
            description: User ids
            type: gauge
            # jq program to extract multiple
            # arbitrary values from REST response
            selector: ".data[]"
            # jq program to extract numeric value
            # from each of the values returned by above selector  
            val_selector: ".id"
            labels:
              - name: last_name
                # jq program to extract label value
                # from each of the values returned by the metric selector
                selector: ".last_name"
      # Second REST endpoint to get data from for this /metrics endpoint
      - url: https://reqres.in/api/unknown
        metrics:
          - name: years_total
            description: Total number of years
            type: gauge
            selector: "[.data[].year] | add"
  # Second /metrics endpoint running on port 9012
  - port: 9012
    cache_time: 10
    targets:
      - url: https://reqres.in/api/unknown
        metrics:
          - name: years_total
            description: Total number of years
            type: gauge
            selector: "[.data[].year] | add"
```