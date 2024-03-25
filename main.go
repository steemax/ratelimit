package traefik_ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func CreateConfig() *Config {
	return &Config{}
}

type Config struct {
	Rate int `json:"rate,omitempty"`
}

type RateLimit struct {
	name   string
	next   http.Handler
	rate   int
	config *Config
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	mlog(fmt.Sprintf("config %v", config))
	return &RateLimit{
		name:   name,
		next:   next,
		config: config,
	}, nil
}

func (r *RateLimit) Allow(ctx context.Context, req *http.Request, rw http.ResponseWriter) bool {
	return true
}

func (r *RateLimit) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	encoder := json.NewEncoder(rw)
	reqCtx := req.Context()
	if r.Allow(reqCtx, req, rw) {
		r.next.ServeHTTP(rw, req)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusTooManyRequests)
	if err := encoder.Encode(map[string]interface{}{"status_code": http.StatusTooManyRequests, "message": "rate limit exceeded, try again later"}); err != nil {
		mlog("Ошибка кодирования данных:", err)
		return
	}
	return
}

func mlog(args ...interface{}) {
	os.Stdout.WriteString(fmt.Sprintf("[rate-limit-middleware-plugin] %s\n", fmt.Sprint(args...)))
}
