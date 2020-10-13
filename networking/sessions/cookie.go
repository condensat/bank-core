// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"context"
	"net/http"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/networking"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
	"github.com/sirupsen/logrus"
)

const (
	SessionDuration = 3 * time.Minute
)

// SessionArgs holds SessionID for operation requests and repls
type SessionArgs struct {
	SessionID string `json:"-"` // SessionID is transmit to client via cookie
}

// SessionReply holds session informations for operation replies
type SessionReply struct {
	SessionArgs
	Status     string `json:"status"`
	ValidUntil int64  `json:"valid_until"`
}

type CookieCodec struct {
	ctx    context.Context
	codec  *json.Codec
	domain string
}

func NewCookieCodec(ctx context.Context) *CookieCodec {
	return &CookieCodec{
		ctx:    ctx,
		codec:  json.NewCodec(),
		domain: appcontext.Domain(ctx),
	}
}

func (p *CookieCodec) NewRequest(r *http.Request) rpc.CodecRequest {
	return &CookieCodecRequest{
		ctx:     p.ctx,
		request: p.codec.NewRequest(r),
		domain:  p.domain,
	}
}

type CookieCodecRequest struct {
	ctx     context.Context
	request rpc.CodecRequest
	domain  string
}

func (p *CookieCodecRequest) Method() (string, error) {
	return p.request.Method()
}

func (p *CookieCodecRequest) ReadRequest(args interface{}) error {
	return p.request.ReadRequest(args)
}

func (p *CookieCodecRequest) WriteResponse(w http.ResponseWriter, args interface{}) {
	if args == nil {
		return
	}

	switch reply := args.(type) {
	case *SessionReply:
		setSessionCookie(p.domain, w, reply)

	default:
		log := logger.Logger(p.ctx).WithField("Method", "CookieCodecRequest.WriteResponse")
		log.Debug("Unknwon Reply")
	}

	// forward to request
	p.request.WriteResponse(w, args)
}

func (p *CookieCodecRequest) WriteError(w http.ResponseWriter, status int, err error) {
	p.request.WriteError(w, status, err)
}

func OpenUserSession(ctx context.Context, session *Session, r *http.Request, userID uint64) (SessionReply, error) {
	log := logger.Logger(ctx).WithField("Method", "OpenUserSession")

	// check rate limit
	openSessionAllowed := OpenSessionAllowed(ctx, userID)
	if !openSessionAllowed {
		log.WithError(ErrTooManyOpenSession).
			Warning("Session open failed")
		return SessionReply{}, ErrTooManyOpenSession
	}

	remoteAddr := networking.RequesterIP(r)
	sessionID, err := session.CreateSession(ctx, userID, remoteAddr, SessionDuration)
	if err != nil {
		return SessionReply{}, err
	}

	reply := SessionReply{
		SessionArgs: SessionArgs{
			SessionID: string(sessionID),
		},
		Status:     "open",
		ValidUntil: makeTimestampMillis(time.Now().UTC().Add(SessionDuration)),
	}

	log.WithFields(logrus.Fields{
		"Status":     reply.Status,
		"ValidUntil": fromTimestampMillis(reply.ValidUntil),
	}).Info("Session opened")

	return reply, nil
}

func CreateSessionWithCookie(ctx context.Context, r *http.Request, w http.ResponseWriter, userID uint64) error {
	log := logger.Logger(ctx).WithField("Method", "CreateSessionWithCookie")
	// Retrieve context values
	session, err := ContextSession(ctx)
	if err != nil {
		log.WithError(err).
			Warning("Session open failed")
		return ErrInternalError
	}

	reply, err := OpenUserSession(ctx, session, r, userID)
	if err != nil {
		log.WithError(err).
			Error("OpenUserSession failed")
	}

	setSessionCookie(appcontext.Domain(ctx), w, &reply)

	return nil
}

func GetSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return ""
	}

	return cookie.Value
}

func setSessionCookie(domain string, w http.ResponseWriter, reply *SessionReply) {
	expires := fromTimestampMillis(reply.ValidUntil)
	var maxAge int
	if expires.After(time.Now()) {
		maxAge = int(time.Until(expires).Seconds())
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "sessionId",
		Value:   reply.SessionID,
		Path:    "/api/v1",
		Domain:  domain,
		MaxAge:  maxAge,
		Expires: expires,

		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
