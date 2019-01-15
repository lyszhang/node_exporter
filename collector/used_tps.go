// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !notime

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/usechain/go-usedrpc"
)

// Namespace defines the common namespace to be used by all metrics.
const used_namespace = "usechain"

type blockCollector struct {
	height        *prometheus.Desc
	totalTx       *prometheus.Desc
	totalTime     *prometheus.Desc
	tps           *prometheus.Desc
	avgDelay      *prometheus.Desc
	maxDelay      *prometheus.Desc
	avgSize       *prometheus.Desc
	maxSize       *prometheus.Desc
	avgTxperBlock *prometheus.Desc
}

func init() {
	registerCollector("block", defaultEnabled, NewBlockCollector)
}

// NewBlockCollector returns a new Collector exposing the current system block stat of used client
func NewBlockCollector() (Collector, error) {
	return &blockCollector{
		height: prometheus.NewDesc(
			used_namespace+"_height",
			"Usechain block height",
			nil, nil,
		),
		totalTx: prometheus.NewDesc(
			used_namespace+"_totalTx",
			"Usechain totalTx in recent 100 blocks",
			nil, nil,
		),
		totalTime: prometheus.NewDesc(
			used_namespace+"_totalTime",
			"Usechain totalTime in recent 100 blocks",
			nil, nil,
		),
		tps: prometheus.NewDesc(
			used_namespace+"_tps",
			"Usechain tps in recent 100 blocks",
			nil, nil,
		),
		avgDelay: prometheus.NewDesc(
			used_namespace+"_avgDelay",
			"Usechain average block delay in recent 100 blocks",
			nil, nil,
		),
		maxDelay: prometheus.NewDesc(
			used_namespace+"_maxDelay",
			"Usechain maximum block delay in recent 100 blocks",
			nil, nil,
		),
		avgSize: prometheus.NewDesc(
			used_namespace+"_avgSize",
			"Usechain average block size in recent 100 blocks",
			nil, nil,
		),
		maxSize: prometheus.NewDesc(
			used_namespace+"_maxSize",
			"Usechain maximum block size in recent 100 blocks",
			nil, nil,
		),
		avgTxperBlock: prometheus.NewDesc(
			used_namespace+"_avgTxperBlock",
			"Usechain average tx number per block",
			nil, nil,
		),
	}, nil
}

func (c *blockCollector) Update(ch chan<- prometheus.Metric) error {
	// create rpc instance
	rpc := usedrpc.NewUseRPC("http://127.0.0.1:8545")

	// get block range info
	end, err := rpc.UseBlockNumber()
	if err != nil {
		return err
	}
	start := end - 100
	if start < 0 {
		start = 0
	}
	slot := end - start
	blockStart, _ := rpc.UseGetBlockByNumber(start, false)
	blockNow, _ := rpc.UseGetBlockByNumber(end, false)
	timeslot := blockNow.Timestamp - blockStart.Timestamp

	// calculate blockchain info
	var totalTxNum, totalSize, maxSize, maxslot int
	for i := start; i < end; i++ {
		num, _ := rpc.UseGetBlockTransactionCountByNumber(i)
		totalTxNum += num
		blockbefore, _ := rpc.UseGetBlockByNumber(i, false)
		block, _ := rpc.UseGetBlockByNumber(i+1, false)
		totalSize += block.Size
		if maxSize < block.Size {
			maxSize = block.Size
		}
		tempslot := block.Timestamp - blockbefore.Timestamp
		if maxslot < tempslot {
			maxslot = tempslot
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.height,
		prometheus.CounterValue,
		float64(end),
	)
	ch <- prometheus.MustNewConstMetric(
		c.totalTx,
		prometheus.GaugeValue,
		float64(totalTxNum),
	)
	ch <- prometheus.MustNewConstMetric(
		c.totalTime,
		prometheus.CounterValue,
		float64(timeslot),
	)
	ch <- prometheus.MustNewConstMetric(
		c.tps,
		prometheus.CounterValue,
		float64(totalTxNum/timeslot),
	)
	ch <- prometheus.MustNewConstMetric(
		c.avgDelay,
		prometheus.CounterValue,
		float64(timeslot/slot),
	)
	ch <- prometheus.MustNewConstMetric(
		c.maxDelay,
		prometheus.CounterValue,
		float64(maxslot),
	)
	ch <- prometheus.MustNewConstMetric(
		c.avgSize,
		prometheus.CounterValue,
		float64(totalSize/slot),
	)
	ch <- prometheus.MustNewConstMetric(
		c.maxSize,
		prometheus.CounterValue,
		float64(maxSize),
	)
	ch <- prometheus.MustNewConstMetric(
		c.avgTxperBlock,
		prometheus.CounterValue,
		float64(totalTxNum/slot),
	)

	return nil
}
