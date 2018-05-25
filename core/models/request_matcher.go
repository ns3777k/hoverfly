package models

import (
	"net/url"

	"github.com/SpectoLabs/hoverfly/core/handlers/v2"
	"github.com/SpectoLabs/hoverfly/core/matching/matchers"
	"github.com/SpectoLabs/hoverfly/core/util"
)

type RequestFieldMatchers struct {
	Matcher string
	Value   interface{}
}

func NewRequestFieldMatchersFromView(matchers []v2.MatcherViewV5) []RequestFieldMatchers {
	if matchers == nil {
		return nil
	}
	convertedMatchers := []RequestFieldMatchers{}
	for _, matcher := range matchers {
		convertedMatchers = append(convertedMatchers, RequestFieldMatchers{
			Matcher: matcher.Matcher,
			Value:   matcher.Value,
		})
	}
	return convertedMatchers
}

func NewRequestFieldMatchersFromMapView(mapMatchers map[string][]v2.MatcherViewV5) map[string][]RequestFieldMatchers {
	var matchers map[string][]RequestFieldMatchers
	for key, view := range mapMatchers {
		if matchers == nil {
			matchers = map[string][]RequestFieldMatchers{}
		}
		matchers[key] = NewRequestFieldMatchersFromView(view)
	}
	return matchers
}

func (this RequestFieldMatchers) BuildView() v2.MatcherViewV5 {
	return v2.MatcherViewV5{
		Matcher: this.Matcher,
		Value:   this.Value,
	}
}

type RequestMatcherResponsePair struct {
	RequestMatcher RequestMatcher
	Response       ResponseDetails
}

func NewRequestMatcherResponsePairFromView(view *v2.RequestMatcherResponsePairViewV5) *RequestMatcherResponsePair {
	for i, matcher := range view.RequestMatcher.Query {
		if matcher.Matcher == matchers.Exact {
			sortedQuery := util.SortQueryString(matcher.Value.(string))
			view.RequestMatcher.Query[i].Value = sortedQuery
		}
	}

	return &RequestMatcherResponsePair{
		RequestMatcher: RequestMatcher{
			Path:                NewRequestFieldMatchersFromView(view.RequestMatcher.Path),
			Method:              NewRequestFieldMatchersFromView(view.RequestMatcher.Method),
			Destination:         NewRequestFieldMatchersFromView(view.RequestMatcher.Destination),
			Scheme:              NewRequestFieldMatchersFromView(view.RequestMatcher.Scheme),
			Query:               NewRequestFieldMatchersFromView(view.RequestMatcher.Query),
			Body:                NewRequestFieldMatchersFromView(view.RequestMatcher.Body),
			Headers:             NewRequestFieldMatchersFromMapView(view.RequestMatcher.Headers),
			QueriesWithMatchers: NewRequestFieldMatchersFromMapView(view.RequestMatcher.QueriesWithMatchers),
			RequiresState:       view.RequestMatcher.RequiresState,
		},
		Response: NewResponseDetailsFromResponse(view.Response),
	}
}

