package libcontainer

import (
	"encoding/json"
	"os"
	"testing"
)

func TestContainerJsonFormat(t *testing.T) {
	f, err := os.Open("container.json")
	if err != nil {
		t.Fatal("Unable to open container.json")
	}
	defer f.Close()

	var container *Container
	if err := json.NewDecoder(f).Decode(&container); err != nil {
		t.Fatalf("failed to decode container config: %s", err)
	}
	if container.Hostname != "koye" {
		t.Log("hostname is not set")
		t.Fail()
	}

	if !container.Tty {
		t.Log("tty should be set to true")
		t.Fail()
	}

	if !container.Namespaces["NEWNET"] {
		t.Log("namespaces should contain NEWNET")
		t.Fail()
	}

	if container.Namespaces["NEWUSER"] {
		t.Log("namespaces should not contain NEWUSER")
		t.Fail()
	}

	if _, exists := container.CapabilitiesMask["SYS_ADMIN"]; !exists {
		t.Log("capabilities mask should contain SYS_ADMIN")
		t.Fail()
	}

	if container.CapabilitiesMask["SYS_ADMIN"] {
		t.Log("SYS_ADMIN should not be enabled in capabilities mask")
		t.Fail()
	}

	if !container.CapabilitiesMask["MKNOD"] {
		t.Log("MKNOD should be enabled in capabilities mask")
		t.Fail()
	}

	if container.CapabilitiesMask["SYS_CHROOT"] {
		t.Log("capabilities mask should not contain SYS_CHROOT")
		t.Fail()
	}
}
