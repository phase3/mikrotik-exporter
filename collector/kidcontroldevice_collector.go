package collector

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
)

type kidControlDeviceCollector struct {
	props        []string
	descriptions map[string]*prometheus.Desc
}

func newKidControlDeviceCollector() routerOSCollector {
	c := &kidControlDeviceCollector{}
	c.init()
	return c
}

func (c *kidControlDeviceCollector) init() {
	c.props = []string{"ip-address", "bytes-down", "bytes-up"}
	labelNames := []string{"name", "address", "ip_address"}
	c.descriptions = make(map[string]*prometheus.Desc)
	for _, p := range c.props {
		c.descriptions[p] = descriptionForPropertyName("kidcontroldevice", p, labelNames)
	}
}

func (c *kidControlDeviceCollector) describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *kidControlDeviceCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectForStat(re, ctx)
	}

	return nil
}

func (c *kidControlDeviceCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/ip/kid-control/device/print", "=.proplist="+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching interface metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *kidControlDeviceCollector) collectForStat(re *proto.Sentence, ctx *collectorContext) {
	for _, p := range c.props[1:] {
		c.collectMetricForProperty(p, re, ctx)
	}
}
func parseBytes(value string) float64 {
	var (
		v   float64
		err error
	)
	if strings.HasSuffix(value, "KiB") {
		s := value[:len(value)-3]
		v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return -1
		}
		v *= 1024

	} else if strings.HasSuffix(value, "MiB") {
		s := value[:len(value)-3]
		v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return -1
		}
		v *= 1024000

	} else if strings.HasSuffix(value, "GiB") {
		s := value[:len(value)-3]
		v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return -1
		}
		v *= 1024000000

	} else {
		v, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return -1
		}
	}
	return v

}

func (c *kidControlDeviceCollector) collectMetricForProperty(property string, re *proto.Sentence, ctx *collectorContext) {
	desc := c.descriptions[property]
	ip := re.Map["ip-address"]
	if ip == "" {
		return
	}
	if value := re.Map[property]; value != "" {
		var v float64
		v = parseBytes(value)
		if v == -1 {
			log.WithFields(log.Fields{
				"device":    ctx.device.Name,
				"interface": re.Map["name"],
				"property":  property,
				"value":     value,
				"error":     "Unable to parse float64 from string",
			}).Error("error parsing interface metric value")
			return

		}
		if strings.Contains(ip, ",") {
			ips := strings.Split(ip, ",")
			for _, i := range ips {
				ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, v, ctx.device.Name, ctx.device.Address, i)
			}
		} else {
			ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, v, ctx.device.Name, ctx.device.Address, ip)
		}
	}
}
