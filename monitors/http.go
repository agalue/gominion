package monitors

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/tools"
)

// HTTPMonitor represents a Monitor implementation
type HTTPMonitor struct {
	Name        string
	Scheme      string
	DefaultPort int
}

// GetID gets the monitor ID (simple class name from its Java counterpart)
func (monitor *HTTPMonitor) GetID() string {
	return monitor.Name
}

// Poll execute the HTTP monitor request and return the the poller response
func (monitor *HTTPMonitor) Poll(request *api.PollerRequestDTO) *api.PollerResponseDTO {
	response := &api.PollerResponseDTO{Status: &api.PollStatus{}}
	start := time.Now()
	client := monitor.getClient(request)
	httpreq, err := monitor.getHTTPRequest(request)
	if err != nil {
		response.Status.Down(err.Error())
		return response
	}
	httpres, err := client.Do(httpreq)
	if err != nil {
		response.Status.Down(err.Error())
		return response
	}
	min, max := monitor.getResponseRange(request)
	if httpres.StatusCode < min || httpres.StatusCode > max {
		response.Status.Down(fmt.Sprintf("Response code %d out of expected range: %d-%d", httpres.StatusCode, min, max))
		return response
	}
	responseText := request.GetAttributeValue("response-text", "")
	if responseText != "" {
		data, err := ioutil.ReadAll(httpres.Body)
		if err != nil {
			response.Status.Down(err.Error())
			return response
		}
		if !strings.Contains(string(data), responseText) {
			response.Status.Down(fmt.Sprintf("Response doesn't containt text %s", responseText))
			return response
		}
	}
	duration := time.Since(start)
	response.Status.Up(duration.Seconds())
	return response
}

func (monitor *HTTPMonitor) getURL(request *api.PollerRequestDTO) *url.URL {
	u := &url.URL{
		Scheme: monitor.Scheme,
		Host:   monitor.getHost(request),
		Path:   request.GetAttributeValue("url", "/"),
	}
	queryString := request.GetAttributeValue("queryString", "")
	if queryString != "" {
		u.RawQuery = queryString
	}
	return u
}

func (monitor *HTTPMonitor) getHTTPRequest(request *api.PollerRequestDTO) (*http.Request, error) {
	u := monitor.getURL(request)
	httpreq, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	basicAuth := request.GetAttributeValue("basic-authentication", "")
	if basicAuth != "" {
		parts := strings.Split(basicAuth, ":")
		if len(parts) == 2 {
			httpreq.SetBasicAuth(parts[0], parts[1])
		}
	} else {
		user := request.GetAttributeValue("user", "")
		passwd := request.GetAttributeValue("password", "")
		if user != "" && passwd != "" {
			httpreq.SetBasicAuth(user, passwd)
		}
	}
	userAgent := request.GetAttributeValue("userAgent", "")
	if userAgent != "" {
		httpreq.Header.Set("User-Agent", userAgent)
	}
	hostName := request.GetAttributeValue("host-name", "")
	if hostName == "" {
		useNodeLabel, _ := strconv.ParseBool(request.GetAttributeValue("nodelabel-host-name", "false"))
		if useNodeLabel {
			httpreq.Header.Set("Host", request.NodeLabel)
		}
	} else {
		httpreq.Header.Set("Host", hostName)
	}
	return httpreq, nil
}

func (monitor *HTTPMonitor) getClient(request *api.PollerRequestDTO) *http.Client {
	useSSLFilter, _ := strconv.ParseBool(request.GetAttributeValue("use-ssl-filter", "false"))
	return tools.GetHTTPClient(useSSLFilter, request.GetTimeout())
}

func (monitor *HTTPMonitor) getHost(request *api.PollerRequestDTO) string {
	port := request.GetAttributeValue("port", fmt.Sprintf("%d", monitor.DefaultPort))
	return request.IPAddress + ":" + port
}

func (monitor *HTTPMonitor) getResponseRange(request *api.PollerRequestDTO) (int, int) {
	defaultRange := "100-399"
	if url := request.GetAttributeValue("url", "/"); url == "/" {
		defaultRange = "100-499"
	}
	responseRange := request.GetAttributeValue("response", defaultRange)
	return tools.ParseHTTPResponseRange(responseRange)
}

func init() {
	RegisterMonitor(&HTTPMonitor{Name: "HttpMonitor", Scheme: "http", DefaultPort: 80})
	RegisterMonitor(&HTTPMonitor{Name: "HttpsMonitor", Scheme: "https", DefaultPort: 443})
}
