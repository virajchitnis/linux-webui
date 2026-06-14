package dbus

import (
	"fmt"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	systemdDest  = "org.freedesktop.systemd1"
	systemdPath  = "/org/freedesktop/systemd1"
	managerIface = "org.freedesktop.systemd1.Manager"
	unitIface    = "org.freedesktop.systemd1.Unit"
)

// Client wraps a D-Bus system connection for systemd operations.
type Client struct {
	mu   sync.Mutex
	conn *dbus.Conn
}

func New() (*Client, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("connect system bus: %w", err)
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Close() {
	_ = c.conn.Close()
}

// UnitInfo holds basic unit information.
type UnitInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LoadState   string `json:"load_state"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
	UnitPath    string `json:"-"`
}

// ListUnits returns all loaded systemd units.
func (c *Client) ListUnits() ([]UnitInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	obj := c.conn.Object(systemdDest, systemdPath)
	var result [][]interface{}
	if err := obj.Call(managerIface+".ListUnits", 0).Store(&result); err != nil {
		return nil, fmt.Errorf("ListUnits: %w", err)
	}

	units := make([]UnitInfo, 0, len(result))
	for _, u := range result {
		if len(u) < 7 {
			continue
		}
		name, _ := u[0].(string)
		desc, _ := u[1].(string)
		load, _ := u[2].(string)
		active, _ := u[3].(string)
		sub, _ := u[4].(string)
		path, _ := u[6].(dbus.ObjectPath)
		units = append(units, UnitInfo{
			Name:        name,
			Description: desc,
			LoadState:   load,
			ActiveState: active,
			SubState:    sub,
			UnitPath:    string(path),
		})
	}
	return units, nil
}

// unitIsValid checks that the unit name exists in the live D-Bus unit list.
func (c *Client) unitIsValid(name string) bool {
	// Basic sanity checks before D-Bus lookup
	if name == "" || strings.ContainsAny(name, " \t\n\r/\\") {
		return false
	}
	units, err := c.ListUnits()
	if err != nil {
		return false
	}
	for _, u := range units {
		if u.Name == name {
			return true
		}
	}
	return false
}

func (c *Client) doUnitAction(action, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	obj := c.conn.Object(systemdDest, systemdPath)
	var jobPath dbus.ObjectPath
	method := managerIface + "." + action
	return obj.Call(method, 0, name, "replace").Store(&jobPath)
}

func (c *Client) StartUnit(name string) error {
	if !c.unitIsValid(name) {
		return fmt.Errorf("unknown unit: %s", name)
	}
	return c.doUnitAction("StartUnit", name)
}

func (c *Client) StopUnit(name string) error {
	if !c.unitIsValid(name) {
		return fmt.Errorf("unknown unit: %s", name)
	}
	return c.doUnitAction("StopUnit", name)
}

func (c *Client) RestartUnit(name string) error {
	if !c.unitIsValid(name) {
		return fmt.Errorf("unknown unit: %s", name)
	}
	return c.doUnitAction("RestartUnit", name)
}

func (c *Client) EnableUnit(name string) error {
	if !c.unitIsValid(name) {
		return fmt.Errorf("unknown unit: %s", name)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	obj := c.conn.Object(systemdDest, systemdPath)
	var _, _ interface{}
	return obj.Call(managerIface+".EnableUnitFiles", 0, []string{name}, false, true).Err
}

func (c *Client) DisableUnit(name string) error {
	if !c.unitIsValid(name) {
		return fmt.Errorf("unknown unit: %s", name)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	obj := c.conn.Object(systemdDest, systemdPath)
	return obj.Call(managerIface+".DisableUnitFiles", 0, []string{name}, false).Err
}
