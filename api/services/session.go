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
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	"github.com/sirupsen/logrus"
)

const (
	SessionDuration = 3 * time.Minute
)

var (
	ErrInvalidCrendential    = errors.New("InvalidCredentials")
	ErrMissingCookie         = errors.New("MissingCookie")
	ErrInvalidCookie         = errors.New("ErrInvalidCookie")
	ErrSessionCreationFailed = errors.New("SessionCreationFailed")
	ErrTooManyOpenSession    = errors.New("TooManyOpenSession")
	ErrSessionExpired        = sessions.ErrSessionExpired
	ErrSessionClose          = errors.New("SessionCloseFailed")
)

// SessionService receiver
type SessionService int

// SessionArgs holds SessionID for operation requests and repls
type SessionArgs struct {
	SessionID string `json:"-"` // SessionID is transmit to client via cookie
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

func GetSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return ""
	}

	return cookie.Value
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
	userID, valid, err := database.CheckCredential(ctx, db, model.Base58(request.Login), model.Base58(request.Password))
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

	sessionReply, err := openUserSession(ctx, session, r, uint64(userID))
	if err != nil {
		log.WithError(err).
			Warning("openSession failed")
		return ErrSessionCreationFailed
	}
	*reply = sessionReply

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
	request.SessionID = GetSessionCookie(r)
	sessionID := sessions.SessionID(request.SessionID)
	remoteAddr := RequesterIP(r)
	userID, err := session.ExtendSession(ctx, remoteAddr, sessionID, SessionDuration)

	log = log.WithFields(logrus.Fields{
		"SessionID":  sessionID,
		"UserID":     userID,
		"RemoteAddr": remoteAddr,
	})
	switch err {
	case sessions.ErrRemoteAddrChanged:
		// force session close if RemoteAddr Changed
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
		}).Info("Session closed (forced)")

		return sessions.ErrRemoteAddrChanged
	}
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
	request.SessionID = GetSessionCookie(r)
	sessionID := sessions.SessionID(request.SessionID)
	userID := session.UserSession(ctx, sessionID)
	log = log.WithFields(logrus.Fields{
		"SessionID": sessionID,
		"UserID":    userID,
	})
	err = session.InvalidateSession(ctx, sessionID)
	if err != nil {
		log.WithError(err).
			WithField("SessionID", sessionID).
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
