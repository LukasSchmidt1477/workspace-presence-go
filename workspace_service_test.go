package main

import "testing"

func TestPublishIfActive(t *testing.T) {
	tests := []struct {
		name    string
		active  bool
		wantErr bool
	}{
		{"active account publishes", true, false},
		{"suspended account is silent", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldPublish(Account{ID: "acct-1", Active: tt.active})
			if got == tt.wantErr {
				t.Fatalf("shouldPublish = %v, want %v", got, !tt.wantErr)
			}
		})
	}
}
