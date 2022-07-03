package ssh

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 3*1024)
	},
}

func init() {
	Meter = &meter{}
	Meter.measurementsAvg10Seconds = make([]int64, 10)
	Meter.measurementsAvgMinute = make([]int64, 60)
	Meter.mu = new(sync.RWMutex)
	Meter.Run()
}

var Meter *meter

type meter struct {
	bytesPer10Seconds int64
	bytesPerMinute    int64

	liveBytesCounter         int64
	measurementsAvg10Seconds []int64
	measurementsAvgMinute    []int64
	mu                       *sync.RWMutex
}

func (m *meter) GetHumanReadablePer10Seconds() string {
	bytes := m.GetBytesPer10Seconds()
	return m.bytesToHumanReadable(float64(bytes))
}

func (m *meter) GetHumanReadablePerMinute() string {
	bytes := m.GetBytesPerMinute()
	return m.bytesToHumanReadable(float64(bytes))
}

func (m *meter) GetBytesPer10Seconds() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bytesPer10Seconds
}

func (m *meter) GetBytesPerMinute() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bytesPerMinute
}

func (m *meter) RegisterBytesWritten(written int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.liveBytesCounter += written
}

func (m *meter) Run() {
	// Take down the stats every second
	go func() {
		for {
			waiter := time.After(1 * time.Second)
			<-waiter
			m.mu.Lock()
			m.measurementsAvg10Seconds = append([]int64{m.liveBytesCounter}, m.measurementsAvg10Seconds[:len(m.measurementsAvg10Seconds)-1]...)
			m.measurementsAvgMinute = append([]int64{m.liveBytesCounter}, m.measurementsAvgMinute[:len(m.measurementsAvgMinute)-1]...)
			m.liveBytesCounter = int64(0)
			m.mu.Unlock()
		}
	}()
	// Update output fields at the specific intervals
	go func() {
		for {
			waiter := time.After(10 * time.Second)
			<-waiter
			m.mu.Lock()
			avg := int64(0)
			for _, bytes := range m.measurementsAvg10Seconds {
				avg += bytes
			}
			m.bytesPer10Seconds = avg / 10
			m.mu.Unlock()
		}
	}()
	go func() {
		for {
			waiter := time.After(1 * time.Minute)
			<-waiter
			m.mu.Lock()
			avg := int64(0)
			for _, bytes := range m.measurementsAvgMinute {
				avg += bytes
			}
			m.bytesPerMinute = avg / 60
			m.mu.Unlock()
		}
	}()
}

func (m *meter) bytesToHumanReadable(bytes float64) string {
	switch {
	case bytes < math.Pow(1024, 1):
		return fmt.Sprintf("%.f bytes", bytes)
	case bytes < math.Pow(1024, 2):
		return fmt.Sprintf("%.2f KiB", bytes/math.Pow(1024, 1))
	case bytes < math.Pow(1024, 3):
		return fmt.Sprintf("%.2f MiB", bytes/math.Pow(1024, 2))
	case bytes < math.Pow(1024, 4):
		return fmt.Sprintf("%.2f GiB", bytes/math.Pow(1024, 3))
	default:
		return fmt.Sprintf("%.2f TiB", bytes/math.Pow(1024, 4))
	}
}

func CopyAndMeasureThroughput(writer, reader net.Conn) {
	var err error
	written := int64(0)
	for {
		buffer := pool.Get().([]byte)
		bytesRead, readErr := reader.Read(buffer)
		if bytesRead > 0 {
			bytesWritten, writeErr := writer.Write(buffer[0:bytesRead])
			if bytesWritten > 0 {
				written += int64(bytesWritten)
			}
			if writeErr != nil {
				err = writeErr
				break
			}
			if bytesRead != bytesWritten {
				err = io.ErrShortWrite
				break
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			err = readErr
			break
		}
		Meter.RegisterBytesWritten(written)
		written = 0
	}
	if err != nil {
		log.Println("Tunneling connection produced an error", err)
		return
	}
	Meter.RegisterBytesWritten(written)
}
