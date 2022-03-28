# microprobe

## run elastic search
```shell
cd deploy/elastic-search
docker-compose up
```

## execute dataloader
```shell
 go run cmd/microprobe.go --config config/metric_elastic.yml
```
generate outliers every second (`humidity == 200`)

## execute query
```shell
go run cmd/microprobe.go --config config/query_elastic.yml
```

need to skip first 5 results, cause we execute busy query (dead loop) after ingest 5 seconds.
