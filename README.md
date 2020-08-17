# golasticdump
tool for moving elasticsearch data written in golang

[![Build Status](https://travis-ci.org/tsaikd/golasticdump.svg?branch=master)](https://travis-ci.org/tsaikd/golasticdump)

## Install
* with docker image
	* [tsaikd/golasticdump](https://registry.hub.docker.com/u/tsaikd/golasticdump/)
```
docker pull tsaikd/golasticdump:0
```
* with source code and golang
```
go get github.com/tsaikd/golasticdump
```

## Examples

* Copy an index from production to staging
```
golasticdump \
	--input="http://production.es.com:9200/my_index" \
	--output="http://staging.es.com:9200/my_index"
```

* Copy an index from production to file
```
golasticdump \
	--input="http://production.es.com:9200/my_index" \
	--output="my_index.dump"
```

* Copy indices
```
golasticdump \
	--input="http://production.es.com:9200/my_index-*" \
	--output="http://staging.es.com:9200"
```

* Move indices
```
golasticdump --delete \
	--input="http://production.es.com:9200/my_index-*" \
	--output="http://staging.es.com:9200"
```

* Merge indices
```
golasticdump --delete \
	--input="http://production.es.com:9200/my_index-*" \
	--output="http://staging.es.com:9200/merged_index"
```

## Reference

inspired by [taskrabbit/elasticsearch-dump](https://github.com/taskrabbit/elasticsearch-dump)
