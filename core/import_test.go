package hoverfly

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/SpectoLabs/hoverfly/core/cache"
	"github.com/SpectoLabs/hoverfly/core/handlers/v2"
	"github.com/SpectoLabs/hoverfly/core/matching"
	"github.com/SpectoLabs/hoverfly/core/matching/matchers"
	"github.com/SpectoLabs/hoverfly/core/models"
	. "github.com/onsi/gomega"
)

const hoverflyIoSimulationPath = "../examples/simulations/hoverfly.io.json"

func TestIsURLHTTP(t *testing.T) {
	RegisterTestingT(t)

	url := "http://somehost.com"

	b := isURL(url)
	Expect(b).To(BeTrue())
}

func TestIsURLEmpty(t *testing.T) {
	RegisterTestingT(t)

	b := isURL("")
	Expect(b).To(BeFalse())
}

func TestIsURLHTTPS(t *testing.T) {
	RegisterTestingT(t)

	url := "https://somehost.com"

	b := isURL(url)
	Expect(b).To(BeTrue())
}

func TestIsURLWrong(t *testing.T) {
	RegisterTestingT(t)

	url := "somehost.com"

	b := isURL(url)
	Expect(b).To(BeFalse())
}

func TestIsURLWrongTLD(t *testing.T) {
	RegisterTestingT(t)

	url := "http://somehost."

	b := isURL(url)
	Expect(b).To(BeFalse())
}

func TestFileExists(t *testing.T) {
	RegisterTestingT(t)

	ex, err := exists(hoverflyIoSimulationPath)
	Expect(err).To(BeNil())
	Expect(ex).To(BeTrue())
}

func TestFileDoesNotExist(t *testing.T) {
	RegisterTestingT(t)

	fp := "shouldnotbehere.yaml"

	ex, err := exists(fp)
	Expect(err).To(BeNil())
	Expect(ex).To(BeFalse())
}

func TestImportFromDisk(t *testing.T) {
	RegisterTestingT(t)

	unit := NewHoverflyWithConfiguration(&Configuration{})

	err := unit.Import(hoverflyIoSimulationPath)
	Expect(err).To(BeNil())

	Expect(unit.Simulation.GetMatchingPairs()).To(HaveLen(2))
}

func TestImportFromDiskBlankPath(t *testing.T) {
	RegisterTestingT(t)

	unit := NewHoverflyWithConfiguration(&Configuration{})

	err := unit.ImportFromDisk("")
	Expect(err).ToNot(BeNil())
}

func TestImportFromDiskWrongJson(t *testing.T) {
	RegisterTestingT(t)

	unit := NewHoverflyWithConfiguration(&Configuration{})

	err := unit.ImportFromDisk("examples/exports/README.md")
	Expect(err).ToNot(BeNil())
}

func TestImportFromURL(t *testing.T) {
	RegisterTestingT(t)

	// reading file and preparing json payload
	pairFile, err := os.Open(hoverflyIoSimulationPath)
	Expect(err).To(BeNil())
	pairFileBytes, err := ioutil.ReadAll(pairFile)
	Expect(err).To(BeNil())

	// pretending this is the endpoint with given json
	server, unit := testTools(200, string(pairFileBytes))
	defer server.Close()

	// importing payloads
	err = unit.Import(server.URL)
	Expect(err).To(BeNil())

	Expect(unit.Simulation.GetMatchingPairs()).To(HaveLen(2))
}

func TestImportFromURLRedirect(t *testing.T) {
	RegisterTestingT(t)

	// reading file and preparing json payload
	pairFile, err := os.Open(hoverflyIoSimulationPath)
	Expect(err).To(BeNil())
	pairFileBytes, err := ioutil.ReadAll(pairFile)
	Expect(err).To(BeNil())

	// pretending this is the endpoint with given json
	server, unit := testTools(200, string(pairFileBytes))
	defer server.Close()

	unit.HTTP = GetDefaultHoverflyHTTPClient(false, "")

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", server.URL)
		w.WriteHeader(301)
	}))
	defer redirectServer.Close()

	// importing payloads
	err = unit.Import(redirectServer.URL)
	Expect(err).To(BeNil())

	Expect(unit.Simulation.GetMatchingPairs()).To(HaveLen(2))
}

func TestImportFromURLHTTPFail(t *testing.T) {
	RegisterTestingT(t)

	// this tests simulates unreachable server
	unit := NewHoverflyWithConfiguration(&Configuration{})

	err := unit.ImportFromURL("somepath")
	Expect(err).ToNot(BeNil())
}

func TestImportFromURLMalformedJSON(t *testing.T) {
	RegisterTestingT(t)

	// testing behaviour when there is no json on the other end
	unit := NewHoverflyWithConfiguration(&Configuration{})

	// importing payloads
	err := unit.Import("http://thiswillbeintercepted.json")
	// we should get error
	Expect(err).ToNot(BeNil())
}

