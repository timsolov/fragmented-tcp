package highproto

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		packet []byte
	}
	tests := []struct {
		name       string
		args       args
		wantKind   MessageKind
		wantParams []string
		wantErr    bool
	}{
		{
			name: "HI",
			args: args{
				packet: []byte("HI Tim"),
			},
			wantKind:   HI,
			wantParams: []string{"Tim"},
			wantErr:    false,
		},
		{
			name: "CLIENTS",
			args: args{
				packet: []byte("CLIENTS"),
			},
			wantKind: CLIENTS,
			wantErr:  false,
		},
		// tests for other cases
		// I can't write all tests because of time.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKind, gotParams, err := Parse(tt.args.packet)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotKind != tt.wantKind {
				t.Errorf("Parse() gotKind = %v, want %v", gotKind, tt.wantKind)
			}
			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("Parse() gotParams = %v, want %v", gotParams, tt.wantParams)
			}
		})
	}
}
