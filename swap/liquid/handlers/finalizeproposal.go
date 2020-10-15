// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/swap/liquid/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/utils/shellexec"

	"github.com/sirupsen/logrus"
)

func FinalizeSwapProposal(ctx context.Context, swapID uint64, payload common.Payload) (common.SwapProposal, error) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.handler.FinalizeSwapProposal")

	log = log.WithField("SwapID", swapID)

	if !payload.Valid() {
		log.WithError(common.ErrInvalidPayload).
			WithField("Payload", payload).
			Error("Invalid Payload")
		return common.SwapProposal{}, common.ErrInvalidPayload
	}

	result := common.SwapProposal{
		Timestamp: time.Now().UTC().Truncate(time.Millisecond),
		SwapID:    swapID,
	}

	ShellExecLock.Lock()
	defer ShellExecLock.Unlock()

	out, err := shellexec.Execute(ctx, LiquidSwapFinalize(payload))
	if len(out.Stdout) == 0 && err == nil {
		err = errors.New("No Output")
	}
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"Stdout": out.Stdout,
				"Stderr": out.Stderr,
				"Code":   out.Code,
			}).
			Error("out")
		return result, err
	}

	result.Payload = common.Payload(out.Stdout)

	if !result.Payload.Valid() {
		log.WithError(common.ErrInvalidPayload).
			WithField("Payload", result.Payload).
			Error("Invalid Payload")
		return common.SwapProposal{}, common.ErrInvalidPayload
	}

	log.WithField("Result", result).
		Debug("Finalize Swap Proposal")

	return result, nil
}

func OnFinalizeSwapProposal(ctx context.Context, subject string, message *messaging.Message) (*messaging.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.handler.OnFinalizeSwapProposal")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.SwapProposal
	return messaging.HandleRequest(ctx, appcontext.AppName(ctx), message, &request,
		func(ctx context.Context, _ messaging.BankObject) (messaging.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"SwapID": request.SwapID,
			})

			response, err := FinalizeSwapProposal(ctx, request.SwapID, request.Payload)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FinalizeSwapProposal")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &response, nil
		})
}
