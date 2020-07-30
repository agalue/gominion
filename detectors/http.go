package detectors

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
)

// HTTPDetector represents a detector implementation
type HTTPDetector struct {
	ID          string
	Scheme      string
	DefaultPort int
}

// GetID gets the detector ID (simple class name from its Java counterpart)
func (detector *HTTPDetector) GetID() string {
	return detector.ID
}

// Detect execute the detector request and return the service status
func (detector *HTTPDetector) Detect(request *api.DetectorRequestDTO) api.DetectResults {
	results := api.DetectResults{
		IsServiceDetected: detector.isDetected(request),
	}
	return results
}

func (detector *HTTPDetector) getClient(request *api.DetectorRequestDTO) *http.Client {
	useSSLFilter, _ := strconv.ParseBool(request.GetAttributeValue("useSSLFilter", "false"))
	return tools.GetHTTPClient(useSSLFilter, request.GetTimeout())
}

func (detector *HTTPDetector) isDetected(request *api.DetectorRequestDTO) bool {
	if detector.ID != "WebDetector" {
		return detector.isDetectedSimple(request)
	}

	client := detector.getClient(request)
	u := url.URL{
		Scheme: request.GetAttributeValue("scheme", detector.Scheme),
		Host:   detector.getHost(request),
		Path:   request.GetAttributeValue("path", "/"),
	}
	queryString := request.GetAttributeValue("queryString", "")
	if queryString != "" {
		u.RawQuery = queryString
	}
	httpreq, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false
	}
	authEnabled, _ := strconv.ParseBool(request.GetAttributeValue("authEnabled", "false"))
	if authEnabled {
		user := request.GetAttributeValue("authUser", "admin")
		passwd := request.GetAttributeValue("authPassword", "admin")
		httpreq.SetBasicAuth(user, passwd)
	}
	userAgent := request.GetAttributeValue("userAgent", "")
	if userAgent != "" {
		httpreq.Header.Set("User-Agent", userAgent)
	}
	virtualHost := request.GetAttributeValue("virtualHost", "")
	if virtualHost != "" {
		httpreq.Header.Set("Host", virtualHost)
	}
	response, err := client.Do(httpreq)
	if err != nil {
		return false
	}
	min, max := detector.getResponseRange(request)
	if response.StatusCode < min || response.StatusCode > max {
		return false
	}
	responseText := request.GetAttributeValue("responseText", "")
	if responseText != "" {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return false
		}
		if !strings.Contains(string(data), responseText) {
			return false
		}
	}
	return true
}

func (detector *HTTPDetector) getResponseRange(request *api.DetectorRequestDTO) (int, int) {
	responseRange := request.GetAttributeValue("responseRange", "100-399")
	return tools.ParseHTTPResponseRange(responseRange)
}

func (detector *HTTPDetector) getHost(request *api.DetectorRequestDTO) string {
	port := request.GetAttributeValue("port", fmt.Sprintf("%d", detector.DefaultPort))
	return request.IPAddress + ":" + port
}

func (detector *HTTPDetector) isDetectedSimple(request *api.DetectorRequestDTO) bool {
	maxRetCode, _ := strconv.Atoi(request.GetAttributeValue("maxRetCode", "399"))
	checkRetCode, _ := strconv.ParseBool(request.GetAttributeValue("checkRetCode", "false"))
	u := url.URL{
		Scheme: detector.Scheme,
		Host:   detector.getHost(request),
		Path:   request.GetAttributeValue("url", "/"),
	}
	client := detector.getClient(request)
	response, err := client.Get(u.String())
	if err != nil {
		return false
	}
	if checkRetCode {
		return response.StatusCode < maxRetCode
	}
	return true
}

var httpDetector = &HTTPDetector{ID: "HttpDetector", Scheme: "http", DefaultPort: 80}
var httpsDetector = &HTTPDetector{ID: "HttpsDetector", Scheme: "https", DefaultPort: 443}
var webDetector = &HTTPDetector{ID: "WebDetector", Scheme: "http", DefaultPort: 80}

func init() {
	RegisterDetector(httpDetector)
	RegisterDetector(httpsDetector)
	RegisterDetector(webDetector)
}
