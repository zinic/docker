package libcontainer

import (
	"errors"

	"github.com/syndtr/gocapability/capability"
)

var (
	ErrUnkownNamespace  = errors.New("Unknown namespace")
	ErrUnkownCapability = errors.New("Unknown capability")
	ErrUnsupported      = errors.New("Unsupported method")
)

type Mounts []Mount

func (s Mounts) OfType(t string) Mounts {
	out := Mounts{}
	for _, m := range s {
		if m.Type == t {
			out = append(out, m)
		}
	}
	return out
}

type Mount struct {
	Type        string `json:"type,omitempty"`
	Source      string `json:"source,omitempty"`      // Source path, in the host namespace
	Destination string `json:"destination,omitempty"` // Destination path, in the container
	Writable    bool   `json:"writable,omitempty"`
	Private     bool   `json:"private,omitempty"`
}

// namespaceList is used to convert the libcontainer types
// into the names of the files located in /proc/<pid>/ns/* for
// each namespace
var (
	namespaceList = Namespaces{}

	capabilityList = Capabilities{
		{Key: "SETPCAP", Value: capability.CAP_SETPCAP},
		{Key: "SYS_MODULE", Value: capability.CAP_SYS_MODULE},
		{Key: "SYS_RAWIO", Value: capability.CAP_SYS_RAWIO},
		{Key: "SYS_PACCT", Value: capability.CAP_SYS_PACCT},
		{Key: "SYS_ADMIN", Value: capability.CAP_SYS_ADMIN},
		{Key: "SYS_NICE", Value: capability.CAP_SYS_NICE},
		{Key: "SYS_RESOURCE", Value: capability.CAP_SYS_RESOURCE},
		{Key: "SYS_TIME", Value: capability.CAP_SYS_TIME},
		{Key: "SYS_TTY_CONFIG", Value: capability.CAP_SYS_TTY_CONFIG},
		{Key: "MKNOD", Value: capability.CAP_MKNOD},
		{Key: "AUDIT_WRITE", Value: capability.CAP_AUDIT_WRITE},
		{Key: "AUDIT_CONTROL", Value: capability.CAP_AUDIT_CONTROL},
		{Key: "MAC_OVERRIDE", Value: capability.CAP_MAC_OVERRIDE},
		{Key: "MAC_ADMIN", Value: capability.CAP_MAC_ADMIN},
		{Key: "NET_ADMIN", Value: capability.CAP_NET_ADMIN},
		{Key: "SYSLOG", Value: capability.CAP_SYSLOG},
	}
)

type (
	Namespace struct {
		Key   string `json:"key,omitempty"`
		Value int    `json:"value,omitempty"`
		File  string `json:"file,omitempty"`
	}
	Namespaces []*Namespace
)

func (ns *Namespace) String() string {
	return ns.Key
}

func GetNamespace(key string) *Namespace {
	for _, ns := range namespaceList {
		if ns.Key == key {
			cpy := *ns
			return &cpy
		}
	}
	return nil
}

// Contains returns true if the specified Namespace is
// in the slice
func (n Namespaces) Contains(ns string) bool {
	return n.Get(ns) != nil
}

func (n Namespaces) Get(ns string) *Namespace {
	for _, nsp := range n {
		if nsp != nil && nsp.Key == ns {
			return nsp
		}
	}
	return nil
}

type (
	Capability struct {
		Key   string         `json:"key,omitempty"`
		Value capability.Cap `json:"value,omitempty"`
	}
	Capabilities []*Capability
)

func (c *Capability) String() string {
	return c.Key
}

func GetCapability(key string) *Capability {
	for _, capp := range capabilityList {
		if capp.Key == key {
			cpy := *capp
			return &cpy
		}
	}
	return nil
}

// Contains returns true if the specified Capability is
// in the slice
func (c Capabilities) Contains(capp string) bool {
	return c.Get(capp) != nil
}

func (c Capabilities) Get(capp string) *Capability {
	for _, cap := range c {
		if cap.Key == capp {
			return cap
		}
	}
	return nil
}
