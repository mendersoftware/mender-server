package rate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

// TooManyRequestsError is the error type returned when hitting the rate
// limits.
type TooManyRequestsError struct {
	Delay time.Duration
}

func (err *TooManyRequestsError) Error() string {
	return "too many requests"
}

type templateEventLimiter struct {
	limiter       EventLimiter
	eventTemplate *template.Template
}

// HTTPLimiter combines a set of EventLimiter (groups) and a http.ServeMux
// for routing API endpoints to a EventLimiter.
//
// Every EventLimiter comes with an eventTemplate for grouping events by
// a Go template expression. The data to the Go template can be customized by
// calling WithTemplateDataFunc and WithTemplateFuncs.
//
// Additional routes are added by calling MatchHTTPPattern which accepts a HTTP
// pattern and a target group which also accepts Go template expressions.
//
// The HTTPLimiter implements both the standard http.Handler and
// gin.HandlerFunc (see (*HTTPLimiter).MiddlewareGin) for use with either the
// standard library or as a middleware for the gin-gonic framework.
type HTTPLimiter struct {
	// rootTemplate stores the root Template that is inherited by all
	// template string parameters (event and group strings).
	rootTemplate *template.Template
	// templateDataCallback is a callback that is called before executing
	// template strings and should output the template data used.
	//
	// See defaultTemplateData.
	templateDataCallback func(r *http.Request) any

	// httpMux implements the request router to a ratelimiter match instance
	httpMux *http.ServeMux

	// defaultLimiter is the default rate limiter if no group matches.
	defaultLimiter *templateEventLimiter
	// limiterGroups maps the group name to rate limiters.
	limiterGroups map[string]*templateEventLimiter

	// rewriteRequests decides whether to autmatically rewrite the request
	// URL using X-Forwarded-* headers.
	rewriteRequests bool
}

// defaultTemplateData is the default callback for executing template strings.
func defaultTemplateData(r *http.Request) any {
	id := identity.FromContext(r.Context())
	ctx := map[string]any{
		"Identity": id,
		// "Request":  r, // not needed, but could be handy
	}
	return ctx
}

// NewHTTPLimiter initializes an empty HTTPLimiter.
// The limiter is built by calling AddRateLimitGroup and AddMatchExpression.
func NewHTTPLimiter() *HTTPLimiter {
	return &HTTPLimiter{
		rootTemplate:         template.New("").Option("missingkey=zero"),
		httpMux:              http.NewServeMux(),
		templateDataCallback: defaultTemplateData,
		limiterGroups:        make(map[string]*templateEventLimiter),
	}
}

func (h *HTTPLimiter) WithTemplateDataFunc(f func(*http.Request) any) *HTTPLimiter {
	h.templateDataCallback = f
	return h
}

func (h *HTTPLimiter) WithTemplateFuncs(funcs map[string]any) *HTTPLimiter {
	h.rootTemplate.Funcs(funcs)
	return h
}

func (h *HTTPLimiter) WithRewriteRequests(rewrite bool) *HTTPLimiter {
	h.rewriteRequests = rewrite
	return h
}

func (h *HTTPLimiter) AddRateLimitGroup(limiter EventLimiter, group, eventTemplate string) error {
	t, err := h.rootTemplate.Clone()
	if err == nil {
		_, err = t.Parse(eventTemplate)
	}
	if err != nil {
		return fmt.Errorf("failed to compile event template: %w", err)
	}
	h.limiterGroups[group] = &templateEventLimiter{
		limiter:       limiter,
		eventTemplate: t,
	}
	return nil
}

// AddMatchExpression creates a new route using pattern to apply the
// rate limiter that is matched by groupTemplate Go Template.
func (h *HTTPLimiter) AddMatchExpression(
	pattern, groupTemplate string,
) error {
	var (
		t   *template.Template
		err error
	)
	// Compile eventTemplate:
	t, err = h.rootTemplate.Clone()
	if err == nil {
		_, err = t.Parse(groupTemplate)
	}
	if err != nil {
		return fmt.Errorf("error parsing group_template: %w", err)
	}
	limiterMatcher := matcher{
		HTTPLimiter:   h,
		groupTemplate: t,
	}
	h.httpMux.Handle(pattern, limiterMatcher)
	return nil
}

// matcher is the HTTPHandle
type matcher struct {
	*HTTPLimiter
	groupTemplate *template.Template
}

func (h *HTTPLimiter) handleRequest(r *http.Request) error {
	if h.rewriteRequests {
		r = rest.RewriteForwardedRequest(r)
	}
	res, err := h.Reserve(r)
	if err != nil {
		return err
	}
	if res == nil || res.OK() {
		return nil
	} else {
		return &TooManyRequestsError{
			Delay: res.Delay(),
		}
	}
}

func handleError(ctx context.Context, w http.ResponseWriter, err error) {
	var tooManyRequests *TooManyRequestsError
	status := http.StatusInternalServerError
	hdr := w.Header()
	hdr.Set("Content-Type", "application/json")
	if errors.As(err, &tooManyRequests) {
		status = http.StatusTooManyRequests
		retryAfter := int64(math.Ceil(tooManyRequests.Delay.Abs().Seconds()))
		hdr.Set("Retry-After", strconv.FormatInt(retryAfter, 10))
	}
	w.WriteHeader(status)
	b, _ := json.Marshal(rest.Error{
		Err:       err.Error(),
		RequestID: requestid.FromContext(ctx),
	})
	_, _ = w.Write(b)
}

// ServeHTTP implements a basic http.Handler so that handler can be used
// as a handler for the mux. It will only write on errors and is expected
// to continue to the actual handler on success.
func (h *HTTPLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.handleRequest(r)
	if err != nil {
		handleError(r.Context(), w, err)
	}
}

// MiddlewareGin implements rate limiting as a middleware for gin-gonic
// web framework.
func (h *HTTPLimiter) MiddlewareGin(c *gin.Context) {
	err := h.handleRequest(c.Request)
	if err != nil {
		_ = c.Error(err)
		handleError(c.Request.Context(), c.Writer, err)
		c.Abort()
	}
}

type okReservation struct{}

func (k okReservation) OK() bool             { return true }
func (k okReservation) Delay() time.Duration { return 0 }
func (k okReservation) Tokens() int64        { return math.MaxInt64 }

func (m *HTTPLimiter) Reserve(r *http.Request) (Reservation, error) {
	var b bytes.Buffer
	var eventLimiter *templateEventLimiter
	templateData := m.templateDataCallback(r)
	ctx := r.Context()
	h, _ := m.httpMux.Handler(r)
	hh, ok := h.(matcher)
	if ok && hh.groupTemplate != nil {
		err := hh.groupTemplate.Execute(&b, templateData)
		if err != nil {
			return nil, fmt.Errorf("error executing ratelimit group template: %w", err)
		}
		eventLimiter = m.limiterGroups[b.String()]
		if eventLimiter == nil {
			return okReservation{}, nil
		}
		b.Reset()
	} else {
		// If no ratelimiter is found, skip rate limits.
		return okReservation{}, nil
	}
	err := eventLimiter.eventTemplate.Execute(&b, templateData)
	if err != nil {
		return nil, fmt.Errorf("error executing template for event ID: %w", err)
	}
	return eventLimiter.limiter.ReserveEvent(ctx, b.String())
}
