package illumioapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TrafficAnalysisRequest represents the payload object for the traffic analysis POST request
type TrafficAnalysisRequest struct {
	Sources          Sources          `json:"sources"`
	Destinations     Destinations     `json:"destinations"`
	ExplorerServices ExplorerServices `json:"services"`
	StartDate        time.Time        `json:"start_date,omitempty"`
	EndDate          time.Time        `json:"end_date,omitempty"`
	PolicyDecisions  []string         `json:"policy_decisions"`
	MaxResults       int              `json:"max_results,omitempty"`
}

// Sources represents the sources query portion of the explorer API
type Sources struct {
	Include [][]Include `json:"include"`
	Exclude []Exclude   `json:"exclude"`
}

// ExplorerServices represent services to be included or excluded in the explorer query
type ExplorerServices struct {
	Include []Include `json:"include"`
	Exclude []Exclude `json:"exclude"`
}

//Destinations represents the destination query portion of the explorer API
type Destinations struct {
	Include [][]Include `json:"include"`
	Exclude []Exclude   `json:"exclude"`
}

// PortProtos represents the ports and protocols query portion of the exporer API
type PortProtos struct {
	Include []Include `json:"include"`
	Exclude []Exclude `json:"exclude"`
}

// Include represents the type of objects used in an include query.
// The include struct should be label only, workload only, IP address only, Port and/or protocol only.
// Example - Label and Workload cannot both be non-nil
// Example - Port and Proto can both be non-nil (e.g., port 3306 and proto 6)
type Include struct {
	Label          *Label     `json:"label,omitempty"`
	Workload       *Workload  `json:"workload,omitempty"`
	IPAddress      *IPAddress `json:"ip_address,omitempty"`
	Port           int        `json:"port,omitempty"`
	ToPort         int        `json:"to_port,omitempty"`
	Proto          int        `json:"proto,omitempty"`
	Process        string     `json:"process_name,omitempty"`
	WindowsService string     `json:"windows_service_name,omitempty"`
}

// Exclude represents the type of objects used in an include query.
// The exclude struct should only have the following combinations: label only, workload only, IP address only, Port and/or protocol only.
// Example - Label and Workload cannot both be non-nil
// Example - Port and Proto can both be non-nil (e.g., port 3306 and proto 6)
type Exclude struct {
	Label          *Label     `json:"label,omitempty"`
	Workload       *Workload  `json:"workload,omitempty"`
	IPAddress      *IPAddress `json:"ip_address,omitempty"`
	Port           int        `json:"port,omitempty"`
	ToPort         int        `json:"to_port,omitempty"`
	Proto          int        `json:"proto,omitempty"`
	Process        string     `json:"process_name,omitempty"`
	WindowsService string     `json:"windows_service_name,omitempty"`
}

// IPAddress represents an IP Address
type IPAddress struct {
	Value string `json:"value,omitempty"`
}

// TrafficAnalysis represents the response from the explorer API
type TrafficAnalysis struct {
	Dst            *Dst            `json:"dst"`
	NumConnections int             `json:"num_connections"`
	PolicyDecision string          `json:"policy_decision"`
	ExpSrv         *ExpSrv         `json:"service"`
	Src            *Src            `json:"src"`
	TimestampRange *TimestampRange `json:"timestamp_range"`
}

// ExpSrv is a service in the explorer response
type ExpSrv struct {
	Port           int    `json:"port,omitempty"`
	Proto          int    `json:"proto,omitempty"`
	Process        string `json:"process_name,omitempty"`
	WindowsService string `json:"windows_service_name,omitempty"`
}

// Dst is the provider workload details
type Dst struct {
	IP       string    `json:"ip"`
	Workload *Workload `json:"workload,omitempty"`
}

// Src is the consumer workload details
type Src struct {
	IP       string    `json:"ip"`
	Workload *Workload `json:"workload,omitempty"`
}

// TimestampRange is used to limit queries ranges for the flow detected
type TimestampRange struct {
	FirstDetected string `json:"first_detected"`
	LastDetected  string `json:"last_detected"`
}

// TrafficQuery is the struct to be passed to the GetTrafficAnalysis function
type TrafficQuery struct {
	SourcesInclude        []string
	SourcesExclude        []string
	DestinationsInclude   []string
	DestinationsExclude   []string
	PortProtoInclude      [][2]int
	PortProtoExclude      [][2]int
	PortRangeInclude      [][2]int
	PortRangeExclude      [][2]int
	ProcessInclude        []string
	WindowsServiceInclude []string
	ProcessExclude        []string
	WindowsServiceExclude []string
	StartTime             time.Time
	EndTime               time.Time
	PolicyStatuses        []string
	MaxFLows              int
}

