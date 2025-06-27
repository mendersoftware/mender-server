package rate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
)

var ErrTooManyRequests = errors.New("too many requests")

type limiterGroup struct {
	limiter       EventLimiter
	eventTemplate *template.Template
}

type HTTPLimiter struct {
	templateData  func(r *http.Request) any
	template      *template.Template
	mux           *http.ServeMux
	DefaultGroup  *limiterGroup
	limiterGroups map[string]*limiterGroup

	rewriteRequests bool
}

func defaultTemplateData(r *http.Request) any {
	id := identity.FromContext(r.Context())
	ctx := map[string]any{
		"Identity": id,
		// "Request":  r, // not needed, but could be handy
	}
	return ctx
}

func NewHTTPLimiter(eventLimiter EventLimiter, eventTemplate string) (*HTTPLimiter, error) {
	template, err := template.New("").Parse(eventTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid eventTemplate: %w", err)
	}
	return &HTTPLimiter{
		template:     template.New("").Option("missingkey=zero"),
		mux:          http.NewServeMux(),
		templateData: defaultTemplateData,
		DefaultGroup: &limiterGroup{
			limiter:       eventLimiter,
			eventTemplate: template,
		},
		limiterGroups: make(map[string]*limiterGroup),
	}, nil
}

func (h *HTTPLimiter) WithTemplateDataFunc(f func(*http.Request) any) *HTTPLimiter {
	h.templateData = f
	return h
}

func (h *HTTPLimiter) WithTemplateFuncs(funcs map[string]any) *HTTPLimiter {
	h.template.Funcs(funcs)
	return h
}

func (h *HTTPLimiter) WithRewriteRequests(rewrite bool) *HTTPLimiter {
	h.rewriteRequests = rewrite
	return h
}

func (h *HTTPLimiter) AddRateLimitGroup(limiter EventLimiter, group, eventTemplate string) error {
	t, err := h.template.Clone()
	if err == nil {
		_, err = t.Parse(eventTemplate)
	}
	if err != nil {
		return fmt.Errorf("failed to compile event template: %w", err)
	}
	h.limiterGroups[group] = &limiterGroup{
		limiter:       limiter,
		eventTemplate: t,
	}
	return nil
}

func (h *HTTPLimiter) MatchHTTPPattern(
	pattern, groupTemplate string,
) error {
	var (
		t   *template.Template
		err error
	)
	if groupTemplate != "" {
		// Compile eventTemplate:
		t, err = h.template.Clone()
		if err == nil {
			_, err = t.Parse(groupTemplate)
		}
		if err != nil {
			return fmt.Errorf("error parsing group_template: %w", err)
		}
	}
	limiterHandle := handle{
		HTTPLimiter:   h,
		groupTemplate: t,
	}
	h.mux.Handle(pattern, limiterHandle)
	return nil
}

type handle struct {
	*HTTPLimiter
	groupTemplate *template.Template
}

// ServeHTTP implements a basic http.Handler so that handler can be used
// as a handler for the mux. It will only write on errors and is expected
// to continue to the actual handler on success.
func (h *HTTPLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if h.rewriteRequests {
		r = rest.RewriteForwardedRequest(r)
	}
	res, err := h.Reserve(r)
	if err != nil {
		hdr := w.Header()
		hdr.Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(rest.Error{
			Err:       "internal server error",
			RequestID: requestid.FromContext(ctx),
		})
		return
	}
	if res == nil || res.OK() {
		return
	} else {
		hdr := w.Header()
		hdr.Set("Content-Type", "application/json")
		retryAfter := math.Ceil(res.Delay().Abs().Seconds())
		hdr.Set("Retry-After", fmt.Sprint(retryAfter))
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(rest.Error{
			Err:       ErrTooManyRequests.Error(),
			RequestID: requestid.FromContext(ctx),
		})
	}
}

func (h *HTTPLimiter) MiddlewareGin(c *gin.Context) {
	r := c.Request
	if h.rewriteRequests {
		r = rest.RewriteForwardedRequest(r)
	}
	res, err := h.Reserve(r)
	if err != nil {
		rest.RenderInternalError(c, err)
		c.Abort()
	}
	if res == nil || res.OK() {
		c.Next()
	} else {
		retryAfter := math.Ceil(res.Delay().Abs().Seconds())
		c.Header("Retry-After", fmt.Sprint(retryAfter))
		rest.RenderError(c, http.StatusTooManyRequests, ErrTooManyRequests)
		c.Abort()
	}
}

func (m *HTTPLimiter) Reserve(r *http.Request) (Reservation, error) {
	var b bytes.Buffer
	limiter := m.DefaultGroup
	templateData := m.templateData(r)
	ctx := r.Context()
	h, _ := m.mux.Handler(r)
	hh, ok := h.(handle)
	if ok {
		if hh.groupTemplate != nil {
			err := hh.groupTemplate.Execute(&b, templateData)
			if err != nil {
				return nil, fmt.Errorf("error executing ratelimit group template: %w", err)
			}
			group := m.limiterGroups[b.String()]
			if group != nil {
				limiter = group
			}
			b.Reset()
		}
	}
	err := limiter.eventTemplate.Execute(&b, templateData)
	if err != nil {
		return nil, fmt.Errorf("error executing template for event ID: %w", err)
	}
	if limiter == nil {
		return nil, nil
	}
	return limiter.limiter.ReserveEvent(ctx, b.String())
}
