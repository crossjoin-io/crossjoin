# crossjoin [![Docker](https://github.com/crossjoin-io/crossjoin/actions/workflows/docker.yml/badge.svg)](https://github.com/crossjoin-io/crossjoin/actions/workflows/docker.yml) [![CLI](https://github.com/crossjoin-io/crossjoin/actions/workflows/go.yml/badge.svg)](https://github.com/crossjoin-io/crossjoin/actions/workflows/go.yml) [![Security scan](https://github.com/crossjoin-io/crossjoin/actions/workflows/shiftleft.yml/badge.svg)](https://github.com/crossjoin-io/crossjoin/blob/main/SECURITY.md)

Crossjoin joins together your data from anywhere.

- Supports PostgreSQL, Redshift, CSV data sources
- Zero dependency CLI, or a single Docker container

## Example

In the [example](https://github.com/crossjoin-io/crossjoin/tree/main/example) directory, there are two CSVs (adapted from
this [AWS blog post](https://aws.amazon.com/blogs/big-data/joining-across-data-sources-on-amazon-quicksight/)) representing
orders and returns data.

The config creates a combined data set in a `joined.db` SQLite3 file.

```yaml
data_sets:
  - name: joined
    data_source:
      name: orders
      type: csv
      path: ./orders.csv
    joins:
      - type: JOIN
        columns:
          - left_column: Order ID
            right_column: Order ID
        data_source:
          name: returns
          type: csv
          path: ./returns.csv
```

```
$ crossjoin --config ./config.yaml
2021/10/14 18:08:06 using config file path config.yaml
2021/10/14 18:08:06 starting crossjoin
2021/10/14 18:08:06 creating data set `joined`
2021/10/14 18:08:06 querying `orders`
2021/10/14 18:08:06 querying `returns`
2021/10/14 18:08:06 joining data
2021/10/14 18:08:06 finished crossjoin
```
