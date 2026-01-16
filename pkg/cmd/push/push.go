package push

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cylonchau/pantheon/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	pushExample = templates.Examples(i18n.T(`
		# Push a single metric to pushgateway
		pantheonctl push my_job --address http://127.0.0.1:9091 --metric cpu_usage=0.8

		# Push multiple metrics with labels
		pantheonctl push my_job --address http://127.0.0.1:9091 \
			--label env=prod --label region=us-east \
			--metric cpu_usage=0.8 \
			--metric request_count=100:counter
	`))
)

type PushOptions struct {
	JobName      string
	Address      string
	Labels       []string
	Metrics      []string
	ParsedLabels map[string]string
}

func NewPushOptions() *PushOptions {
	return &PushOptions{
		ParsedLabels: make(map[string]string),
	}
}

func NewCmdPush() *cobra.Command {
	o := NewPushOptions()

	cmd := &cobra.Command{
		Use:     "push <job_name> --address <url> [flags]",
		Short:   i18n.T("Push metrics to a Pushgateway"),
		Long:    i18n.T("Push custom metrics to a Prometheus Pushgateway"),
		Example: pushExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.JobName = args[0]
			if err := o.Complete(cmd); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run()
		},
	}

	cmd.Flags().StringVar(&o.Address, "address", "addr", "Address of the Pushgateway (e.g. http://127.0.0.1:9091)")
	cmd.Flags().StringSliceVarP(&o.Labels, "label", "l", []string{}, "Labels to attach to the metrics (format: key=value). Can be specified multiple times.")
	cmd.Flags().StringSliceVarP(&o.Metrics, "metric", "m", []string{}, "Metrics to push (format: name=value[:type]). Type can be 'gauge' (default) or 'counter'. Can be specified multiple times.")

	cmd.MarkFlagRequired("address")

	return cmd
}

func (o *PushOptions) Complete(cmd *cobra.Command) error {
	// Parse Labels
	for _, l := range o.Labels {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid label format '%s', expected key=value", l)
		}
		o.ParsedLabels[parts[0]] = parts[1]
	}

	// Add mandatory client labels
	o.ParsedLabels["cli_version"] = version.Version

	// Hostname
	if host, err := os.Hostname(); err == nil {
		o.ParsedLabels["hostname"] = host
	}

	// Client IP detection
	// Try to dial the pushgateway address to see which local IP is usedTo reach it
	targetAddr := o.Address
	if strings.Contains(targetAddr, "://") {
		u, err := url.Parse(targetAddr)
		if err == nil {
			targetAddr = u.Host
		}
	}
	// If the host part doesn't have a port, net.Dial might fail if we don't handle it,
	// but mostly users provide host:port for pushgateway or the url logic handles it.
	// We use a short timeout.

	// Try connecting to the target first (most accurate source IP)
	conn, err := net.DialTimeout("tcp", targetAddr, time.Second*2)
	if err == nil {
		defer conn.Close()
		// Get local IP from connection
		o.ParsedLabels["client_ip"] = strings.Split(conn.LocalAddr().String(), ":")[0]
	} else {
		// Fallback: try to dial a common public IP (Google DNS) to find default outbound IP
		conn, err := net.DialTimeout("udp", "8.8.8.8:80", time.Second*2)
		if err == nil {
			defer conn.Close()
			o.ParsedLabels["client_ip"] = conn.LocalAddr().(*net.UDPAddr).IP.String()
		}
		// If both fail, we just don't set client_ip
	}

	return nil
}

func (o *PushOptions) Validate() error {
	if o.Address == "" {
		return fmt.Errorf("--address is required")
	}
	if o.JobName == "" {
		return fmt.Errorf("job name is required")
	}
	if len(o.Metrics) == 0 {
		return fmt.Errorf("at least one --metric is required")
	}
	return nil
}

func (o *PushOptions) Run() error {
	registry := prometheus.NewRegistry()

	for _, m := range o.Metrics {
		name, value, metricType, err := parseMetricString(m)
		if err != nil {
			return err
		}

		switch metricType {
		case "gauge":
			g := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: name,
				Help: fmt.Sprintf("Custom gauge metric %s pushed via pantheonctl-%s", name, version.Version),
			})
			g.Set(value)
			registry.MustRegister(g)
		case "counter":
			c := prometheus.NewCounter(prometheus.CounterOpts{
				Name: name,
				Help: fmt.Sprintf("Custom counter metric %s pushed via pantheonctl-%s", name, version.Version),
			})
			c.Add(value)
			registry.MustRegister(c)
		default:
			return fmt.Errorf("unsupported metric type '%s' for metric '%s'. Supported types: gauge, counter", metricType, name)
		}
	}

	pusher := push.New(o.Address, o.JobName).Gatherer(registry)

	for k, v := range o.ParsedLabels {
		pusher.Grouping(k, v)
	}

	fmt.Printf("Pushing metrics to %s (job: %s)...\n", o.Address, o.JobName)
	if err := pusher.Push(); err != nil {
		return fmt.Errorf("failed to push metrics: %v", err)
	}

	fmt.Println("Successfully pushed metrics.")
	return nil
}

func parseMetricString(m string) (name string, value float64, metricType string, err error) {
	// Format: name=value[:type]
	// type defaults to gauge

	parts := strings.SplitN(m, "=", 2)
	if len(parts) != 2 {
		return "", 0, "", fmt.Errorf("invalid metric format '%s', expected name=value[:type]", m)
	}
	name = parts[0]
	rest := parts[1]

	var valueStr string

	// Check for type
	if strings.Contains(rest, ":") {
		valTypeParts := strings.SplitN(rest, ":", 2)
		valueStr = valTypeParts[0]
		metricType = strings.ToLower(valTypeParts[1])
	} else {
		valueStr = rest
		metricType = "gauge"
	}

	value, err = strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid value for metric '%s': %v", name, err)
	}

	return name, value, metricType, nil
}
