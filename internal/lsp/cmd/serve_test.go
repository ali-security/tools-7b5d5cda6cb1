// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/internal/jsonrpc2"
	"golang.org/x/tools/internal/tool"
)

func TestListenParsing(t *testing.T) {
	tests := []struct {
		input, wantNetwork, wantAddr string
	}{
		{"127.0.0.1:0", "tcp", "127.0.0.1:0"},
		{"unix;/tmp/sock", "unix", "/tmp/sock"},
		{"auto", "auto", ""},
		{"auto;foo", "auto", "foo"},
	}

	for _, test := range tests {
		gotNetwork, gotAddr := parseAddr(test.input)
		if gotNetwork != test.wantNetwork {
			t.Errorf("network = %q, want %q", gotNetwork, test.wantNetwork)
		}
		if gotAddr != test.wantAddr {
			t.Errorf("addr = %q, want %q", gotAddr, test.wantAddr)
		}
	}
}

// TestListenRejectsImplicitAllInterfaces checks that a -listen address with
// an empty host (such as ":0" or "tcp;:0"), which would silently bind every
// network interface, is rejected before any listener is opened.
func TestListenRejectsImplicitAllInterfaces(t *testing.T) {
	for _, listen := range []string{":0", "tcp;:0"} {
		app := New("gopls-test", "", nil, nil)
		app.Serve.app = app
		// The idle timeout bounds Run if the address is (wrongly) accepted.
		err := tool.Run(context.Background(), &app.Serve, []string{"-listen=" + listen, "-listen.timeout=100ms"})
		if err == nil || !strings.Contains(err.Error(), "implicitly binds all network interfaces") {
			t.Errorf("serve -listen=%s: got error %v, want rejection of implicit all-interface bind", listen, err)
		}
	}

	// An explicit loopback host is still accepted.
	app := New("gopls-test", "", nil, nil)
	app.Serve.app = app
	err := tool.Run(context.Background(), &app.Serve, []string{"-listen=localhost:0", "-listen.timeout=100ms"})
	if err != jsonrpc2.ErrIdleTimeout {
		t.Errorf("serve -listen=localhost:0: got error %v, want %v", err, jsonrpc2.ErrIdleTimeout)
	}
}

// TestServeHasNoPortFlag checks that the serve command no longer offers the
// -port flag, which listened on all network interfaces.
func TestServeHasNoPortFlag(t *testing.T) {
	typ := reflect.TypeOf(Serve{})
	for i := 0; i < typ.NumField(); i++ {
		if f := typ.Field(i); f.Tag.Get("flag") == "port" {
			t.Errorf("Serve.%s registers the -port flag, which binds all network interfaces", f.Name)
		}
	}
}
