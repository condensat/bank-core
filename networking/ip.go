// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package networking

import (
	"net/http"
)

func RequesterIP(r *http.Request) string {
	// Header added by reverse proxy
	const cfConnectingIP = "CF-Connecting-IP" // CloudFlare
	const xRealIP = "X-Real-Ip"               // traefik
	const xForwardedFor = "X-Forwarded-For"   // Generic proxy

	// Priority order
	if ips, ok := r.Header[cfConnectingIP]; ok && len(ips) > 0 {
		return ips[0]
	} else if ips, ok := r.Header[xRealIP]; ok && len(ips) > 0 {
		return ips[0]
	} else if ips, ok := r.Header[xForwardedFor]; ok && len(ips) > 0 {
		return ips[0]
	} else {
		return r.RemoteAddr // fallback with RemoteAddr
	}
}
