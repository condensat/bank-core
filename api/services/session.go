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
	"context"
	"net/http"
	"time"

	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/networking"
	"github.com/condensat/bank-core/networking/sessions"

	"github.com/sirupsen/logrus"
)

// SessionService receiver
type SessionService struct {
	checkCredential CheckCredentialHandler
}

type CheckCredentialHandler func(ctx context.Context, login, password string) (uint64, bool, error)

func NewSessionService(checkCredential CheckCredentialHandler) SessionService {
	return SessionService{
		checkCredential: checkCredential,
	}
}

// SessionOpenRequest holds args for open requests
type SessionOpenRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	OTP      string `json:"otp,omitempty"`
}

// Open operation perform check regarding credentials and return a sessionID
// session has a status [open, close] and a validation period
func (p *SessionService) Open(r *http.Request, request *SessionOpenRequest, reply *sessions.SessionReply) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.SessionService.Open")
	log = networking.GetServiceRequestLog(log, r, "Session", "Open")

	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Warning("Session open failed")
		return ErrServiceInternalError
	}
	if p.checkCredential == nil {
		log.WithError(err).
			Error("checkCredential")
		return ErrServiceInternalError
	}

	// Check credentials
	userID, valid, err := p.checkCredential(ctx, request.Login, request.Password)
	if err != nil {
		log.WithError(err).
			Warning("Session open failed")
		return sessions.ErrInvalidCrendential
	}
	log = log.WithField("UserID", userID)
	if !valid {
		log.WithError(sessions.ErrInvalidCrendential).
			Warning("Session open failed")
		return sessions.ErrInvalidCrendential
	}

	sessionReply, err := sessions.OpenUserSession(ctx, session, r, userID)
	if err != nil {
		log.WithError(err).
			Warning("openSession failed")
		return sessions.ErrSessionCreationFailed
	}
	*reply = sessionReply

	return nil
}

// Open operation perform check the session validity and extends the validation period
func (p *SessionService) Renew(r *http.Request, request *sessions.SessionArgs, reply *sessions.SessionReply) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.SessionService.Renew")
	log = networking.GetServiceRequestLog(log, r, "Session", "Renew")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Warning("Session renew failed")
		return ErrServiceInternalError
	}

	// Extend session
	request.SessionID = sessions.GetSessionCookie(r)
	sessionID := sessions.SessionID(request.SessionID)
	remoteAddr := networking.RequesterIP(r)
	userID, err := session.ExtendSession(ctx, remoteAddr, sessionID, sessions.SessionDuration)

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
			return sessions.ErrSessionClose
		}

		// Reply
		*reply = sessions.SessionReply{
			SessionArgs: sessions.SessionArgs{
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
		return sessions.ErrSessionExpired
	}

	// Reply
	*reply = sessions.SessionReply{
		SessionArgs: sessions.SessionArgs{
			SessionID: request.SessionID,
		},
		Status:     "open",
		ValidUntil: makeTimestampMillis(time.Now().UTC().Add(sessions.SessionDuration)),
	}

	log.WithFields(logrus.Fields{
		"Status":     reply.Status,
		"ValidUntil": fromTimestampMillis(reply.ValidUntil),
	}).Info("Session renewed")

	return nil
}

// Close operation close the session and set status to closed
func (p *SessionService) Close(r *http.Request, request *sessions.SessionArgs, reply *sessions.SessionReply) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.SessionService.Close")
	log = networking.GetServiceRequestLog(log, r, "Session", "Close")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Invalidate session
	request.SessionID = sessions.GetSessionCookie(r)
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
		return sessions.ErrSessionClose
	}

	// Reply
	*reply = sessions.SessionReply{
		SessionArgs: sessions.SessionArgs{
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
