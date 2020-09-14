package westworld2

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type metricsInstrument struct {
	prefix           string
	lock             *sync.Mutex
	txBytes          []*sample
	retxBytes        []*sample
	rxBytes          []*sample
	txPortalSz       []*sample
	duplicateRxBytes []*sample
	duplicateAcks    []*sample
	retxMs           []*sample
	allocations      []*sample
}

type sample struct {
	ts time.Time
	v  int64
}

func newMetricsInstrument(config map[string]interface{}) (Instrument, error) {
	mi := &metricsInstrument{
		lock: new(sync.Mutex),
	}
	if err := mi.configure(config); err != nil {
		return nil, err
	}
	go mi.signalHandler()
	return mi, nil
}

func (self *metricsInstrument) connected(_ *net.UDPAddr) {
	self.txBytes = nil
	self.retxBytes = nil
	self.rxBytes = nil
	self.txPortalSz = nil
	self.duplicateRxBytes = nil
	self.duplicateAcks = nil
	self.retxMs = nil
	self.allocations = nil
	logrus.Infof("new connection, metrics collection reset")
}

func (self *metricsInstrument) closed(_ *net.UDPAddr) {
	if err := self.writeAllSamples(); err != nil {
		logrus.Errorf("error writing samples (%v)", err)
	}
}

func (self *metricsInstrument) wireMessageRx(_ *net.UDPAddr, wm *wireMessage) {
	self.lock.Lock()
	self.rxBytes = append(self.rxBytes, &sample{time.Now(), int64(len(wm.data))})
	self.lock.Unlock()
}

func (self *metricsInstrument) wireMessageTx(_ *net.UDPAddr, wm *wireMessage) {
	self.lock.Lock()
	self.txBytes = append(self.txBytes, &sample{time.Now(), int64(len(wm.data))})
	self.lock.Unlock()
}

func (self *metricsInstrument) wireMessageRetx(_ *net.UDPAddr, wm *wireMessage) {
	self.lock.Lock()
	self.retxBytes = append(self.retxBytes, &sample{time.Now(), int64(len(wm.data))})
	self.lock.Unlock()
}

func (self *metricsInstrument) portalCapacitySz(_ *net.UDPAddr, capacity int) {
	self.lock.Lock()
	self.txPortalSz = append(self.txPortalSz, &sample{time.Now(), int64(capacity)})
	self.lock.Unlock()
}

func (self *metricsInstrument) unknownPeer(peer *net.UDPAddr) {
	logrus.Errorf("unknownPeer (%s)", peer)
}

func (self *metricsInstrument) readError(_ *net.UDPAddr, err error) {
	logrus.Errorf("readError (%v)", err)
}

func (self *metricsInstrument) connectError(_ *net.UDPAddr, _ error) {
}

func (self *metricsInstrument) unexpectedMessageType(_ *net.UDPAddr, mt messageType) {
	logrus.Errorf("unexpectedMessageType (%s)", mt.string())
}

func (self *metricsInstrument) duplicateRx(_ *net.UDPAddr, wm *wireMessage) {
	self.lock.Lock()
	self.duplicateRxBytes = append(self.duplicateRxBytes, &sample{time.Now(), int64(len(wm.data))})
	self.lock.Unlock()
}

func (self *metricsInstrument) duplicateAck(_ *net.UDPAddr, _ int32) {
	self.lock.Lock()
	self.duplicateAcks = append(self.duplicateAcks, &sample{time.Now(), 1})
	self.lock.Unlock()
}

func (self *metricsInstrument) newRetxMs(_ *net.UDPAddr, retxMs int) {
	self.lock.Lock()
	self.retxMs = append(self.retxMs, &sample{time.Now(), int64(retxMs)})
	self.lock.Unlock()
}

func (self *metricsInstrument) allocate(_ string) {
	self.lock.Lock()
	self.allocations = append(self.allocations, &sample{time.Now(), 1})
	self.lock.Unlock()
}

func (self *metricsInstrument) configure(data map[string]interface{}) error {
	if v, found := data["prefix"]; found {
		if prefix, ok := v.(string); ok {
			self.prefix = prefix
			logrus.Infof("writing metrics data to prefix [%s]", self.prefix)
		} else {
			return errors.New("invalid 'prefix' type")
		}
	}
	return nil
}

func (self *metricsInstrument) writeAllSamples() error {
	if err := os.MkdirAll(self.prefix, os.ModePerm); err == nil {
		outPath, err := ioutil.TempDir(self.prefix, "")
		if err == nil {
			logrus.Infof("writing metrics to prefix [%s]", outPath)
			if err := self.writeSamples("txBytes", outPath, self.txBytes); err != nil {
				logrus.Errorf("error writing txBytes (%v)", err)
			}
			if err := self.writeSamples("retxBytes", outPath, self.retxBytes); err != nil {
				logrus.Errorf("error writing retxBytes (%v)", err)
			}
			if err := self.writeSamples("rxBytes", outPath, self.rxBytes); err != nil {
				logrus.Errorf("error writing rxBytes (%v)", err)
			}
			if err := self.writeSamples("txPortalSz", outPath, self.txPortalSz); err != nil {
				logrus.Errorf("error writing txPortalSz (%v)", err)
			}
			if err := self.writeSamples("duplicateRxBytes", outPath, self.duplicateRxBytes); err != nil {
				logrus.Errorf("error writing duplicateRxBytes (%v)", err)
			}
			if err := self.writeSamples("duplicateAcks", outPath, self.duplicateAcks); err != nil {
				logrus.Errorf("error writing duplicateAcks (%v)", err)
			}
			if err := self.writeSamples("retxMs", outPath, self.retxMs); err != nil {
				logrus.Errorf("error writing retxMs (%v)", err)
			}
			if err := self.writeSamples("allocations", outPath, self.allocations); err != nil {
				logrus.Errorf("error writing allocations (%v)", err)
			}
		} else {
			logrus.Errorf("error writing metrics (%v)", err)
		}
	} else {
		logrus.Errorf("unable to make output parent [%s] (%v)", self.prefix, err)
	}
	return nil
}

func (self *metricsInstrument) writeSamples(name, outPath string, samples []*sample) error {
	path := filepath.Join(outPath, fmt.Sprintf("%s.csv", name))
	oF, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		_ = oF.Close()
	}()
	for _, sample := range samples {
		line := fmt.Sprintf("%d,%d\n", sample.ts.UnixNano(), sample.v)
		n, err := oF.Write([]byte(line))
		if err != nil {
			return err
		}
		if n != len(line) {
			return errors.New("short write")
		}
	}
	logrus.Infof("wrote [%d] samples to [%s]", len(samples), path)
	return nil
}

func (self *metricsInstrument) signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR2)

	for {
		s := <-c
		if s == syscall.SIGUSR2 {
			if err := self.writeAllSamples(); err != nil {
				logrus.Errorf("error writing samples (%v)", err)
			}
		}
	}
}
