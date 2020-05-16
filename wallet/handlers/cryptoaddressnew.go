// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/messaging"

	"github.com/sirupsen/logrus"
)

func CryptoAddressNewDeposit(ctx context.Context, address common.CryptoAddress) (common.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.CryptoAddressNewDeposit")
	var result common.CryptoAddress

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return result, ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"Chain":     address.Chain,
		"AccountID": address.AccountID,
	})

	if len(address.Chain) == 0 {
		log.WithError(ErrInvalidChain).
			Debug("AddressNext Failed")
		return result, ErrInvalidChain
	}
	if address.AccountID == 0 {
		log.WithError(ErrInvalidAccountID).
			Debug("AddressNext Failed")
		return result, ErrInvalidAccountID
	}

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {

		chain := model.String(address.Chain)
		accountID := model.AccountID(address.AccountID)

		addr, err := txNewCryptoAddress(ctx, db, chainHandler, chain, accountID)
		if err != nil {
			log.WithError(err).
				Error("Failed to txNewCryptoAddress")
			return err
		}

		result = common.CryptoAddress{
			CryptoAddressID: uint64(addr.ID),
			Chain:           string(addr.Chain),
			AccountID:       uint64(addr.AccountID),
			PublicAddress:   string(addr.PublicAddress),
			Unconfidential:  string(addr.Unconfidential),
		}

		return nil
	})

	if err == nil {
		log.WithField("PublicAddress", result.PublicAddress).
			Debug("Next deposit publicAddress")
	}

	return result, err
}

func OnCryptoAddressNewDeposit(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnCryptoAddressNewDeposit")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CryptoAddress
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Chain":     request.Chain,
				"AccountID": request.AccountID,
			})

			newDeposit, err := CryptoAddressNewDeposit(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoAddressNewsDeposit")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"PublicAddress": newDeposit.PublicAddress,
			})

			log.Info("New Deposit Address")

			// create & return response
			return &common.CryptoAddress{
				CryptoAddressID: newDeposit.CryptoAddressID,
				Chain:           request.Chain,
				AccountID:       request.AccountID,
				PublicAddress:   newDeposit.PublicAddress,
				Unconfidential:  newDeposit.Unconfidential,
			}, nil
		})
}

func txNewCryptoAddress(ctx context.Context, db bank.Database, chainHandler ChainHandler, chain model.String, accountID model.AccountID) (model.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.txNewCryptoAddress")
	account := genAccountLabelFromAccountID(accountID)
	publicAddress, err := chainHandler.GetNewAddress(ctx, string(chain), account)
	if err != nil {
		log.WithError(err).
			Error("Failed to GetNewAddress")
		return model.CryptoAddress{}, ErrGenAddress
	}

	info, err := chainHandler.GetAddressInfo(ctx, string(chain), publicAddress)
	if err != nil {
		log.WithError(err).
			Error("Failed to GetAddressInfo")
		return model.CryptoAddress{}, ErrGenAddress
	}

	addr, err := database.AddOrUpdateCryptoAddress(db, model.CryptoAddress{
		Chain:          chain,
		AccountID:      accountID,
		PublicAddress:  model.String(publicAddress),
		Unconfidential: model.String(info.Unconfidential),
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to AddOrUpdateCryptoAddress")
		return model.CryptoAddress{}, err
	}

	return addr, nil
}
