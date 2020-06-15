// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START trace_setup_go_quickstart]

// Sample trace_quickstart traces incoming and outgoing requests.
package main

import (
	"log"
	"net/http"
	"os"
	"time"
	"go.opencensus.io/stats/view"
    // "contrib.go.opencensus.io/resource/auto"
    "contrib.go.opencensus.io/resource/gke"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"go.opencensus.io/resource"
)

func main() {
	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)
    exporter.StartMetricsExporter()
	view.SetReportingPeriod(60 * time.Second)
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		log.Printf("Error registering default server views")
	} else {
		log.Printf("Registered default server views Duke")
	}	

	// By default, traces will be sampled relatively rarely. To change the
	// sampling frequency for your entire program, call ApplyConfig. Use a
	// ProbabilitySampler to sample a subset of traces, or use AlwaysSample to
	// collect a trace on every run.
	//
	// Be careful about using trace.AlwaysSample in a production application
	// with significant traffic: a new trace will be started and exported for
	// every request.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	client := &http.Client{
		Transport: &ochttp.Transport{
			// Use Google Cloud propagation format.
			Propagation: &propagation.HTTPFormat{},
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest("GET", "https://www.google.com", nil)

		// The trace ID from the incoming request will be
		// propagated to the outgoing request.
		req = req.WithContext(r.Context())

		span1 := trace.FromContext(r.Context())

		abc, err := gke.Detect(r.Context())
		if err != nil {
			log.Print(err)
		} else {
			log.Print(abc.Type)
			for k, v := range abc.Labels { 
				log.Printf("key[%s] value[%s]\n", k, v)
				attr := trace.StringAttribute(k,v)
				span1.AddAttributes(attr)
			}
		}

		if span1 != nil {
			log.Printf("%+v\n", span1)
		}

		env, err := resource.FromEnv(r.Context())
		log.Printf("%+v\n", env)

		// The outgoing request will be traced with r's trace ID.
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		// Because we don't read the resp.Body, need to manually call Close().
		resp.Body.Close()
	})
	http.Handle("/foo", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, &ochttp.Handler {
		Propagation: &propagation.HTTPFormat{},
		}); err != nil {
		log.Fatal(err)
	}
}

// [END trace_setup_go_quickstart]