func TestImportRequestResponsePairs_CanImportASinglePair(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	originalPair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        "hello_world",
			EncodedBody: false,
			Headers:     map[string][]string{"Content-Type": {"text/plain"}},
			Templated:   true,
		},
		RequestMatcher: v2.RequestMatcherViewV6{
			Path: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Method: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "GET",
				},
			},
			Destination: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Scheme: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "scheme",
				},
			},
			Body: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "",
				},
			},
			Headers: map[string][]v2.MatcherViewV6{
				"Hoverfly": {
					{
						Matcher: matchers.Exact,
						Value:   "testing",
					},
				},
			}}}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{originalPair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(0))

	Expect(hv.Simulation.GetMatchingPairs()[0]).To(Equal(models.RequestMatcherResponsePair{
		Response: models.ResponseDetails{
			Status:    200,
			Body:      "hello_world",
			Headers:   map[string][]string{"Content-Type": {"text/plain"}},
			Templated: true,
		},
		RequestMatcher: models.RequestMatcher{
			Path: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Method: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "GET",
				},
			},
			Destination: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Scheme: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "scheme",
				},
			},
			Body: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "",
				},
			},
			Headers: map[string][]models.RequestFieldMatchers{
				"Hoverfly": {
					{
						Matcher: matchers.Exact,
						Value:   "testing",
					},
				},
			},
		},
	}))
}

func TestImportImportRequestResponsePairs_CanImportAMultiplePairsAndSetTemplateExplicitlyOrExplicitly(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	originalPair1 := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        "hello_world",
			EncodedBody: false,
			Headers:     map[string][]string{"Hoverfly": {"testing"}},
		},
		RequestMatcher: v2.RequestMatcherViewV6{
			Path: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Method: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "GET",
				},
			},
			Destination: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Scheme: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "scheme",
				},
			},
			Body: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "",
				},
			},
			Headers: map[string][]v2.MatcherViewV6{
				"Hoverfly": {
					{
						Matcher: matchers.Exact,
						Value:   "testing",
					},
				},
			}}}

	originalPair2 := originalPair1
	originalPair2.Response.Templated = false
	originalPair2.RequestMatcher.Path = []v2.MatcherViewV6{
		{
			Matcher: matchers.Exact,
			Value:   "/new/path",
		},
	}

	originalPair3 := originalPair1
	originalPair3.RequestMatcher.Path = []v2.MatcherViewV6{
		{
			Matcher: matchers.Exact,
			Value:   "/newer/path",
		},
	}
	originalPair3.Response.Templated = true

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{originalPair1, originalPair2, originalPair3})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(0))

	Expect(hv.Simulation.GetMatchingPairs()).To(HaveLen(3))
	Expect(hv.Simulation.GetMatchingPairs()[0]).To(Equal(models.RequestMatcherResponsePair{
		Response: models.ResponseDetails{
			Status:    200,
			Body:      "hello_world",
			Headers:   map[string][]string{"Hoverfly": {"testing"}},
			Templated: false,
		},
		RequestMatcher: models.RequestMatcher{
			Path: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Method: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "GET",
				},
			},
			Destination: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Scheme: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "scheme",
				},
			},
			Body: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "",
				},
			},
			Headers: map[string][]models.RequestFieldMatchers{
				"Hoverfly": {
					{
						Matcher: matchers.Exact,
						Value:   "testing",
					},
				},
			},
		},
	}))

	Expect(hv.Simulation.GetMatchingPairs()[1]).To(Equal(models.RequestMatcherResponsePair{
		Response: models.ResponseDetails{
			Status:    200,
			Body:      "hello_world",
			Headers:   map[string][]string{"Hoverfly": {"testing"}},
			Templated: false,
		},
		RequestMatcher: models.RequestMatcher{
			Path: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "/new/path",
				},
			},
			Method: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "GET",
				},
			},
			Destination: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Scheme: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "scheme",
				},
			},
			Body: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "",
				},
			},
			Headers: map[string][]models.RequestFieldMatchers{
				"Hoverfly": {
					{
						Matcher: matchers.Exact,
						Value:   "testing",
					},
				},
			},
		},
	}))

	Expect(hv.Simulation.GetMatchingPairs()[2]).To(Equal(models.RequestMatcherResponsePair{
		Response: models.ResponseDetails{
			Status:    200,
			Body:      "hello_world",
			Headers:   map[string][]string{"Hoverfly": {"testing"}},
			Templated: true,
		},
		RequestMatcher: models.RequestMatcher{
			Path: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "/newer/path",
				},
			},
			Method: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "GET",
				},
			},
			Destination: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Scheme: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "scheme",
				},
			},
			Body: []models.RequestFieldMatchers{
				{
					Matcher: matchers.Exact,
					Value:   "",
				},
			},
			Headers: map[string][]models.RequestFieldMatchers{
				"Hoverfly": {
					{
						Matcher: matchers.Exact,
						Value:   "testing",
					},
				},
			},
		},
	}))
}

