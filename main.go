package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bmatcuk/doublestar"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var sniffCertificatePEM *regexp.Regexp = regexp.MustCompile(`-----BEGIN CERTIFICATE-----`)
var sniffJWT *regexp.Regexp = regexp.MustCompile(`([A-Za-z0-9_-]+)\.([A-Za-z0-9_-]+)\.[A-Za-z0-9_-]+\s*`) // base64 url encoded.

var startTime time.Time

func uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
}

// see https://github.com/kubernetes/client-go/tree/v0.30.0/examples/in-cluster-client-configuration
// see https://github.com/kubernetes/client-go/blob/v0.30.0/kubernetes/typed/core/v1/pod.go
func getPodContainers() (string, error) {
	if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
		return "", nil
	}
	podNamespace := os.Getenv("POD_NAMESPACE")
	podName := os.Getenv("POD_NAME")
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}
	pod, err := client.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, container := range pod.Spec.Containers {
		fmt.Fprintf(&b, "%s %s\n", container.Name, container.Image)
	}
	return b.String(), nil
}

func getCertificateText(pem []byte) (string, error) {
	cmd := exec.Command("openssl", "x509", "-text")
	cmd.Stdin = bytes.NewReader(pem)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getJWTText(encodedHeader []byte, encodedPayload []byte) (string, error) {
	header, err := base64.RawURLEncoding.DecodeString(string(encodedHeader))
	if err != nil {
		return "", fmt.Errorf("failed to decode encodedHeader: %v", err)
	}
	payload, err := base64.RawURLEncoding.DecodeString(string(encodedPayload))
	if err != nil {
		return "", fmt.Errorf("failed to decode encodedPayload: %v", err)
	}
	return fmt.Sprintf(
		"header: %s\n\npayload: %s",
		getPrettyJSON(header),
		getPrettyJSON(payload)), nil
}

func getPrettyJSON(jsonString []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonString, "", "  "); err != nil {
		return string(jsonString)
	}
	return prettyJSON.String()
}

func getFileText(path string) string {
	value, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("ERROR %v", err)
	}

	if sniffCertificatePEM.Find(value) != nil {
		info, err := getCertificateText(value)
		if err != nil {
			info = fmt.Sprintf("ERROR %v", err)
		}
		return info
	}

	jwtMatches := sniffJWT.FindSubmatch(value)
	if jwtMatches != nil {
		info, err := getJWTText(jwtMatches[1], jwtMatches[2])
		if err != nil {
			info = fmt.Sprintf("ERROR %v", err)
		}
		return fmt.Sprintf("%s\n\n%s", info, value)
	}

	return string(value)
}

func getFilesText(parentDirectoryPath string) []nameValuePair {
	filesText := make([]nameValuePair, 0)
	filesPath, _ := doublestar.Glob(fmt.Sprintf("%s/**", parentDirectoryPath))
	for _, v := range filesPath {
		if strings.Contains(v, "/.") {
			continue
		}
		fi, err := os.Stat(v)
		if err != nil {
			log.Fatal(err)
		}
		mode := fi.Mode()
		if !mode.IsRegular() {
			continue
		}
		uid := fi.Sys().(*syscall.Stat_t).Uid
		gid := fi.Sys().(*syscall.Stat_t).Gid
		name := v[len(parentDirectoryPath)+1:]
		value := getFileText(v)
		filesText = append(filesText, nameValuePair{fmt.Sprintf("%s %d %d %s", mode, uid, gid, name), value})
	}
	sort.Sort(nameValuePairs(filesText))
	return filesText
}

