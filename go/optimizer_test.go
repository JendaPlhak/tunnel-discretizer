package main

import (
	"testing"
)

func TestOptimizerSimple(t *testing.T) {
	t.Run("../tunnels/tun_cl_007_1.pdb", func(t *testing.T) {
		tunnel := LoadTunnelFromPdbFile("../tunnels/tun_cl_007_1.pdb")
		disks := generateInitialDisks(tunnel)
		optimizeDisks(tunnel, disks)
	})
}
