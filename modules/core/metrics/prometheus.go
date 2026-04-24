package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchup_http_requests_total",
			Help: "Total HTTP requests by method, route, and status code",
		},
		[]string{"method", "route", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "matchup_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method", "route", "status"},
	)

	// Business metrics — feed
	SwipeEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchup_swipe_events_total",
			Help: "Total swipe events by action (LIKE/PASS) and recommendation source tier",
		},
		[]string{"action", "source"},
	)

	MatchEventsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "matchup_match_events_total",
			Help: "Total mutual matches created",
		},
	)

	RecommendationTierHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchup_recommendation_tier_hits_total",
			Help: "Feed candidates served by recommendation tier",
		},
		[]string{"tier"},
	)

	RecommendationTierErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchup_recommendation_tier_errors_total",
			Help: "Errors returned by each recommendation tier",
		},
		[]string{"tier"},
	)

	RecommendationTierEmpty = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchup_recommendation_tier_empty_total",
			Help: "Times a recommendation tier returned zero candidates",
		},
		[]string{"tier"},
	)

	// Database connection pool metrics
	DBPoolAcquiredConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "matchup_db_pool_acquired_conns",
		Help: "Number of currently acquired pgxpool connections",
	})

	DBPoolIdleConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "matchup_db_pool_idle_conns",
		Help: "Number of currently idle pgxpool connections",
	})

	DBPoolTotalConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "matchup_db_pool_total_conns",
		Help: "Total pgxpool connections (acquired + idle + constructing)",
	})

	// Cron job metrics
	CronJobSuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchup_cron_job_success_total",
			Help: "Total successful cron job executions by job name",
		},
		[]string{"job"},
	)

	CronJobFailureTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matchup_cron_job_failure_total",
			Help: "Total failed cron job executions by job name",
		},
		[]string{"job"},
	)
)
