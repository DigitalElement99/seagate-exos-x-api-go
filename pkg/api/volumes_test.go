package mcapi

import (
	"errors"
	"strings"
	"testing"

	"github.com/Seagate/seagate-exos-x-api-go/v2/pkg/common"
)

// Regression (2026-10-03): the ME5024 answered `show volumes` for a just-copied
// volume with -10058 "does not exist" ~60 ms after reporting it visible.
// GetVolumeWwn returned ("", nil) on that API-level error, so CreateVolume
// minted a volume ID with an empty WWN and the node could never publish it.
func TestVolumeWwnFromShowVolumes(t *testing.T) {
	ok := &common.ResponseStatus{ResponseTypeNumeric: ApiSuccess}
	notExist := &common.ResponseStatus{ResponseTypeNumeric: 1, ReturnCode: -10058,
		Response: "Bad parameter(s) were specified. (k3s_x) - The specified volume does not exist."}
	vol := []common.VolumeObject{{VolumeName: "k3s_x", Wwn: "600C0FF000FB6CA9057CC16A01000000"}}

	cases := []struct {
		name    string
		vols    []common.VolumeObject
		status  *common.ResponseStatus
		err     error
		want    string
		wantErr string
	}{
		{"found", vol, ok, nil, "600c0ff000fb6ca9057cc16a01000000", ""},
		{"transport error", nil, &common.ResponseStatus{}, errors.New("boom"), "", "boom"},
		{"api error", nil, notExist, nil, "", "does not exist"},
		{"nil status", nil, nil, nil, "", "k3s_x"},
		{"no match", []common.VolumeObject{{VolumeName: "other", Wwn: "abc"}}, ok, nil, "", "k3s_x"},
		{"empty wwn", []common.VolumeObject{{VolumeName: "k3s_x"}}, ok, nil, "", "k3s_x"},
	}
	for _, c := range cases {
		got, err := volumeWwnFromShowVolumes("k3s_x", c.vols, c.status, c.err)
		if got != c.want {
			t.Errorf("%s: wwn = %q, want %q", c.name, got, c.want)
		}
		if c.wantErr == "" && err != nil {
			t.Errorf("%s: unexpected error %v", c.name, err)
		}
		if c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)) {
			t.Errorf("%s: err = %v, want containing %q", c.name, err, c.wantErr)
		}
	}
}

// Regression (2026-10-22): the CreateVolume retry for a snapshot clone
// asked for the snapshot's restoreSize (blocks reported as bytes) and the exact
// size match returned AlreadyExists for the full-size clone it had just made.
func TestVolumeExistsIn(t *testing.T) {
	vols := []common.VolumeObject{{VolumeName: "k3s_x", Blocks: 209715200, BlockSize: 512}} // 100Gi
	cases := []struct {
		name       string
		id         string
		size       int64
		wantExists bool
		wantErr    bool
	}{
		{"exact", "k3s_x", 107374182400, true, false},
		{"larger than requested", "k3s_x", 209715200, true, false},
		{"smaller than requested", "k3s_x", 107374182401, true, true},
		{"absent", "k3s_y", 1, false, false},
	}
	for _, c := range cases {
		exists, err := volumeExistsIn(c.id, c.size, vols)
		if exists != c.wantExists || (err != nil) != c.wantErr {
			t.Errorf("%s: got (%v, %v), want exists=%v err=%v", c.name, exists, err, c.wantExists, c.wantErr)
		}
	}
}
