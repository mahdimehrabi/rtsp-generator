package packet

import (
	"errors"
	"log"
	"os"
	"path"
	"testing"
)

func TestH264PacketGenerator_Read(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	var tests = []struct {
		name      string
		assetFile string
		err       error
		trakNum   int8
		stszFull  bool
		mdatEmpty bool
	}{
		{
			name:      "success",
			assetFile: path.Join(wd, "..", "assets", "sample.mp4"),
			err:       nil,
			trakNum:   0,
			stszFull:  true,
			mdatEmpty: false,
		},
		{
			name:      "ErrorTrackNotFound",
			assetFile: path.Join(wd, "..", "assets", "sample.flv"),
			err:       ErrorTrackNotFound,
			trakNum:   -1,
			stszFull:  false,
			mdatEmpty: true,
		},
		{
			name:      "ErrorCodecNotH264",
			assetFile: path.Join(wd, "..", "assets", "sample_h265.mp4"),
			err:       ErrorCodecNotH264,
			trakNum:   0,
			stszFull:  false,
			mdatEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hpg := NewH264PacketGenerator()
			f, err := os.Open(tt.assetFile)
			if err != nil {
				t.Fatal("Cannot open asset file")
			}
			defer f.Close()
			err = hpg.Read(f)
			if !errors.Is(err, tt.err) {
				t.Fatalf("error is not equal to expected error: %v,target:%v", err, tt.err)
			}
			if hpg.trakNum != tt.trakNum {
				t.Fatalf("tack num is not equal %d %d", hpg.trakNum, tt.trakNum)
			}
			if tt.stszFull != (len(hpg.stsz) > 0) {
				t.Fatalf("stsz full flag is not equal to %v", tt.stszFull)
			}
			if (len(hpg.mdat) < 1) != tt.mdatEmpty {
				t.Fatalf("mdatEmpty status is not equal to %v", tt.mdatEmpty)
			}
		})
	}
}
