package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeouts_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Timeouts
		wantErr bool
	}{
		{
			name: "success:seconds-all",
			args: args{data: []byte(`{"shutdown_timeout": "1s", "idle_timeout": "1s", "read_timeout": "1s", "write_timeout": "1s", "read_header_timeout": "1s"}`)},
			want: Timeouts{
				ShutdownTimeout:   1 * time.Second,
				IdleTimeout:       1 * time.Second,
				ReadTimeout:       1 * time.Second,
				WriteTimeout:      1 * time.Second,
				ReadHeaderTimeout: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "success:minutes-all",
			args: args{data: []byte(`{"shutdown_timeout": "1m", "idle_timeout": "1m", "read_timeout": "1m", "write_timeout": "1m", "read_header_timeout": "1m"}`)},
			want: Timeouts{
				ShutdownTimeout:   1 * time.Minute,
				IdleTimeout:       1 * time.Minute,
				ReadTimeout:       1 * time.Minute,
				WriteTimeout:      1 * time.Minute,
				ReadHeaderTimeout: 1 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "success:seconds-and-minutes-all",
			args: args{data: []byte(`{"shutdown_timeout": "5m", "idle_timeout": "5s", "read_timeout": "5s", "write_timeout": "5s", "read_header_timeout": "1m"}`)},
			want: Timeouts{
				ShutdownTimeout:   5 * time.Minute,
				IdleTimeout:       5 * time.Second,
				ReadTimeout:       5 * time.Second,
				WriteTimeout:      5 * time.Second,
				ReadHeaderTimeout: 1 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "success:seconds-shutdown-timeout-only",
			args: args{data: []byte(`{"shutdown_timeout": "30s"}`)},
			want: Timeouts{
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "success:seconds-shutdown-timeout-only-no-units",
			args: args{data: []byte(`{"shutdown_timeout": 30}`)},
			want: Timeouts{
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name:    "fail:seconds-shutdown-timeout-only-no-units-string",
			args:    args{data: []byte(`{"shutdown_timeout": "30"}`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Timeouts{}
			if err := tr.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Timeouts.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tr)
		})
	}
}
