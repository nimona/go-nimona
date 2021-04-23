package peerstore

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
)

type (
	PeerCache struct {
		m                   sync.Map
		promKnownPeersGauge prometheus.Gauge
		promGCedPeersGauge  prometheus.Gauge
		promIncPeersGauge   prometheus.Gauge
	}
)

type entry struct {
	ttl       time.Duration
	createdAt time.Time
	pr        *hyperspace.Announcement
}

var promMetrics = map[string]prometheus.Gauge{}

func NewPeerCache(
	gcTime time.Duration,
	metricPrefix string,
) *PeerCache {
	promKnownPeersGauge, ok := promMetrics[metricPrefix+"_known_peers"]
	if !ok {
		promKnownPeersGauge = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: metricPrefix + "_known_peers",
				Help: "Total number of known peers",
			},
		)
		promMetrics[metricPrefix+"_known_peers"] = promKnownPeersGauge
	}
	promIncPeersGauge, ok := promMetrics[metricPrefix+"_incoming_peers"]
	if !ok {
		promIncPeersGauge = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: metricPrefix + "_incoming_peers",
				Help: "Total number of incoming peers",
			},
		)
		promMetrics[metricPrefix+"_incoming_peers"] = promIncPeersGauge
	}
	promGCedPeersGauge, ok := promMetrics[metricPrefix+"_gced_peers"]
	if !ok {
		promGCedPeersGauge = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: metricPrefix + "_gced_peers",
				Help: "Total number of GCed peers",
			},
		)
		promMetrics[metricPrefix+"_gced_peers"] = promGCedPeersGauge
	}
	pc := &PeerCache{
		m:                   sync.Map{},
		promKnownPeersGauge: promKnownPeersGauge,
		promIncPeersGauge:   promIncPeersGauge,
		promGCedPeersGauge:  promGCedPeersGauge,
	}
	go func() {
		for {
			time.Sleep(gcTime)
			pc.m.Range(func(key, value interface{}) bool {
				e := value.(entry)
				if e.ttl != 0 {
					now := time.Now()
					diff := now.Sub(e.createdAt)
					if diff >= e.ttl {
						pc.m.Delete(key)
						pc.promKnownPeersGauge.Dec()
						pc.promGCedPeersGauge.Inc()
					}
				}
				return true
			})
		}
	}()
	return pc
}

// Put -
func (m *PeerCache) Put(
	p *hyperspace.Announcement,
	ttl time.Duration,
) (updated bool) {
	// check if we already have a newer announcement
	pann, ok := m.m.Load(p.ConnectionInfo.PublicKey.String())
	// if the announcement is already know update it, but return that
	// updated was false, this is done to renew the created attribute
	updated = true
	if ok {
		if pann.(entry).pr.Version > p.Version {
			return false
		}
		if pann.(entry).pr.Version == p.Version {
			updated = false
		}
	} else {
		m.promKnownPeersGauge.Inc()
	}
	// increment the incoming peers counter
	m.promIncPeersGauge.Inc()
	// and finally store it
	m.m.Store(p.ConnectionInfo.PublicKey.String(), entry{
		ttl:       ttl,
		createdAt: time.Now(),
		pr:        p,
	})

	return updated
}

// Put -
func (m *PeerCache) Touch(k crypto.PublicKey, ttl time.Duration) {
	v, ok := m.m.Load(k.String())
	if !ok {
		return
	}
	e := v.(entry)
	m.m.Store(k.String(), entry{
		ttl:       ttl,
		createdAt: time.Now(),
		pr:        e.pr,
	})
}

// Get -
func (m *PeerCache) Get(k crypto.PublicKey) (*hyperspace.Announcement, error) {
	p, ok := m.m.Load(k.String())
	if !ok {
		return nil, fmt.Errorf("missing")
	}
	return p.(entry).pr, nil
}

// Remove -
func (m *PeerCache) Remove(k crypto.PublicKey) {
	m.m.Delete(k.String())
	m.promKnownPeersGauge.Dec()
}

// List -
func (m *PeerCache) List() []*hyperspace.Announcement {
	ps := []*hyperspace.Announcement{}
	m.m.Range(func(_, p interface{}) bool {
		ps = append(ps, p.(entry).pr)
		return true
	})
	return ps
}

// Lookup -
func (m *PeerCache) Lookup(q hyperspace.Bloom) []*hyperspace.Announcement {
	ps := []*hyperspace.Announcement{}
	m.m.Range(func(_, p interface{}) bool {
		if hyperspace.Bloom(p.(entry).pr.PeerVector).Test(q) {
			ps = append(ps, p.(entry).pr)
		}
		return true
	})
	return ps
}