func TestImportImportRequestResponsePairs_CanImportARequestResponsePairView(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	request := v2.RequestMatcherViewV6{
		Method: []v2.MatcherViewV6{
			{
				Matcher: matchers.Exact,
				Value:   "GET",
			},
		},
	}

	responseView := v2.ResponseDetailsViewV6{
		Status:      200,
		Body:        "hello_world",
		EncodedBody: false,
		Headers:     map[string][]string{"Hoverfly": {"testing"}},
	}

	requestResponsePair := v2.RequestMatcherResponsePairViewV6{
		Response:       responseView,
		RequestMatcher: request,
	}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{requestResponsePair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(0))

	Expect(len(hv.Simulation.GetMatchingPairs())).To(Equal(1))

	Expect(hv.Simulation.GetMatchingPairs()[0].RequestMatcher.Method).To(HaveLen(1))
	Expect(hv.Simulation.GetMatchingPairs()[0].RequestMatcher.Method[0].Matcher).To(Equal("exact"))
	Expect(hv.Simulation.GetMatchingPairs()[0].RequestMatcher.Method[0].Value).To(Equal("GET"))

	Expect(hv.Simulation.GetMatchingPairs()[0].Response.Status).To(Equal(200))
	Expect(hv.Simulation.GetMatchingPairs()[0].Response.Body).To(Equal("hello_world"))
	Expect(hv.Simulation.GetMatchingPairs()[0].Response.Headers).To(Equal(map[string][]string{"Hoverfly": {"testing"}}))
}

// Helper function for base64 encoding
func base64String(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func TestImportImportRequestResponsePairs_CanImportASingleBase64EncodedPair(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	encodedPair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        base64String("hello_world"),
			EncodedBody: true,
			Headers:     map[string][]string{"Content-Encoding": {"gzip"}}},
		RequestMatcher: v2.RequestMatcherViewV6{
			Path: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Method: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "GET",
				},
			},
			Destination: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "/",
				},
			},
			Scheme: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "scheme",
				},
			},
			Body: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "",
				},
			},
			Headers: map[string][]v2.MatcherViewV6{
				"Hoverfly": {
					{
						Matcher: matchers.Exact,
						Value:   "testing",
					},
				},
			},
		},
	}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{encodedPair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(0))

	Expect(hv.Simulation.GetMatchingPairs()[0]).ToNot(Equal(models.RequestResponsePair{
		Response: models.ResponseDetails{
			Status:  200,
			Body:    "hello_world",
			Headers: map[string][]string{"Content-Encoding": {"gzip"}}},
		Request: models.RequestDetails{
			Path:        "/",
			Method:      "GET",
			Destination: "/",
			Scheme:      "scheme",
			Query:       map[string][]string{},
			Body:        "",
			Headers:     map[string][]string{"Hoverfly": {"testing"}}}}))
}

func TestImportImportRequestResponsePairs_SetsState(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	encodedPair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        base64String("hello_world"),
			EncodedBody: true,
			Headers:     map[string][]string{"Content-Encoding": {"gzip"}}},
		RequestMatcher: v2.RequestMatcherViewV6{
			Body: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "",
				},
			},
			RequiresState: map[string]string{
				"sequence:1": "1",
			},
		},
	}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{encodedPair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(0))

	Eventually(func() string {
		val, _ := hv.state.GetState("sequence:1")
		return val
	}).Should(Equal("1"))
}

func TestImportImportRequestResponsePairsMultipleTimes_SetsAllStates(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	pair1 := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status: 200,
			Body:   "hello_world"},
		RequestMatcher: v2.RequestMatcherViewV6{
			Path: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "/1",
				},
			},
			RequiresState: map[string]string{
				"sequence:1": "1",
			},
		},
	}
	pair2 := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status: 200,
			Body:   "hello_world"},
		RequestMatcher: v2.RequestMatcherViewV6{
			Body: []v2.MatcherViewV6{
				{
					Matcher: matchers.Exact,
					Value:   "/2",
				},
			},
			RequiresState: map[string]string{
				"sequence:2": "1",
			},
		},
	}

	result1 := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{pair1})
	Expect(result1.GetError()).To(BeNil())
	Expect(result1.WarningMessages).To(HaveLen(0))

	result2 := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{pair2})
	Expect(result2.GetError()).To(BeNil())
	Expect(result2.WarningMessages).To(HaveLen(0))

	Eventually(func() string {
		val, _ := hv.state.GetState("sequence:1")
		return val
	}).Should(Equal("1"))
	Eventually(func() string {
		val, _ := hv.state.GetState("sequence:2")
		return val
	}).Should(Equal("1"))
}

