package devices

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// DeviceLog represents the data to store and show logs from specific device
type DeviceLog struct {
	TimeStamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Level     LogLevel  `json:"level"`
}

type LogLevel string

const (
	LogLevel_Error = "error"
	LogLevel_Info  = "info"
)

// DeviceStatus holds all the data from a device to keep track of it's status
type DeviceStatus struct {
	Name string

	Logs     []*DeviceLog `json:"logs"`
	LastSeen time.Time    `json:"lastSeen"`
	State    DeviceState  `json:"state"`

	logsLock *sync.Mutex
}

type DeviceState string

const (
	DeviceState_On    = "on"
	DeviceState_Off   = "off"
	DeviceState_Error = "error"
)

func NewDeviceStatus(name string) *DeviceStatus {
	return &DeviceStatus{
		Name:     name,
		State:    DeviceState_On,
		logsLock: &sync.Mutex{},
	}
}

func (s *DeviceStatus) Log(timeStamp time.Time, level LogLevel, msg string) {
	newLog := &DeviceLog{
		TimeStamp: timeStamp,
		Level:     level,
		Message:   msg,
	}

	s.logsLock.Lock()
	defer s.logsLock.Unlock()

	s.Logs = append(s.Logs, newLog)

	// Make sure the logs array isn't too long
	const MaxLogsHistoryForDevice = 200
	if len(s.Logs) > MaxLogsHistoryForDevice {
		s.Logs = s.Logs[5:]
	}
}

// DevicesMonitor handles monitoring and logging of devices on the field, it samples
// the devices healthcheck and stores all log data
type DevicesMonitor struct {
	Devices     map[string]*DeviceStatus
	devicesLock *sync.Mutex

	maxUnseenDeviceDuration time.Duration

	statusChangedNotifier func()
}

func NewDevicesMonitor(statusChangeNotifier func()) *DevicesMonitor {
	const maxAllowedDeviceUnseen_Seconds = 6
	return &DevicesMonitor{
		Devices:                 map[string]*DeviceStatus{},
		devicesLock:             &sync.Mutex{},
		maxUnseenDeviceDuration: maxAllowedDeviceUnseen_Seconds * time.Second,
		statusChangedNotifier:   statusChangeNotifier,
	}
}

func (c *DevicesMonitor) Init() {
	const deviceSamplingInterval_Seconds = 1
	c.StartDevicesSampler(deviceSamplingInterval_Seconds * time.Second)
}

func (c *DevicesMonitor) SetDeviceSeen(deviceName string) {
	now := time.Now()

	c.devicesLock.Lock()
	defer c.devicesLock.Unlock()

	// Create new device if not exists
	deviceStatus, exists := c.Devices[deviceName]
	if !exists {
		deviceStatus = NewDeviceStatus(deviceName)
		c.Devices[deviceName] = deviceStatus
	}

	// Update last seen
	deviceStatus.LastSeen = now

	if deviceStatus.State == DeviceState_Off {
		log.Default().Printf("Field device [%s] reconnected \n", deviceStatus.Name)

		deviceStatus.State = DeviceState_Error

		deviceStatus.Log(
			now,
			LogLevel_Info,
			"Device was off but now seen and recovered with healthcheck, check the logs",
		)
	}

	c.statusChangedNotifier()
}

func (c *DevicesMonitor) SetDeviceError(deviceName, deviceErr string, extraData ...string) {
	deviceStatus, exists := c.Devices[deviceName]
	if !exists {
		return
	}

	now := time.Now()

	if deviceStatus.State == DeviceState_Off {
		log.Default().Printf("Field device [%s] reconnected \n", deviceStatus.Name)

		deviceStatus.Log(
			now,
			LogLevel_Info,
			"Device was off but now recovered with error, check the logs",
		)
	}

	deviceStatus.State = DeviceState_Error

	msg := fmt.Sprintf("Error: %s", deviceErr)

	if len(extraData) > 0 {
		extraDataMsg := strings.Join(extraData, ", ")
		msg = fmt.Sprintf("Error: %s (%s)", deviceErr, extraDataMsg)
	}

	deviceStatus.Log(
		now,
		LogLevel_Error,
		msg,
	)

	c.statusChangedNotifier()
}

func (c *DevicesMonitor) ResetDeviceError(deviceName string) {
	deviceStatus, exists := c.Devices[deviceName]
	if !exists {
		return
	}

	if deviceStatus.State != DeviceState_Error {
		return
	}

	deviceStatus.State = DeviceState_On

	deviceStatus.Log(
		time.Now(),
		LogLevel_Info,
		"Device error was handled",
	)

	c.statusChangedNotifier()
}

func (c *DevicesMonitor) StartDevicesSampler(interval time.Duration) {
	go func() {
		for {
			c.devicesLock.Lock()

			now := time.Now()

			statusChanged := false

			for _, deviceStatus := range c.Devices {
				// Skip devices that are off
				if deviceStatus.State == DeviceState_Off {
					continue
				}

				unseenDuration := now.Sub(deviceStatus.LastSeen)

				if unseenDuration > c.maxUnseenDeviceDuration {
					deviceStatus.State = DeviceState_Off

					log.Default().Printf("Error: Lost connection to field device [%s]\n", deviceStatus.Name)

					msg := fmt.Sprintf("Device is off, didn't receive health check for longer then [%s]", unseenDuration.String())

					deviceStatus.Log(
						now,
						LogLevel_Error,
						msg,
					)

					statusChanged = true
				}
			}

			if statusChanged {
				c.statusChangedNotifier()
			}

			c.devicesLock.Unlock()

			// Delay the next check to keep this loop from running too much
			time.Sleep(interval)
		}
	}()
}
