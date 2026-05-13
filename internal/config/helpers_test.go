package config

import (
	"errors"
	"testing"
)

func TestParseConnectionMapping(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    connectionMapping
		wantErr error
	}{
		{
			name:  "plain tcp default",
			input: "5432:db.internal:5432",
			want:  connectionMapping{SourcePort: 5432, TargetAddr: "db.internal", TargetPort: 5432, HTTPS: false},
		},
		{
			name:  "https prefix lowercase",
			input: "https:443:web.internal:8080",
			want:  connectionMapping{SourcePort: 443, TargetAddr: "web.internal", TargetPort: 8080, HTTPS: true},
		},
		{
			name:  "https prefix uppercase",
			input: "HTTPS:443:web.internal:8080",
			want:  connectionMapping{SourcePort: 443, TargetAddr: "web.internal", TargetPort: 8080, HTTPS: true},
		},
		{
			name:  "https prefix mixed case",
			input: "HtTpS:443:web.internal:8080",
			want:  connectionMapping{SourcePort: 443, TargetAddr: "web.internal", TargetPort: 8080, HTTPS: true},
		},
		{
			name:  "short hostname",
			input: "https:443:web:8080",
			want:  connectionMapping{SourcePort: 443, TargetAddr: "web", TargetPort: 8080, HTTPS: true},
		},
		{
			name:  "unrecognized protocol",
			input: "ftp:443:web.internal:8080",
			want:  connectionMapping{SourcePort: 443, TargetAddr: "web.internal", TargetPort: 8080, HTTPS: false},
		},
		{
			name:  "ipv4 target",
			input: "5432:10.0.0.1:5432",
			want:  connectionMapping{SourcePort: 5432, TargetAddr: "10.0.0.1", TargetPort: 5432, HTTPS: false},
		},
		{
			name:    "missing target entirely",
			input:   "5432",
			wantErr: errInvalidConnectionMapping,
		},
		{
			name:    "missing target port",
			input:   "5432:db.internal",
			wantErr: errInvalidConnectionMapping,
		},
		{
			name:    "missing target port after https prefix",
			input:   "https:443:web.internal",
			wantErr: errInvalidConnectionMapping,
		},
		{
			name:    "non-numeric source port",
			input:   "abc:db.internal:5432",
			wantErr: errInvalidSourcePort,
		},
		{
			name:    "non-numeric source port after https prefix",
			input:   "https:abc:web.internal:8080",
			wantErr: errInvalidSourcePort,
		},
		{
			name:    "non-numeric target port",
			input:   "5432:db.internal:abc",
			wantErr: errInvalidTargetPort,
		},
		{
			name:    "empty value",
			input:   "",
			wantErr: errInvalidConnectionMapping,
		},
		{
			name:    "https in non-prefix position rejected as too many parts",
			input:   "443:https:host:8080",
			wantErr: errInvalidConnectionMapping,
		},
		{
			name:    "extra segment between source and host",
			input:   "8080:http:web.internal:5432",
			wantErr: errInvalidConnectionMapping,
		},
		{
			name:    "trailing colon",
			input:   "5432:host:5432:",
			wantErr: errInvalidConnectionMapping,
		},
		{
			name:  "max valid ports",
			input: "65535:host:65535",
			want:  connectionMapping{SourcePort: 65535, TargetAddr: "host", TargetPort: 65535, HTTPS: false},
		},
		{
			name:  "min valid ports",
			input: "1:host:1",
			want:  connectionMapping{SourcePort: 1, TargetAddr: "host", TargetPort: 1, HTTPS: false},
		},
		{
			name:    "source port zero",
			input:   "0:host:5432",
			wantErr: errPortOutOfRange,
		},
		{
			name:    "source port negative",
			input:   "-1:host:5432",
			wantErr: errPortOutOfRange,
		},
		{
			name:    "source port above 65535",
			input:   "65536:host:5432",
			wantErr: errPortOutOfRange,
		},
		{
			name:    "target port zero",
			input:   "5432:host:0",
			wantErr: errPortOutOfRange,
		},
		{
			name:    "target port above 65535",
			input:   "5432:host:65536",
			wantErr: errPortOutOfRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConnectionMapping(tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
