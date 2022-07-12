package video_test

import (
	"testing"

	"github.com/meteorae/meteorae-server/scanners/video"
)

func TestCleanName(t *testing.T) {
	t.Parallel()

	type args struct {
		name string
	}

	tests := []struct {
		name  string
		args  args
		want  string
		want1 int
	}{
		{
			name:  "StandardFile",
			args:  args{name: "12.Angry.Men.1957.BluRay.1080p.FLAC.x264-DON.mkv"},
			want:  "12 Angry Men",
			want1: 1957,
		},
		{
			name:  "DoubleGarbageToken",
			args:  args{name: "Internal.Affairs.1990-INTERNAL.mkv"},
			want:  "Internal Affairs",
			want1: 1990,
		},
		{
			name:  "YearInName",
			args:  args{name: "Blade.Runner.2049.2017.1080p.BluRay.DTS.x264-ZQ.mkv"},
			want:  "Blade Runner 2049",
			want1: 2017,
		},
		{
			name:  "UnicodeName",
			args:  args{name: "올드보이.2003.Repack.2160p.UHD.BluRay.Remux.HDR.HEVC.DTS-HD.MA.7.1-PmP.mkv"},
			want:  "올드보이",
			want1: 2003,
		},
		{
			name:  "NameWithSpaces",
			args:  args{name: "Oldboy (2003) (2160p UHD Dolby Vision DTS Remux) - ramyDoVi@PTP.mkv"},
			want:  "Oldboy",
			want1: 2003,
		},
		{
			name:  "NameWithoutYear",
			args:  args{name: "Oldboy.Repack.2160p.UHD.BluRay.Remux.HDR.HEVC.DTS-HD.MA.7.1-PmP.mkv"},
			want:  "Oldboy",
			want1: 0,
		},
		{
			name:  "NameOnly",
			args:  args{name: "Oldboy.mkv"},
			want:  "Oldboy Mkv",
			want1: 0,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, got1 := video.CleanName(tc.args.name)
			if got != tc.want {
				t.Errorf("CleanName() got = %v, want %v", got, tc.want)
			}
			if got1 != tc.want1 {
				t.Errorf("CleanName() got1 = %v, want %v", got1, tc.want1)
			}
		})
	}
}

func TestGetSource(t *testing.T) {
	t.Parallel()

	type args struct {
		name string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "BluRay",
			args: args{name: "Blade.Runner.2049.2017.1080p.BluRay.DTS.x264-ZQ.mkv"},
			want: "bluray",
		},
		{
			name: "CAM",
			args: args{name: "Amor.1980.(Robert.Beavers).CAM.x264.mkv"},
			want: "cam",
		},
		{
			name: "NoSource",
			args: args{name: "Oldboy.mkv"},
			want: "",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := video.GetSource(tc.args.name); got != tc.want {
				t.Errorf("GetSource() = %v, want %v", got, tc.want)
			}
		})
	}
}
