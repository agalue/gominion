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

// Poll execute the monitor request and return the service status
func (monitor *HTTPMonitor) Poll(request *api.PollerRequestDTO) api.PollStatus {
	status := api.PollStatus{}
	start := time.Now()
	client := monitor.getClient(request)
	u := url.URL{
		Scheme: monitor.Scheme,
		Host:   monitor.getHost(request),
		Path:   request.GetAttributeValue("url", "/"),
	}
	queryString := request.GetAttributeValue("queryString", "")
	if queryString != "" {
		u.RawQuery = queryString
	}
	httpreq, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		status.Down(err.Error())
		return status
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
	response, err := client.Do(httpreq)
	if err != nil {
		status.Down(err.Error())
		return status
	}
	min, max := monitor.getResponseRange(request)
	if response.StatusCode < min || response.StatusCode > max {
		status.Down(fmt.Sprintf("Response code %d out of expected range: %d-%d", response.StatusCode, min, max))
		return status
	}
	responseText := request.GetAttributeValue("response-text", "")
	if responseText != "" {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			status.Down(err.Error())
			return status
		}
		if !strings.Contains(string(data), responseText) {
			status.Down(fmt.Sprintf("Response doesn't containt text %s", responseText))
			return status
		}
	}

	duration := time.Since(start)
	status.Up(duration.Seconds())
	return status
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

var httpMonitor = &HTTPMonitor{Name: "HttpMonitor", Scheme: "http", DefaultPort: 80}
var httpsMonitor = &HTTPMonitor{Name: "HttpsMonitor", Scheme: "https", DefaultPort: 443}

func init() {
	RegisterMonitor(httpMonitor)
	RegisterMonitor(httpsMonitor)
}
