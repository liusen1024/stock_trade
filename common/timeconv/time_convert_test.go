package timeconv

import (
	"reflect"
	"testing"
	"time"
)

func TestInt32ToTime(t *testing.T) {
	type args struct {
		v int32
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "20120304",
			args: args{
				20120101,
			},
			want:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.Local),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Int32ToTime(tt.args.v)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int32ToTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt64ToTime(t *testing.T) {
	type args struct {
		v int64
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "20120101020304500",
			args: args{
				20120101020304500,
			},
			want:    time.Date(2012, 1, 1, 2, 3, 4, 500*1e6, time.Local),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Int64ToTime(tt.args.v)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int64ToTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeToInt32(t *testing.T) {
	type args struct {
		v time.Time
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "20120101",
			args: args{
				time.Date(2012, 1, 1, 0, 0, 0, 0, time.Local),
			},
			want: 20120101,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeToInt32(tt.args.v); got != tt.want {
				t.Errorf("TimeToInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeToInt64(t *testing.T) {
	type args struct {
		v time.Time
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "20120101020304500",
			args: args{
				time.Date(2012, 1, 1, 2, 3, 4, 500*1e6, time.Local),
			},
			want: 20120101020304500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeToInt64(tt.args.v); got != tt.want {
				t.Errorf("TimeToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}
