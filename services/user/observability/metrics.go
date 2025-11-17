package metrics

import "github.com/prometheus/client_golang/prometheus"

type UserMetrics struct {
	UsersCreated        prometheus.Counter
	LoginSuccess        prometheus.Counter
	LoginErrors         prometheus.Counter
	LoginDuration       prometheus.Histogram
	SessionRenew        prometheus.Counter
	SessionRenewLatency prometheus.Histogram
}

func NewUserMetrics(reg prometheus.Registerer) *UserMetrics {
	m := &UserMetrics{
		UsersCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "user_service",
			Name:      "users_created_total",
			Help:      "Total number of users created.",
		}),

		LoginSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "user_service",
			Name:      "logins_successful_total",
			Help:      "Total number of successful logins.",
		}),

		LoginErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "user_service",
			Name:      "login_errors_total",
			Help:      "Total number of login errors.",
		}),

		LoginDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "user_service",
			Name:      "login_duration_seconds",
			Help:      "Duration of user login process.",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10),
		}),

		SessionRenew: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "user_service",
			Name:      "session_renew_total",
			Help:      "Number of session renew operations.",
		}),

		SessionRenewLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "user_service",
			Name:      "session_renew_duration_seconds",
			Help:      "Duration of session renewal process.",
			Buckets:   prometheus.ExponentialBuckets(0.005, 2, 10),
		}),
	}

	reg.MustRegister(m.UsersCreated)
	reg.MustRegister(m.LoginSuccess)
	reg.MustRegister(m.LoginErrors)
	reg.MustRegister(m.LoginDuration)
	reg.MustRegister(m.SessionRenew)
	reg.MustRegister(m.SessionRenewLatency)

	return m
}

// Métodos CORRETOS agora
func (m *UserMetrics) IncUsersCreated() {
	m.UsersCreated.Inc()
}

func (m *UserMetrics) IncLoginSuccess() {
	m.LoginSuccess.Inc()
}

func (m *UserMetrics) IncLoginError() {
	m.LoginErrors.Inc()
}

func (m *UserMetrics) ObserveLoginDuration(seconds float64) {
	m.LoginDuration.Observe(seconds)
}

func (m *UserMetrics) IncSessionRenew() {
	m.SessionRenew.Inc()
}

func (m *UserMetrics) ObserveSessionRenewDuration(seconds float64) {
	m.SessionRenewLatency.Observe(seconds)
}
