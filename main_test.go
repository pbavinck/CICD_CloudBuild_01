package main

import (
	"testing"
)

func Test_startupMessages(t *testing.T) {
	type args struct {
		staticRoot string
		port       string
	}
	tests := []struct {
		name        string
		args        args
		wantMessage string
	}{
		{
			name: "Default message",
			args: args{
				staticRoot: "/somestaticroot",
				port:       "8080",
			},
			wantMessage: "Demo webserver\nServing content from /somestaticroot on port: 8080\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMessage := startupMessages(tt.args.staticRoot, tt.args.port); gotMessage != tt.wantMessage {
				t.Errorf("startupMessages() = %v, want %v", gotMessage, tt.wantMessage)
			}
		})
	}
}
