# GraphQL
A powerful GraphQL server implementation for Golang. Its aim is to be the fastest GraphQL implementation.

```sh
$ cat test.go
```
```go
package main

import (
	"fmt"

	"github.com/playlyfe/go-graphql"
)

func main() {
	schema := `
	## double hashed comments be parsed as descriptions and show up in
	## introspection queries
	interface Pet {
	    name: String
	}
	# this is an internal comment
	type Dog implements Pet {
	    name: String
	    woofs: Boolean
	}
	type Cat implements Pet {
	    name: String
	    meows: Boolean
	}
	type QueryRoot {
	    pets: [Pet]
	}
	`
	resolvers := map[string]interface{}{}
	resolvers["QueryRoot/pets"] = func(params *graphql.ResolveParams) (interface{}, error) {
		return []map[string]interface{}{
			{
				"__typename": "Dog",
				"name":       "Odie",
				"woofs":      true,
			},
			{
				"__typename": "Cat",
				"name":       "Garfield",
				"meows":      false,
			},
		}, nil
	}
	context := map[string]interface{}{}
	variables := map[string]interface{}{}
	executor, err := graphql.NewExecutor(schema, "QueryRoot", "", resolvers)
	executor.ResolveType = func(value interface{}) string {
		if object, ok := value.(map[string]interface{}); ok {
			return object["__typename"].(string)
		}
		return ""
	}
	query := `{
		pets {
			name
			... on Dog {
				woofs
			}
			... on Cat {
				meows
			}
		}
	}`
	result, err := executor.Execute(context, query, variables, "")
	if err != nil {
	    panic(err)
	}
	fmt.Printf("%v", result)
}
```
## Benchmarks
```
Name                                 Repetitions   
BenchmarkGoGraphQLMaster-4             10000        230846 ns/op       29209 B/op        543 allocs/op
BenchmarkPlaylyfeGraphQLMaster-4       50000         27647 ns/op        3269 B/op         61 allocs/op
```

## More
### graphql-js master
```
wrk -t12 -c400 -d30s --timeout 10s "http://localhost:3002/graphql?query={hello}"
Running 30s test @ http://localhost:3002/graphql?query={hello}
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   219.79ms   80.35ms 613.69ms   78.38%
    Req/Sec   149.99     96.37   494.00     58.29%
  52157 requests in 30.05s, 9.96MB read
Requests/sec:   1735.60
Transfer/sec:    339.33KB

```

### graphql-go master
```
wrk -t12 -c400 -d30s --timeout 10s "http://localhost:3003/graphql?query={hello}"
Running 30s test @ http://localhost:3003/graphql?query={hello}
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   134.97ms  163.47ms   1.85s    86.12%
    Req/Sec   372.46    236.09     1.58k    70.99%
  133607 requests in 30.05s, 18.35MB read
Requests/sec:   4445.99
Transfer/sec:    625.22KB
```

### playlyfe/go-graphql master
```
wrk -t12 -c400 -d30s --timeout 10s "http://localhost:3003/graphql?query={hello}"
Running 30s test @ http://localhost:3003/graphql?query={hello}
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    34.89ms   43.72ms 518.00ms   87.58%
    Req/Sec     1.44k     0.90k    6.10k    81.35%
  514095 requests in 30.05s, 70.60MB read
Requests/sec:  17108.13
Transfer/sec:      2.35MB
```

# TODO
Validator

License
=======
Playlyfe GraphQL  
http://dev.playlyfe.com/  
Copyright(c) 2015-2016, Playlyfe IT Solutions Pvt. Ltd, support@playlyfe.com  

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:  

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.  

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

