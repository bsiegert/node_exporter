// Copyright 2018 The Prometheus Authors
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

//go:build !nocpu
// +build !nocpu

package collector

import (
	"unsafe"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/unix"
)

const (
	CP_USER = iota
	CP_NICE
	CP_SYS
	CP_INTR
	CP_IDLE
	CPUSTATES
)

type cpuCollector struct {
	cpu    typedDesc
	logger log.Logger
}

func init() {
	registerCollector("cpu", defaultEnabled, NewCPUCollector)
}

func NewCPUCollector(logger log.Logger) (Collector, error) {
	return &cpuCollector{
		cpu:    typedDesc{nodeCPUSecondsDesc, prometheus.CounterValue},
		logger: logger,
	}, nil
}

func (c *cpuCollector) Update(ch chan<- prometheus.Metric) (err error) {
	clock, err := unix.SysctlClockinfo("kern.clockrate")
	if err != nil {
		return err
	}
	hz := float64(clock.Stathz)

	cpb, err := unix.SysctlRaw("kern.cp_time")
	if err != nil {
		return err
	}
	var times [CPUSTATES]uint64
	for n := 0; n < len(cpb); n += 8 {
		times[n/8] = *(*uint64)(unsafe.Pointer(&cpb[n]))
	}

	ch <- c.cpu.mustNewConstMetric(float64(times[CP_USER])/hz, "", "user")
	ch <- c.cpu.mustNewConstMetric(float64(times[CP_NICE])/hz, "", "nice")
	ch <- c.cpu.mustNewConstMetric(float64(times[CP_SYS])/hz, "", "system")
	ch <- c.cpu.mustNewConstMetric(float64(times[CP_INTR])/hz, "", "interrupt")
	ch <- c.cpu.mustNewConstMetric(float64(times[CP_IDLE])/hz, "", "idle")

	return err
}