var indexTemplate = template.Must(template.New("Index").Parse(`<!DOCTYPE html>
<html>
<head>
<title>kubernetes-hello</title>
<style>
body {
    font-family: monospace;
    color: #555;
    background: #e6edf4;
    padding: 1.25rem;
    margin: 0;
}
table {
    background: #fff;
    border: .0625rem solid #c4cdda;
    border-radius: 0 0 .25rem .25rem;
    border-spacing: 0;
    margin-bottom: 1.25rem;
    padding: .75rem 1.25rem;
    text-align: left;
    width: 100%;
}
table > caption {
    background: #f1f6fb;
    text-align: left;
    font-weight: bold;
    padding: .75rem 1.25rem;
    border: .0625rem solid #c4cdda;
    border-radius: .25rem .25rem 0 0;
    border-bottom: 0;
}
table td, table th {
    padding: .25rem;
    white-space: pre-wrap;
}
table td {
    overflow-wrap: anywhere;
}
table > tbody > tr:hover {
    background: #f1f6fb;
}
</style>
</head>
<body>
    <table>
        <caption>Properties</caption>
        <tbody>
            <tr><th>Pid</th><td>{{.Pid}}</td></tr>
            <tr><th>Uid</th><td>{{.Uid}}</td></tr>
            <tr><th>Gid</th><td>{{.Gid}}</td></tr>
            <tr><th>Request</th><td>{{.Request}}</td></tr>
            <tr><th>Client Address</th><td>{{.ClientAddress}}</td></tr>
            <tr><th>Server Address</th><td>{{.ServerAddress}}</td></tr>
            <tr><th>Hostname</th><td>{{.Hostname}}</td></tr>
            <tr><th>Pod Containers</th><td>{{.PodContainers}}</td></tr>
            <tr><th>Cgroup</th><td>{{.Cgroup}}</td></tr>
            <tr><th>Go Max CPUs</th><td>{{.GoMaxCPUs}}</td></tr>
            <tr><th>Memory Limit</th><td>{{.MemoryLimit}}</td></tr>
            <tr><th>Memory Usage</th><td>{{.MemoryUsage}}</td></tr>
            <tr><th>Os</th><td>{{.Os}}</td></tr>
            <tr><th>Architecture</th><td>{{.Architecture}}</td></tr>
            <tr><th>Runtime</th><td>{{.Runtime}}</td></tr>
            <tr><th>Uptime</th><td>{{.Uptime}}</td></tr>
        </tbody>
    </table>
    <table>
        <caption>Request Headers</caption>
        <tbody>
            {{- range .RequestHeaders}}
            <tr>
                <th>{{.Name}}</th>
                <td>{{.Value}}</td>
            </tr>
            {{- end}}
        </tbody>
    </table>
    <table>
        <caption>Environment Variables</caption>
        <tbody>
            {{- range .Environment}}
            <tr>
                <th>{{.Name}}</th>
                <td>{{.Value}}</td>
            </tr>
            {{- end}}
        </tbody>
    </table>
    {{- if .Secrets}}
    <table>
        <caption>Secrets</caption>
        <tbody>
            {{- range .Secrets}}
            <tr>
                <th>{{.Name}}</th>
                <td>{{.Value}}</td>
            </tr>
            {{- end}}
        </tbody>
    </table>
    {{- end}}
    {{- if .Configs}}
    <table>
        <caption>Configs</caption>
        <tbody>
            {{- range .Configs}}
            <tr>
                <th>{{.Name}}</th>
                <td>{{.Value}}</td>
            </tr>
            {{- end}}
        </tbody>
    </table>
    {{- end}}
    {{- if .AzureDNSZones}}
    <table>
        <caption>Azure DNS Zones</caption>
        <tbody>
            {{- range .AzureDNSZones}}
            <tr>
                <th>{{.Name}}</th>
                <td>{{.Value}}</td>
            </tr>
            {{- end}}
        </tbody>
    </table>
    {{- end}}
    {{- if .AWSDNSZones}}
    <table>
        <caption>AWS DNS Zones</caption>
        <tbody>
            {{- range .AWSDNSZones}}
            <tr>
                <th>{{.Name}}</th>
                <td>{{.Value}}</td>
            </tr>
            {{- end}}
        </tbody>
    </table>
    {{- end}}
</body>
</html>
`))

type nameValuePair struct {
	Name  string
	Value string
}

type indexData struct {
	Pid            int
	Uid            int
	Gid            int
	PodContainers  string
	Cgroup         string
	GoMaxCPUs      int
	MemoryLimit    uint64
	MemoryUsage    uint64
	Request        string
	RequestHeaders []nameValuePair
	ClientAddress  string
	ServerAddress  string
	Hostname       string
	Os             string
	Architecture   string
	Runtime        string
	Uptime         string
	Environment    []nameValuePair
	Secrets        []nameValuePair
	Configs        []nameValuePair
	AzureDNSZones  []nameValuePair
	AWSDNSZones    []nameValuePair
}

type nameValuePairs []nameValuePair

func (a nameValuePairs) Len() int           { return len(a) }
func (a nameValuePairs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a nameValuePairs) Less(i, j int) bool { return a[i].Name < a[j].Name }

func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("begin handling request", "method", r.Method, "host", r.Host, "url", r.URL)
		handler.ServeHTTP(w, r)
		slog.Info("end handling request", "method", r.Method, "host", r.Host, "url", r.URL)
	}
}

func headersToNameValuePairs(headers http.Header) nameValuePairs {
	var pairs nameValuePairs
	for name, values := range headers {
		for _, value := range values {
			pairs = append(pairs, nameValuePair{name, value})
		}
	}
	sort.Sort(nameValuePairs(pairs))
	return pairs
}