func TestImportImportRequestResponsePairs_ReturnsWarningsIfDeprecatedQuerytSet(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	encodedPair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        base64String("hello_world"),
			EncodedBody: true,
			Headers:     map[string][]string{"Content-Encoding": {"gzip"}}},
		RequestMatcher: v2.RequestMatcherViewV6{
			DeprecatedQuery: []v2.MatcherViewV6{
				{
					Matcher: "exact",
					Value:   "deprecated",
				},
			},
		},
	}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{encodedPair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(1))
	Expect(result.WarningMessages[0].Message).To(ContainSubstring("data.pairs[0].request.deprecatedQuery"))
}

func TestImportImportRequestResponsePairs_ReturnsWarningsContentLengthAndTransferEncodingSet(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	encodedPair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        base64String("hello_world"),
			EncodedBody: true,
			Headers: map[string][]string{
				"Content-Length":    {"11"},
				"Transfer-Encoding": {"chunked"},
			},
		},
		RequestMatcher: v2.RequestMatcherViewV6{
			Destination: []v2.MatcherViewV6{
				{
					Matcher: "exact",
					Value:   "hoverfly.io",
				},
			},
		},
	}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{encodedPair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(1))
	Expect(result.WarningMessages[0].Message).To(ContainSubstring("Response contains both Content-Length and Transfer-Encoding headers on data.pairs[0].response"))
}

func TestImportImportRequestResponsePairs_ReturnsWarningsContentLengthMismatch(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	encodedPair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        base64String("hello_world"),
			EncodedBody: true,
			Headers: map[string][]string{
				"Content-Length": {"1234"},
			},
		},
		RequestMatcher: v2.RequestMatcherViewV6{
			Destination: []v2.MatcherViewV6{
				{
					Matcher: "exact",
					Value:   "hoverfly.io",
				},
			},
		},
	}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{encodedPair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(1))
	Expect(result.WarningMessages[0].Message).To(ContainSubstring("Response contains incorrect Content-Length header on data.pairs[0].response, please correct or remove header"))
}

func TestImportRequestResponsePairs_ReturnsWarningsIfAPairIsNotAddedDueToConflict(t *testing.T) {
	RegisterTestingT(t)

	pair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        base64String("hello_world"),
			EncodedBody: false,
		},
		RequestMatcher: v2.RequestMatcherViewV6{
			Destination: []v2.MatcherViewV6{
				{
					Matcher: "exact",
					Value:   "hoverfly.io",
				},
			},
		},
	}

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{pair})
	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(0))

	// Importing it the second time
	result = hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{pair})
	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(1))
	Expect(result.WarningMessages[0].Message).To(ContainSubstring("data.pairs[0] is not added due to a conflict with the existing simulation"))
}

func TestImportImportRequestResponsePairs_ReturnsNoWarnings(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	encodedPair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status: 200,
			Body:   "hello_world",
			Headers: map[string][]string{
				"Content-Length": {"11"},
			},
		},
		RequestMatcher: v2.RequestMatcherViewV6{
			Destination: []v2.MatcherViewV6{
				{
					Matcher: "exact",
					Value:   "hoverfly.io",
				},
			},
		},
	}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{encodedPair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(0))
}

func TestImportImportRequestResponsePairs_ReturnsNoWarnings_Encoded(t *testing.T) {
	RegisterTestingT(t)

	cache := cache.NewDefaultLRUCache()
	cfg := Configuration{Webserver: false}
	cacheMatcher := matching.CacheMatcher{RequestCache: cache, Webserver: cfg.Webserver}
	hv := Hoverfly{Cfg: &cfg, CacheMatcher: cacheMatcher, Simulation: models.NewSimulation()}

	encodedPair := v2.RequestMatcherResponsePairViewV6{
		Response: v2.ResponseDetailsViewV6{
			Status:      200,
			Body:        base64String("hello_world"),
			EncodedBody: true,
			Headers: map[string][]string{
				"Content-Length": {"11"},
			},
		},
		RequestMatcher: v2.RequestMatcherViewV6{
			Destination: []v2.MatcherViewV6{
				{
					Matcher: "exact",
					Value:   "hoverfly.io",
				},
			},
		},
	}

	result := hv.importRequestResponsePairViews([]v2.RequestMatcherResponsePairViewV6{encodedPair})

	Expect(result.GetError()).To(BeNil())
	Expect(result.WarningMessages).To(HaveLen(0))
}
