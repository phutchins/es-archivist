# es-archivist
Shipping logs or other data to Elasticsearch creates the never ending battle of removing your oldest indices and/or cleaning up unwanted data to keep from running out of space in your cluster. This tool is designed to automate this process in a smart and configurable way allowing you to spend your time doing other things like manually griding you coffee beans or learning a new programming language.

## Spec

+ Watch the total amount of disk space available
`curl -XGET 'http://localhost:9200/_nodes/storj-prod-1/stats' | jq '.nodes [] .fs.total.available_in_bytes'`
+
