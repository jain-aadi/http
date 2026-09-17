package main

import (
	"encoding/json"
	"http_server/internal/server"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxSimulationRequests    = 10000
	maxSimulationConcurrency = 500
)

type loadRunner struct {
	running atomic.Bool
	client  http.Client
}

type loadRequest struct {
	Requests    int `json:"requests"`
	Concurrency int `json:"concurrency"`
}

type loadResult struct {
	Requested         int     `json:"requested"`
	Completed         int64   `json:"completed"`
	Failed            int64   `json:"failed"`
	ElapsedMS         float64 `json:"elapsedMs"`
	RequestsPerSecond float64 `json:"requestsPerSecond"`
	AverageLatencyMS  float64 `json:"averageLatencyMs"`
}

// startProfiler exposes local-only diagnostic pages. It deliberately uses a
// different listener from the raw TCP server, so the two implementations do
// not share request handling code.
func startProfiler(tcpServer *server.Server) {
	runner := &loadRunner{client: http.Client{Timeout: 10 * time.Second}}

	http.HandleFunc("/debug/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(tcpServer.Stats())
	})
	http.HandleFunc("/debug/dashboard", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(dashboardHTML))
	})
	http.HandleFunc("/debug/load", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST requests only", http.StatusMethodNotAllowed)
			return
		}
		if !runner.running.CompareAndSwap(false, true) {
			http.Error(w, "a simulation is already running", http.StatusConflict)
			return
		}
		defer runner.running.Store(false)

		var request loadRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid load settings", http.StatusBadRequest)
			return
		}
		if request.Requests < 1 || request.Requests > maxSimulationRequests || request.Concurrency < 1 || request.Concurrency > maxSimulationConcurrency {
			http.Error(w, "requests must be 1-10000 and concurrency must be 1-500", http.StatusBadRequest)
			return
		}
		if request.Concurrency > request.Requests {
			request.Concurrency = request.Requests
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(runner.run(request))
	})

	go func() {
		log.Printf("pprof profiler available at http://%s/debug/pprof/", pprofAddress)
		log.Printf("live metrics dashboard available at http://%s/debug/dashboard", pprofAddress)
		if err := http.ListenAndServe(pprofAddress, nil); err != nil {
			log.Printf("diagnostics sidecar stopped: %v", err)
		}
	}()
}

