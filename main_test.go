package t_test

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
)

func TestSubnetName(t *testing.T) {
	subnetEvm := ids.ID{'s', 'u', 'b', 'n', 'e', 't', 'e', 'v', 'm'}
	t.Logf("SubnetEVM: %s", subnetEvm.String())

	landslideVm := ids.ID{'l', 'a', 'n', 'd', 's', 'l', 'i', 'd', 'e', 'v', 'm'}
	t.Logf("LandslideVM: %s", landslideVm.String())

	landslideWasm := ids.ID{'l', 'a', 'n', 'd', 's', 'l', 'i', 'd', 'e', 'w', 'a', 's', 'm'}
	t.Logf("LandslideWasm: %s", landslideWasm.String())
}
