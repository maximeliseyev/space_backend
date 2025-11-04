package utils

import (
	"testing"
	"time"
)

func TestParseFlexibleTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "RFC3339 full format",
			input:   "2006-01-02T15:04:05Z",
			wantErr: false,
		},
		{
			name:    "RFC3339 with timezone",
			input:   "2006-01-02T15:04:05+07:00",
			wantErr: false,
		},
		{
			name:    "Date only",
			input:   "2025-11-01",
			wantErr: false,
		},
		{
			name:    "PostgreSQL format",
			input:   "2025-11-04 17:00:00",
			wantErr: false,
		},
		{
			name:    "PostgreSQL format with timezone",
			input:   "2025-11-04 17:00:00+00",
			wantErr: false,
		},
		{
			name:    "PostgreSQL format with timezone offset",
			input:   "2025-11-04 17:00:00-07:00",
			wantErr: false,
		},
		{
			name:    "ISO format without timezone",
			input:   "2025-11-04T17:00:00",
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   "invalid-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlexibleTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlexibleTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.IsZero() {
				t.Errorf("ParseFlexibleTime() returned zero time for valid input %q", tt.input)
			}
		})
	}
}

func TestFlexibleTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Date only in JSON",
			input:   `"2025-11-01"`,
			wantErr: false,
		},
		{
			name:    "RFC3339 in JSON",
			input:   `"2006-01-02T15:04:05Z"`,
			wantErr: false,
		},
		{
			name:    "PostgreSQL format in JSON",
			input:   `"2025-11-04 17:00:00+00"`,
			wantErr: false,
		},
		{
			name:    "Null value",
			input:   `null`,
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   `""`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FlexibleTime
			err := ft.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("FlexibleTime.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlexibleTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		time    time.Time
		want    string
		wantErr bool
	}{
		{
			name:    "Valid time",
			time:    time.Date(2025, 11, 1, 12, 0, 0, 0, time.UTC),
			want:    `"2025-11-01T12:00:00Z"`,
			wantErr: false,
		},
		{
			name:    "Zero time",
			time:    time.Time{},
			want:    `null`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := FlexibleTime{Time: tt.time}
			got, err := ft.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("FlexibleTime.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("FlexibleTime.MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
