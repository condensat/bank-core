// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

import (
	"context"
	"net/http"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/api/sessions"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
	"github.com/sirupsen/logrus"
)

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

func openUserSession(ctx context.Context, session *sessions.Session, r *http.Request, userID uint64) (SessionReply, error) {
	log := logger.Logger(ctx).WithField("Method", "services.openUserSession")

	// check rate limit
	openSessionAllowed := OpenSessionAllowed(ctx, userID)
	if !openSessionAllowed {
		log.WithError(ErrTooManyOpenSession).
			Warning("Session open failed")
		return SessionReply{}, ErrTooManyOpenSession
	}

	remoteAddr := RequesterIP(r)
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
	log := logger.Logger(ctx).WithField("Method", "services.CreateSessionWithCookie")
	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Warning("Session open failed")
		return ErrServiceInternalError
	}

	reply, err := openUserSession(ctx, session, r, userID)
	if err != nil {
		log.WithError(err).
			Error("openUserSession failed")
	}

	setSessionCookie(appcontext.Domain(ctx), w, &reply)

	return nil
}
