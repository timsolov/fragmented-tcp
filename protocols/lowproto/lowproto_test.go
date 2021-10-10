package lowproto

import (
	"net"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestConn_ReadPacket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := NewMockNetConn(ctrl)
	conn.EXPECT().SetDeadline(gomock.Any()).Return(nil).AnyTimes()

	type fields struct {
		config Config
		conn   net.Conn
	}
	tests := []struct {
		name       string
		fields     fields
		wantPacket []byte
		wantErr    bool
		prepare    func() func()
	}{
		{
			name: "success",
			fields: fields{
				config: Config{},
				conn:   conn,
			},
			wantErr:    false,
			wantPacket: []byte{0x41, 0x42},
			prepare: func() func() {
				conn.EXPECT().Read(gomock.Any()).DoAndReturn(func(b []byte) (n int, err error) {
					b[0] = 0x00
					b[1] = 0x02
					return 2, nil
				})
				conn.EXPECT().Read(gomock.Any()).DoAndReturn(func(b []byte) (n int, err error) {
					b[0] = 0x41
					b[1] = 0x42
					return 2, nil
				})
				return nil
			},
		},
		{
			name: "error",
			fields: fields{
				config: Config{},
				conn:   conn,
			},
			wantErr: true,
			prepare: func() func() {
				conn.EXPECT().Read(gomock.Any()).DoAndReturn(func(b []byte) (n int, err error) {
					b[0] = 0x00
					return 1, nil
				})
				return nil
			},
		},
		// it's possible to write more tests but it's not neccessary now because of time
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				clean := tt.prepare()
				if clean != nil {
					clean()
				}
			}

			c := &Conn{
				config: tt.fields.config,
				conn:   tt.fields.conn,
			}
			gotPacket, err := c.ReadPacket()
			if (err != nil) != tt.wantErr {
				t.Errorf("Conn.ReadPacket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPacket, tt.wantPacket) {
				t.Errorf("Conn.ReadPacket() = %v, want %v", gotPacket, tt.wantPacket)
			}
		})
	}
}

func TestConn_WritePacket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := NewMockNetConn(ctrl)
	conn.EXPECT().SetDeadline(gomock.Any()).Return(nil).AnyTimes()

	type fields struct {
		config Config
		conn   net.Conn
	}
	type args struct {
		packet []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		prepare func() func()
	}{
		{
			name: "success",
			fields: fields{
				config: Config{},
				conn:   conn,
			},
			wantErr: false,
			args: args{
				packet: []byte{0x41, 0x42, 0x41},
			},
			prepare: func() func() {
				conn.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (n int, err error) {
					assert.Equal(t, []byte{0x00, 0x03, 0x41, 0x42, 0x41}, b)
					return 5, nil
				})
				return nil
			},
		},
		{
			name: "not all bytes written",
			fields: fields{
				config: Config{},
				conn:   conn,
			},
			wantErr: true,
			args: args{
				packet: []byte{0x41, 0x42, 0x41},
			},
			prepare: func() func() {
				conn.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (n int, err error) {
					assert.Equal(t, []byte{0x00, 0x03, 0x41, 0x42, 0x41}, b)
					return 4, nil
				})
				return nil
			},
		},
		// it's possible to write more tests but it's not neccessary now because of time
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				clean := tt.prepare()
				if clean != nil {
					clean()
				}
			}

			c := &Conn{
				config: tt.fields.config,
				conn:   tt.fields.conn,
			}
			if err := c.WritePacket(tt.args.packet); (err != nil) != tt.wantErr {
				t.Errorf("Conn.WritePacket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