type MemoryInfo struct {
	Limit uint64
	Usage uint64
}

func getMemoryInfo() (*MemoryInfo, error) {
	// cgroup v1.
	if _, err := os.Stat("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		limitBytes, err := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes")
		if err != nil {
			return nil, err
		}

		usageBytes, err := os.ReadFile("/sys/fs/cgroup/memory/memory.usage_in_bytes")
		if err != nil {
			return nil, err
		}

		limit, err := strconv.ParseUint(string(bytes.TrimSpace(limitBytes)), 10, 64)
		if err != nil {
			return nil, err
		}

		usage, err := strconv.ParseUint(string(bytes.TrimSpace(usageBytes)), 10, 64)
		if err != nil {
			return nil, err
		}

		return &MemoryInfo{
			Limit: limit,
			Usage: usage,
		}, nil
	}

	// cgroup v2.
	if _, err := os.Stat("/sys/fs/cgroup/memory.max"); err == nil {
		limitBytes, err := os.ReadFile("/sys/fs/cgroup/memory.max")
		if err != nil {
			return nil, err
		}

		usageBytes, err := os.ReadFile("/sys/fs/cgroup/memory.current")
		if err != nil {
			return nil, err
		}

		limit, err := strconv.ParseUint(string(bytes.TrimSpace(limitBytes)), 10, 64)
		if err != nil {
			return nil, err
		}

		usage, err := strconv.ParseUint(string(bytes.TrimSpace(usageBytes)), 10, 64)
		if err != nil {
			return nil, err
		}

		return &MemoryInfo{
			Limit: limit,
			Usage: usage,
		}, nil
	}

	// get this process memory information.
	var limit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_DATA, &limit)
	if err != nil {
		return nil, err
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryInfo{
		Limit: limit.Cur,
		Usage: m.Sys,
	}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	var listenAddress = flag.String("listen", ":8000", "Listen address.")

	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		log.Fatalf("\nERROR You MUST NOT pass any positional arguments")
	}

	http.HandleFunc("/", logRequest(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		hostname, err := os.Hostname()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")

		environment := make([]nameValuePair, 0)
		for _, v := range os.Environ() {
			parts := strings.SplitN(v, "=", 2)
			name := parts[0]
			value := parts[1]
			switch name {
			case "PATH":
				fallthrough
			case "XDG_DATA_DIRS":
				fallthrough
			case "XDG_CONFIG_DIRS":
				value = strings.Join(
					strings.Split(value, string(os.PathListSeparator)),
					"\n")
			}
			environment = append(environment, nameValuePair{name, value})
		}
		sort.Sort(nameValuePairs(environment))

		secrets := getFilesText("/var/run/secrets")
		configs := getFilesText("/var/run/configs")

		cgroup, err := os.ReadFile("/proc/self/cgroup")
		if err != nil {
			panic(err)
		}

		memoryInfo, err := getMemoryInfo()
		if err != nil {
			panic(err)
		}

		requestHeaders := headersToNameValuePairs(r.Header)

		podContainers, err := getPodContainers()
		if err != nil {
			panic(err)
		}

		azureDNSZones := getAzureDNSZones(r.Context())
		awsDNSZones := getAWSDNSZones(r.Context())

		err = indexTemplate.ExecuteTemplate(w, "Index", indexData{
			Pid:            os.Getpid(),
			Uid:            os.Getuid(),
			Gid:            os.Getgid(),
			PodContainers:  podContainers,
			Cgroup:         string(cgroup),
			GoMaxCPUs:      runtime.GOMAXPROCS(0),
			MemoryLimit:    memoryInfo.Limit,
			MemoryUsage:    memoryInfo.Usage,
			Request:        fmt.Sprintf("%s %s%s", r.Method, r.Host, r.URL),
			RequestHeaders: requestHeaders,
			ClientAddress:  r.RemoteAddr,
			ServerAddress:  r.Context().Value(http.LocalAddrContextKey).(net.Addr).String(),
			Hostname:       hostname,
			Os:             runtime.GOOS,
			Architecture:   runtime.GOARCH,
			Runtime:        runtime.Version(),
			Uptime:         uptime().String(),
			Environment:    environment,
			Secrets:        secrets,
			Configs:        configs,
			AzureDNSZones:  azureDNSZones,
			AWSDNSZones:    awsDNSZones,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	slog.Info("listening", "addr", fmt.Sprintf("http://%s", *listenAddress))

	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatalf("Failed to ListenAndServe: %v", err)
	}
}
