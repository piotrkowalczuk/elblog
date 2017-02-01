# elblog [![GoDoc](https://godoc.org/github.com/piotrkowalczuk/elblog?status.svg)](https://godoc.org/github.com/piotrkowalczuk/elblog)

Library helps to parse [ELB](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/access-log-collection.html) logs. 

Elastic Load Balancing provides access logs that capture detailed information about requests sent to your load balancer. 
Each log contains information such as the time the request was received, the client's IP address, latencies, request paths, and server responses. 
You can use these access logs to analyze traffic patterns and to troubleshoot issues.