// GetTrafficAnalysis gets flow data from Explorer.
func GetTrafficAnalysis(pce PCE, query TrafficQuery) ([]TrafficAnalysis, APIResponse, error) {
	var api APIResponse

	// Initialize arrays using "make" so JSON is marshaled with empty arrays and not null values to meet Illumio API spec
	sourceInc := make([]Include, 0)
	destInc := make([]Include, 0)

	sourceExcl := make([]Exclude, 0)
	destExcl := make([]Exclude, 0)

	// Process source include, destination include, source exclude, and destination exclude
	queryLists := [][]string{query.SourcesInclude, query.DestinationsInclude, query.SourcesExclude, query.DestinationsExclude}

	// Start counter
	i := 0

	// For each list there are 4 possibilities: empty, label, workload, ipaddress
	for _, queryList := range queryLists {

		// Labels
		if len(queryList) > 0 {
			if strings.Contains(queryList[0], "label") == true {
				for _, label := range queryLists[i] {
					queryLabel := Label{Href: label}
					switch i {
					case 0:
						sourceInc = append(sourceInc, Include{Label: &queryLabel})
					case 1:
						destInc = append(destInc, Include{Label: &queryLabel})
					case 2:
						sourceExcl = append(sourceExcl, Exclude{Label: &queryLabel})
					case 3:
						destExcl = append(destExcl, Exclude{Label: &queryLabel})
					}

				}

				// Workloads
			} else if strings.Contains(queryList[0], "workload") == true {
				for _, workload := range queryLists[i] {
					queryWorkload := Workload{Href: workload}
					switch i {
					case 0:
						sourceInc = append(sourceInc, Include{Workload: &queryWorkload})
					case 1:
						destInc = append(destInc, Include{Workload: &queryWorkload})
					case 2:
						sourceExcl = append(sourceExcl, Exclude{Workload: &queryWorkload})
					case 3:
						destExcl = append(destExcl, Exclude{Workload: &queryWorkload})
					}

				}

				// Assume all else are IP addresses (API will error when needed)
			} else if len(queryList[0]) > 0 {
				for _, ipAddress := range queryLists[i] {
					queryIPAddress := IPAddress{Value: ipAddress}
					switch i {
					case 0:
						sourceInc = append(sourceInc, Include{IPAddress: &queryIPAddress})
					case 1:
						destInc = append(destInc, Include{IPAddress: &queryIPAddress})
					case 2:
						sourceExcl = append(sourceExcl, Exclude{IPAddress: &queryIPAddress})
					case 3:
						destExcl = append(destExcl, Exclude{IPAddress: &queryIPAddress})
					}
				}
			}
		}

		i++
	}

	// Get the service data ready
	serviceInclude := make([]Include, 0)
	serviceExclude := make([]Exclude, 0)

	// Port and protocol - include
	for _, portProto := range query.PortProtoInclude {
		serviceInclude = append(serviceInclude, Include{Port: portProto[0], Proto: portProto[1]})
	}

	// Port and protocol - exclude
	for _, portProto := range query.PortProtoExclude {
		serviceExclude = append(serviceExclude, Exclude{Port: portProto[0], Proto: portProto[1]})
	}

	// Port Range - include
	for _, portRange := range query.PortRangeInclude {
		serviceInclude = append(serviceInclude, Include{Port: portRange[0], ToPort: portRange[1]})
	}

	// Port Range - exclude
	for _, portRange := range query.PortRangeExclude {
		serviceExclude = append(serviceExclude, Exclude{Port: portRange[0], ToPort: portRange[1]})
	}

	// Process - include
	for _, process := range query.ProcessInclude {
		serviceInclude = append(serviceInclude, Include{Process: process})
	}

	// Process - exclude
	for _, process := range query.ProcessExclude {
		serviceExclude = append(serviceExclude, Exclude{Process: process})
	}

	// Windows Service - include
	for _, winSrv := range query.WindowsServiceInclude {
		serviceInclude = append(serviceInclude, Include{WindowsService: winSrv})
	}

	// Windows Service - exclude
	for _, winSrv := range query.WindowsServiceExclude {
		serviceExclude = append(serviceExclude, Exclude{WindowsService: winSrv})
	}

	// Build the TrafficAnalysisRequest struct
	traffic := TrafficAnalysisRequest{
		Sources: Sources{
			Include: [][]Include{sourceInc},
			Exclude: sourceExcl},
		Destinations: Destinations{
			Include: [][]Include{destInc},
			Exclude: destExcl},
		ExplorerServices: ExplorerServices{
			Include: serviceInclude,
			Exclude: serviceExclude},
		PolicyDecisions: query.PolicyStatuses,
		StartDate:       query.StartTime,
		EndDate:         query.EndTime,
		MaxResults:      query.MaxFLows}

	// Create JSON Payload
	jsonPayload, err := json.Marshal(traffic)
	if err != nil {
		return nil, api, fmt.Errorf("get traffic analysis - %s", err)
	}

	var trafficResponses []TrafficAnalysis

	// Build the API URL
	apiURL, err := url.Parse("https://" + pceSanitization(pce.FQDN) + ":" + strconv.Itoa(pce.Port) + "/api/v1/orgs/" + strconv.Itoa(pce.Org) + "/traffic_flows/traffic_analysis_queries")
	if err != nil {
		return nil, api, fmt.Errorf("get traffic analysis - %s", err)
	}

	// Call the API
	api, err = apicall("POST", apiURL.String(), pce, jsonPayload, false)
	if err != nil {
		return nil, api, fmt.Errorf("get traffic analysis - %s", err)
	}

	// Unmarshal response to struct
	json.Unmarshal([]byte(api.RespBody), &trafficResponses)

	return trafficResponses, api, nil

}