func (r *loadRunner) run(request loadRequest) loadResult {
	started := time.Now()
	jobs := make(chan struct{})
	var completed, failed, totalLatencyNanos atomic.Int64
	var workers sync.WaitGroup

	for range request.Concurrency {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range jobs {
				requestStarted := time.Now()
				response, err := r.client.Get("http://127.0.0.1:8000/")
				elapsed := time.Since(requestStarted)
				totalLatencyNanos.Add(elapsed.Nanoseconds())
				completed.Add(1)
				if err != nil {
					failed.Add(1)
					continue
				}

				_, _ = io.Copy(io.Discard, response.Body)
				response.Body.Close()
				if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
					failed.Add(1)
				}
			}
		}()
	}

	for range request.Requests {
		jobs <- struct{}{}
	}
	close(jobs)
	workers.Wait()

	elapsed := time.Since(started)
	completedRequests := completed.Load()
	averageLatency := 0.0
	if completedRequests > 0 {
		averageLatency = float64(totalLatencyNanos.Load()) / float64(completedRequests) / float64(time.Millisecond)
	}

	return loadResult{
		Requested:         request.Requests,
		Completed:         completedRequests,
		Failed:            failed.Load(),
		ElapsedMS:         float64(elapsed) / float64(time.Millisecond),
		RequestsPerSecond: float64(completedRequests) / elapsed.Seconds(),
		AverageLatencyMS:  averageLatency,
	}
}

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>TCP Server Load Dashboard</title>
  <style>
    :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, sans-serif; background: #101827; color: #e5edf7; }
    * { box-sizing: border-box; }
    body { margin: 0; padding: 28px; }
    main { max-width: 1120px; margin: auto; }
    header { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 22px; }
    h1 { font-size: 24px; margin: 0; }
    .subtle { color: #9fb0c4; margin: 5px 0 0; }
    .links { display: flex; gap: 10px; flex-wrap: wrap; }
    a { color: #dbeafe; text-decoration: none; border: 1px solid #4b6384; border-radius: 7px; padding: 8px 10px; font-size: 14px; }
    a:hover { background: #243550; }
    .metrics { display: grid; grid-template-columns: repeat(4, minmax(150px, 1fr)); gap: 12px; }
    .metric, .chart { background: #172235; border: 1px solid #2c3e59; border-radius: 10px; }
    .metric { padding: 16px; }
    .metric label { color: #9fb0c4; display: block; font-size: 13px; }
    .value { font-size: 28px; font-variant-numeric: tabular-nums; font-weight: 650; margin-top: 6px; }
    .charts { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; margin-top: 14px; }
    .simulator { background: #172235; border: 1px solid #2c3e59; border-radius: 10px; padding: 16px; margin-top: 14px; display: flex; align-items: end; gap: 12px; flex-wrap: wrap; }
    .simulator h2 { width: 100%; font-size: 15px; margin: 0; }
    .simulator p { width: 100%; margin: -4px 0 0; color: #9fb0c4; font-size: 13px; }
    .simulator label { color: #9fb0c4; font-size: 13px; display: grid; gap: 5px; }
    input { background: #101827; border: 1px solid #4b6384; border-radius: 6px; color: #e5edf7; font: inherit; padding: 8px; width: 135px; }
    button { background: #2563eb; border: 0; border-radius: 7px; color: white; cursor: pointer; font: inherit; font-weight: 600; padding: 9px 14px; }
    button:disabled { cursor: wait; opacity: .65; }
    .simulation-result { color: #dbeafe; font-size: 14px; min-height: 20px; }
    .chart { padding: 16px; min-width: 0; }
    .chart h2 { font-size: 15px; margin: 0 0 8px; font-weight: 600; }
    svg { width: 100%; height: 180px; display: block; overflow: visible; }
    .axis { stroke: #41516a; stroke-width: 1; }
    .line { fill: none; stroke-width: 3; stroke-linejoin: round; stroke-linecap: round; }
    .legend { color: #9fb0c4; font-size: 12px; display: flex; justify-content: space-between; }
    .status { margin-top: 16px; color: #9fb0c4; font-size: 13px; }
    @media (max-width: 720px) { body { padding: 16px; } header { display: block; } .links { margin-top: 14px; } .metrics, .charts { grid-template-columns: 1fr 1fr; } }
    @media (max-width: 440px) { .metrics, .charts { grid-template-columns: 1fr; } }
  </style>
</head>
<body>
  <main>
    <header>
      <div><h1>TCP server load dashboard</h1><p class="subtle">Rolling 60-second view · updates every second</p></div>
      <nav class="links" aria-label="Diagnostics"><a href="/debug/pprof/">pprof index</a><a href="/debug/pprof/goroutine?debug=1">goroutine stack dump</a></nav>
    </header>
    <section class="metrics" aria-label="Current server metrics">
      <div class="metric"><label>Requests / second</label><div class="value" id="rps">0</div></div>
      <div class="metric"><label>Active connections</label><div class="value" id="active">0</div></div>
      <div class="metric"><label>Average handler latency</label><div class="value" id="latency">0 ms</div></div>
      <div class="metric"><label>Goroutines</label><div class="value" id="goroutines">0</div></div>
      <div class="metric"><label>Total requests</label><div class="value" id="requests">0</div></div>
      <div class="metric"><label>Failed requests</label><div class="value" id="failed">0</div></div>
      <div class="metric"><label>Heap allocation</label><div class="value" id="heap">0 MB</div></div>
      <div class="metric"><label>Garbage collections</label><div class="value" id="gc">0</div></div>
    </section>
    <section class="simulator" aria-label="In-dashboard load simulation">
      <h2>Run a local load simulation</h2>
      <p>Traffic is generated on this computer against the raw TCP server. Start small, then increase gradually.</p>
      <label>Requests<input id="load-requests" type="number" min="1" max="10000" value="500"></label>
      <label>Concurrent requests<input id="load-concurrency" type="number" min="1" max="500" value="50"></label>
      <button id="run-load" type="button">Run simulation</button>
      <div class="simulation-result" id="simulation-result" aria-live="polite"></div>
    </section>
    <section class="charts" aria-label="Metric charts">
      <article class="chart"><h2>Request throughput</h2><svg id="throughput" viewBox="0 0 520 180" role="img" aria-label="Requests per second over time"></svg><div class="legend"><span>60 seconds ago</span><span id="throughput-max">0 req/s</span><span>now</span></div></article>
      <article class="chart"><h2>Connections and goroutines</h2><svg id="concurrency" viewBox="0 0 520 180" role="img" aria-label="Active connections and goroutines over time"></svg><div class="legend"><span>Connections: blue</span><span>Goroutines: green</span></div></article>
      <article class="chart"><h2>Average handler latency</h2><svg id="latency-chart" viewBox="0 0 520 180" role="img" aria-label="Average handler latency over time"></svg><div class="legend"><span>60 seconds ago</span><span id="latency-max">0 ms</span><span>now</span></div></article>
      <article class="chart"><h2>Heap allocation</h2><svg id="heap-chart" viewBox="0 0 520 180" role="img" aria-label="Heap allocation over time"></svg><div class="legend"><span>60 seconds ago</span><span id="heap-max">0 MB</span><span>now</span></div></article>
    </section>
    <p class="status" id="status" aria-live="polite">Connecting to local metrics…</p>
  </main>
  <script>
    const history = [];
    const maximumSamples = 60;
    const $ = id => document.getElementById(id);
    const number = value => new Intl.NumberFormat().format(value);
    const megabytes = value => (value / 1024 / 1024).toFixed(1) + ' MB';

    function values(key) { return history.map(point => point[key]); }
    function pathFor(series, max) {
      const width = 520, height = 180, pad = 12;
      return series.map((value, index) => {
        const x = pad + (width - pad * 2) * (series.length <= 1 ? 1 : index / (series.length - 1));
        const y = height - pad - (height - pad * 2) * (value / max);
        return (index ? 'L' : 'M') + x.toFixed(1) + ' ' + y.toFixed(1);
      }).join(' ');
    }
    function draw(id, seriesList) {
      const svg = $(id);
      const all = seriesList.flatMap(series => series.values);
      const max = Math.max(1, ...all);
      const lines = seriesList.map(series => '<path class="line" stroke="' + series.color + '" d="' + pathFor(series.values, max) + '"/>').join('');
      svg.innerHTML = '<line class="axis" x1="12" y1="168" x2="508" y2="168"/><line class="axis" x1="12" y1="12" x2="12" y2="168"/>' + lines;
      return max;
    }
    function show(metrics) {
      const previous = history.at(-1);
      const rate = previous ? Math.max(0, metrics.totalRequests - previous.totalRequests) : 0;
      const point = { ...metrics, rate, heapMB: metrics.heapAllocBytes / 1024 / 1024 };
      history.push(point);
      if (history.length > maximumSamples) history.shift();
      $('rps').textContent = rate.toFixed(0);
      $('active').textContent = number(metrics.activeConnections);
      $('latency').textContent = metrics.averageLatencyMs.toFixed(2) + ' ms';
      $('goroutines').textContent = number(metrics.goroutines);
      $('requests').textContent = number(metrics.totalRequests);
      $('failed').textContent = number(metrics.failedRequests);
      $('heap').textContent = megabytes(metrics.heapAllocBytes);
      $('gc').textContent = number(metrics.gcCount);
      $('throughput-max').textContent = draw('throughput', [{ values: values('rate'), color: '#60a5fa' }]).toFixed(0) + ' req/s';
      draw('concurrency', [{ values: values('activeConnections'), color: '#60a5fa' }, { values: values('goroutines'), color: '#4ade80' }]);
      $('latency-max').textContent = draw('latency-chart', [{ values: values('averageLatencyMs'), color: '#fbbf24' }]).toFixed(2) + ' ms';
      $('heap-max').textContent = draw('heap-chart', [{ values: values('heapMB'), color: '#c084fc' }]).toFixed(1) + ' MB';
      $('status').textContent = 'Last updated ' + new Date(metrics.capturedAt).toLocaleTimeString();
    }
    async function runSimulation() {
      const button = $('run-load');
      const requests = Number($('load-requests').value);
      const concurrency = Number($('load-concurrency').value);
      button.disabled = true;
      $('simulation-result').textContent = 'Running ' + requests + ' requests at ' + concurrency + ' concurrent connections…';
      try {
        const response = await fetch('/debug/load', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ requests, concurrency })
        });
        if (!response.ok) throw new Error(await response.text());
        const result = await response.json();
        $('simulation-result').textContent = result.completed + '/' + result.requested + ' completed · ' + result.failed + ' failed · ' + result.requestsPerSecond.toFixed(1) + ' req/s · ' + result.averageLatencyMs.toFixed(2) + ' ms end-to-end average';
      } catch (error) {
        $('simulation-result').textContent = 'Simulation failed: ' + error.message;
      } finally {
        button.disabled = false;
      }
    }
    async function update() {
      try {
        const response = await fetch('/debug/metrics', { cache: 'no-store' });
        if (!response.ok) throw new Error('HTTP ' + response.status);
        show(await response.json());
      } catch (error) {
        $('status').textContent = 'Metrics unavailable: ' + error.message;
      }
    }
    update();
    setInterval(update, 1000);
    $('run-load').addEventListener('click', runSimulation);
  </script>
</body>
</html>`
