package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"os"
	"strconv"
)

var (
	count = 0
)

func init() {
	flag.IntVar(&count, "count", 100, "number of becaons to generate")
}

type M map[string]interface{}

func main() {
	flag.Parse()
	e := json.NewEncoder(os.Stdout)
	out := make([]M, 0)
	var base int64
	base = 300234010030450
	for i := 1; i <= count; i++ {
		v := M{
			"device": "omnicom",
			"imei":   strconv.FormatInt(base + int64(i), 10),
			"pos": []M{
				M{
					"latitude":  between(29.7, 38.5),
					"longitude": between(-13.5, 0.3),
					"speed":     between(1, 30),
				},
				M{
					"latitude":  between(29.7, 38.5),
					"longitude": between(-13.5, 0.3),
					"speed":     between(1, 30),
				},
			},
			"reportperiod": 1,
		}
		out = append(out, v)
	}
	e.Encode(M{"objects": out})
}

func between(lo float64, hi float64) float64 {
	offset := (hi - lo) * rand.Float64()
	return lo + offset
}
