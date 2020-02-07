// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// services package is au json-rpc service for session management.
// client can perform three operations on session:
// - Open to initiate a session with valid credentials
// - Renew to extends the validty period from a valid session
// - Close to invalidate a session
package services

import (
	"errors"
	"net/http"
	"time"

	"github.com/condensat/bank-core/api/sessions"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

const (
	SessionDuration = 3 * time.Minute
)

var (
	ErrInvalidCrendential    = errors.New("InvalidCredentials")
	ErrSessionCreationFailed = errors.New("SessionCreationFailed")
	ErrTooManyOpenSession    = errors.New("TooManyOpenSession")
	ErrSessionExpired        = sessions.ErrSessionExpired
	ErrSessionClose          = errors.New("SessionCloseFailed")
)

// SessionService receiver
type SessionService int

// SessionArgs holds SessionID for operation requests and repls
type SessionArgs struct {
	SessionID string `json:"sessionId"`
}

// SessionOpenRequest holds args for open requests
type SessionOpenRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	OTP      string `json:"otp,omitempty"`
}

// SessionReply holds session informations for operation replies
type SessionReply struct {
	SessionArgs
	Status     string `json:"status"`
	ValidUntil int64  `json:"valid_until"`
}

// Open operation perform check regarding credentials and return a sessionID
// session has a status [open, close] and a validation period
func (p *SessionService) Open(r *http.Request, request *SessionOpenRequest, reply *SessionReply) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.SessionService.Open")
	log = GetServiceRequestLog(log, r, "Session", "Open")

	// Retrieve context values
	db, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Warning("Session open failed")
		return ErrServiceInternalError
	}

	// Check credentials
	userID, valid, err := database.CheckCredential(ctx, db, request.Login, request.Password)
	if err != nil {
		log.WithError(err).
			Warning("Session open failed")
		return ErrInvalidCrendential
	}
	log = log.WithField("UserID", userID)
	if !valid {
		log.WithError(ErrInvalidCrendential).
			Warning("Session open failed")
		return ErrInvalidCrendential
	}

	// check rate limit
	openSessionAllowed := OpenSessionAllowed(ctx, userID)
	if !openSessionAllowed {
		log.WithError(ErrTooManyOpenSession).
			Warning("Session open failed")
		return ErrTooManyOpenSession
	}

	// Create session
	sessionID, err := session.CreateSession(ctx, userID, SessionDuration)
	if err != nil {
		log.WithError(err).
			Warning("Session open failed")
		return ErrSessionCreationFailed
	}
	log = log.WithField("SessionID", sessionID)

	// Reply
	*reply = SessionReply{
		SessionArgs: SessionArgs{
			SessionID: string(sessionID),
		},
		Status:     "open",
		ValidUntil: makeTimestampMillis(time.Now().UTC().Add(time.Minute)),
	}

	log.WithFields(logrus.Fields{
		"Status":     reply.Status,
		"ValidUntil": fromTimestampMillis(reply.ValidUntil),
	}).Info("Session opened")

	return nil
}

// Open operation perform check the session validity and extends the validation period
func (p *SessionService) Renew(r *http.Request, request *SessionArgs, reply *SessionReply) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.SessionService.Renew")
	log = GetServiceRequestLog(log, r, "Session", "Renew")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Warning("Session renew failed")
		return ErrServiceInternalError
	}

	// Extend session
	sessionID := sessions.SessionID(request.SessionID)
	userID, err := session.ExtendSession(ctx, sessionID, SessionDuration)
	log = log.WithFields(logrus.Fields{
		"SessionID": sessionID,
		"UserID":    userID,
	})
	if err != nil {
		log.WithError(err).
			Warning("Session renew failed")
		return ErrSessionExpired
	}

	// Reply
	*reply = SessionReply{
		SessionArgs: SessionArgs{
			SessionID: request.SessionID,
		},
		Status:     "open",
		ValidUntil: makeTimestampMillis(time.Now().UTC().Add(SessionDuration)),
	}

	log.WithFields(logrus.Fields{
		"Status":     reply.Status,
		"ValidUntil": fromTimestampMillis(reply.ValidUntil),
	}).Info("Session renewed")

	return nil
}

// Close operation close the session and set status to closed
func (p *SessionService) Close(r *http.Request, request *SessionArgs, reply *SessionReply) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.SessionService.Close")
	log = GetServiceRequestLog(log, r, "Session", "Close")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Invalidate session
	sessionID := sessions.SessionID(request.SessionID)
	userID := session.UserSession(ctx, sessionID)
	log = log.WithFields(logrus.Fields{
		"SessionID": sessionID,
		"UserID":    userID,
	})
	err = session.InvalidateSession(ctx, sessionID)
	if err != nil {
		log.WithError(err).
			Warning("Session close failed")
		return ErrSessionClose
	}

	// Reply
	*reply = SessionReply{
		SessionArgs: SessionArgs{
			SessionID: request.SessionID,
		},
		Status:     "closed",
		ValidUntil: makeTimestampMillis(time.Now().UTC()),
	}

	log.WithFields(logrus.Fields{
		"Status":     reply.Status,
		"ValidUntil": fromTimestampMillis(reply.ValidUntil),
	}).Info("Session closed")

	return nil
}
