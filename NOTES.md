# TODO
1. Set up a small webserver
2. Create our route
3. Channel vs atomic counter ?? 
4. Create a function to call the 3rd party service
5. Sort data
6. Format the response
7. Create a caching layer 
8. TESTS
9. Docker compose ?

# SORTING
1. Driving time
2. Distance

# ROUTE
/routes?
src=lat,long
[]dst=lat,long

# RESPONSE
```json
{
  "source": "13.388860,52.517037",
  "routes": [
    {
      "destination": "13.397634,52.529407",
      "duration": 465.2,
      "distance": 1879.4
    },
    {
      "destination": "13.428555,52.523219",
      "duration": 712.6,
      "distance": 4123
    }
  ]
}
```

# 3RD Party service

```bash
$ curl 'http://router.project-osrm.org/route/v1/driving/13.388860,52.517037;13.397634,52.5
29407?overview=false'
```
 
 Response

 ```json
{
  "code": "Ok",
  "routes": [
    {
      "legs": [
        {
          "steps": [],
          "summary": "",
          "weight": 263,
          "duration": 260.1,
          "distance": 1886.3
        }
      ],
      "weight_name": "routability",
      "weight": 263,
      "duration": 260.1,
      "distance": 1886.3
    }
  ],
  "waypoints": [
    {
      "hint": "hv4JgOLXmoUXAAAABQAAAAAAAAAgAAAAIXRPQYXNK0AAAAAAcPePQQsAAAADAAAAAAAAABAAAAAC_wAA_kvMAKlYIQM8TMwArVghAwAA7wpImaGf",
      "distance": 4.231521214,
      "name": "Friedrichstraße",
      "location": [
        13.388798,
        52.517033
      ]
    },
    {
      "hint": "aGvdgfcUi4cGAAAACQAAAAAAAAB3AAAAppONQCyhu0AAAAAA1J-EQgYAAAAJAAAAAAAAAHcAAAAC_wAAfm7MABiJIQOCbswA_4ghAwAAXwVImaGf",
      "distance": 2.795148358,
      "name": "Torstraße",
      "location": [
        13.39763,
        52.529432
      ]
    }
  ]
}
 
 ```

 # Designing with microservices in mind
 # No ratelimit headers present on the demo api
 # slog doesn't work with golangs test package.. whoops
 # note on ratelimiting 
 # note on 400+500 code string returns
 # why no cordinate validation

 # Config variables / parameters ?
 - request queue length
 - serve IP / Port
 - Duration Request timeout
 - Total Request timeout 
 - Channel Processor sleep timer



 # Making requests async
> HTTP GET > ?x HTTP GET



curl 'http://127.0.0.1/routes?src=13.388860,52.517037&dst=13.397634,52.529407&dst=13.428555,52.523219&dst=13.428555,52.523219'

curl -TimeoutSec 1 'http://127.0.0.1/routes?src=13.388860,52.517037&dst=13.397634,52.529407&dst=13.428555,52.523219&dst=11.428555,51.523219&dst=13.428255,52.533219&dst=13.421555,52.523239'