func (this *RequestMatcherResponsePair) BuildView() v2.RequestMatcherResponsePairViewV5 {

	var path, method, destination, scheme, query, body []v2.MatcherViewV5

	if this.RequestMatcher.Path != nil && len(this.RequestMatcher.Path) != 0 {
		views := []v2.MatcherViewV5{}
		for _, matcher := range this.RequestMatcher.Path {
			views = append(views, matcher.BuildView())
		}
		path = views
	}

	if this.RequestMatcher.Method != nil && len(this.RequestMatcher.Method) != 0 {
		views := []v2.MatcherViewV5{}
		for _, matcher := range this.RequestMatcher.Method {
			views = append(views, matcher.BuildView())
		}
		method = views
	}

	if this.RequestMatcher.Destination != nil && len(this.RequestMatcher.Destination) != 0 {
		views := []v2.MatcherViewV5{}
		for _, matcher := range this.RequestMatcher.Destination {
			views = append(views, matcher.BuildView())
		}
		destination = views
	}

	if this.RequestMatcher.Scheme != nil && len(this.RequestMatcher.Scheme) != 0 {
		views := []v2.MatcherViewV5{}
		for _, matcher := range this.RequestMatcher.Scheme {
			views = append(views, matcher.BuildView())
		}
		scheme = views
	}

	if this.RequestMatcher.Body != nil && len(this.RequestMatcher.Body) != 0 {
		views := []v2.MatcherViewV5{}
		for _, matcher := range this.RequestMatcher.Body {
			views = append(views, matcher.BuildView())
		}
		body = views
	}

	if this.RequestMatcher.Query != nil && len(this.RequestMatcher.Query) != 0 {
		views := []v2.MatcherViewV5{}
		for _, matcher := range this.RequestMatcher.Query {
			views = append(views, matcher.BuildView())
		}
		query = views
	}

	headersWithMatchers := map[string][]v2.MatcherViewV5{}
	for key, matchers := range this.RequestMatcher.Headers {
		views := []v2.MatcherViewV5{}
		for _, matcher := range matchers {
			views = append(views, matcher.BuildView())
		}
		headersWithMatchers[key] = views
	}

	queriesWithMatchers := map[string][]v2.MatcherViewV5{}
	for key, matchers := range this.RequestMatcher.QueriesWithMatchers {
		views := []v2.MatcherViewV5{}
		for _, matcher := range matchers {
			views = append(views, matcher.BuildView())
		}
		queriesWithMatchers[key] = views
	}

	return v2.RequestMatcherResponsePairViewV5{
		RequestMatcher: v2.RequestMatcherViewV5{
			Path:                path,
			Method:              method,
			Destination:         destination,
			Scheme:              scheme,
			Query:               query,
			Body:                body,
			Headers:             headersWithMatchers,
			QueriesWithMatchers: queriesWithMatchers,
			RequiresState:       this.RequestMatcher.RequiresState,
		},
		Response: this.Response.ConvertToResponseDetailsViewV5(),
	}
}

type RequestMatcher struct {
	Path                []RequestFieldMatchers
	Method              []RequestFieldMatchers
	Destination         []RequestFieldMatchers
	Scheme              []RequestFieldMatchers
	Query               []RequestFieldMatchers
	Body                []RequestFieldMatchers
	Headers             map[string][]RequestFieldMatchers
	QueriesWithMatchers map[string][]RequestFieldMatchers
	RequiresState       map[string]string
}

func (this RequestMatcher) IncludesHeaderMatching() bool {
	return (this.Headers != nil && len(this.Headers) > 0)
}

func (this RequestMatcher) IncludesStateMatching() bool {
	return this.RequiresState != nil && len(this.RequiresState) > 0
}

func (this RequestMatcher) ToEagerlyCachable() *RequestDetails {
	if this.Body == nil || len(this.Body) != 1 || this.Body[0].Matcher != matchers.Exact ||
		this.Destination == nil || len(this.Destination) != 1 || this.Destination[0].Matcher != matchers.Exact ||
		this.Method == nil || len(this.Method) != 1 || this.Method[0].Matcher != matchers.Exact ||
		this.Path == nil || len(this.Path) != 1 || this.Path[0].Matcher != matchers.Exact ||
		this.Query == nil || len(this.Query) != 1 || this.Query[0].Matcher != matchers.Exact ||
		this.Scheme == nil || len(this.Scheme) != 1 || this.Scheme[0].Matcher != matchers.Exact {
		return nil
	}

	if this.Headers != nil && len(this.Headers) > 0 {
		return nil
	}

	if this.RequiresState != nil && len(this.RequiresState) > 0 {
		return nil
	}

	query, _ := url.ParseQuery(this.Query[0].Value.(string))

	return &RequestDetails{
		Body:        this.Body[0].Value.(string),
		Destination: this.Destination[0].Value.(string),
		Method:      this.Method[0].Value.(string),
		Path:        this.Path[0].Value.(string),
		Query:       query,
		Scheme:      this.Scheme[0].Value.(string),
	}
}

type MatchError struct {
	ClosestMiss *ClosestMiss
	error       string
}

func NewMatchErrorWithClosestMiss(closestMiss *ClosestMiss, error string, isCachable bool) *MatchError {
	return &MatchError{
		ClosestMiss: closestMiss,
		error:       error,
	}
}

func NewMatchError(error string, matchedOnAllButHeadersAtLeastOnce bool) *MatchError {
	return &MatchError{
		error: error,
	}
}

func (this MatchError) Error() string {
	return this.error
}
