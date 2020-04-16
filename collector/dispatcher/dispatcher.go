package dispatcher

import (
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/SUSE/sap_host_exporter/collector"
	"github.com/SUSE/sap_host_exporter/internal/sapcontrol"
)

func NewCollector(webService sapcontrol.WebService, currentSapInstance sapcontrol.CurrentSapInstance) (*dispatcherCollector, error) {

	c := &dispatcherCollector{
		collector.NewDefaultCollector("dispatcher"),
		webService,
		currentSapInstance,
	}

	c.SetDescriptor("queue_now", "Work process current queue length", []string{"type", "instance_name", "instance_number", "sid", "instance_hostname"})
	c.SetDescriptor("queue_high", "Work process peak queue length", []string{"type", "instance_name", "instance_number", "sid", "instance_hostname"})
	c.SetDescriptor("queue_max", "Work process maximum queue length", []string{"type", "instance_name", "instance_number", "sid", "instance_hostname"})
	c.SetDescriptor("queue_writes", "Work process queue writes", []string{"type", "instance_name", "instance_number", "sid", "instance_hostname"})
	c.SetDescriptor("queue_reads", "Work process queue reads", []string{"type", "instance_name", "instance_number", "sid", "instance_hostname"})

	return c, nil
}

type dispatcherCollector struct {
	collector.DefaultCollector
	webService sapcontrol.WebService
	currentSapInstance sapcontrol.CurrentSapInstance
}

func (c *dispatcherCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting Dispatcher metrics")

	err := c.recordWorkProcessQueueStats(ch)
	if err != nil {
		log.Warnf("Dispatcher Collector scrape failed: %s", err)
		return
	}
}

func (c *dispatcherCollector) recordWorkProcessQueueStats(ch chan<- prometheus.Metric) error {
	queueStatistic, err := c.webService.GetQueueStatistic()

	if err != nil {
		return errors.Wrap(err, "SAPControl web service error")
	}

	commonLabels := []string{
		c.currentSapInstance.Name,
		strconv.Itoa(int(c.currentSapInstance.Number)),
		c.currentSapInstance.SID,
		c.currentSapInstance.Hostname,
	}

	// for each work queue, we record a different line for each stat of that queue, with the type as a common label
	for _, queue := range queueStatistic.Queues {
		labels := append([]string{queue.Type}, commonLabels...)
		ch <- c.MakeGaugeMetric("queue_now", float64(queue.Now), labels...)
		ch <- c.MakeCounterMetric("queue_high", float64(queue.High), labels...)
		ch <- c.MakeGaugeMetric("queue_max", float64(queue.Max), labels...)
		ch <- c.MakeCounterMetric("queue_writes", float64(queue.Writes), labels...)
		ch <- c.MakeCounterMetric("queue_reads", float64(queue.Reads), labels...)
	}

	return nil
}
