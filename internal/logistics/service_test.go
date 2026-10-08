package logistics

import "testing"

func TestDeriveBatchProgress(t *testing.T) {
	tests := []struct {
		name      string
		batch     string
		statuses  map[string]int
		shipments int
		want      string
	}{
		{"no shipments retains batch state", "packed", nil, 0, "packed"},
		{"all delivered", "dispatched", map[string]int{"received": 2}, 2, "received"},
		{"mixed delivered", "dispatched", map[string]int{"received": 1, "dispatched": 1}, 2, "partially_received"},
		{"one route departed", "planned", map[string]int{"dispatched": 1, "planned": 1}, 2, "partially_dispatched"},
		{"all routes departed", "planned", map[string]int{"dispatched": 2}, 2, "dispatched"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := deriveBatchProgress(test.batch, test.statuses, test.shipments); got != test.want {
				t.Fatalf("deriveBatchProgress() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestShipmentTransportModes(t *testing.T) {
	for _, mode := range []string{"ground", "airlift", "sea_lift", "mixed", "other"} {
		if !validTransportMode(mode) {
			t.Fatalf("%q should be valid", mode)
		}
	}
	if validTransportMode("aircraft") {
		t.Fatal("unexpected transport mode was accepted")
	}
}
