// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/common"

	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/messaging"

	"github.com/shengdoushi/base58"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidChain     = errors.New("Invalid Chain")
	ErrInvalidAccountID = errors.New("Invalid AccountID")
	ErrGenAddress       = errors.New("Gen Address Error")
)

func CryptoAddressNextDeposit(ctx context.Context, address common.CryptoAddress) (common.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.CryptoAddressNextDeposit")
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

		addresses, err := database.AllUnusedAccountCryptoAddresses(db, accountID)
		if err != nil {
			log.WithError(err).
				Error("Failed to AllUnusedAccountCryptoAddresses")
			return err
		}

		// return last unised address
		if len(addresses) > 0 {
			addr := addresses[len(addresses)-1]

			log.Debug("Found unused deposit address")

			result = common.CryptoAddress{
				Chain:         string(addr.Chain),
				AccountID:     uint64(addr.AccountID),
				PublicAddress: string(addr.PublicAddress),
			}
			return nil
		}

		account := genAccountLabelFromAccountID(accountID)
		publicAddress, err := chainHandler.GetNewAddress(ctx, string(chain), account)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetNewAddress")
			return ErrGenAddress
		}

		addr, err := database.AddOrUpdateCryptoAddress(db, model.CryptoAddress{
			Chain:         chain,
			AccountID:     accountID,
			PublicAddress: model.String(publicAddress),
		})
		if err != nil {
			log.WithError(err).
				Error("Failed to AddOrUpdateCryptoAddress")
			return err
		}

		result = common.CryptoAddress{
			Chain:         string(addr.Chain),
			AccountID:     uint64(addr.AccountID),
			PublicAddress: string(addr.PublicAddress),
		}

		return nil
	})

	if err == nil {
		log.WithField("PublicAddress", result.PublicAddress).
			Debug("Next deposit publicAddress")
	}

	return result, err
}

func OnCryptoAddressNextDeposit(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnCryptoAddressNextDeposit")
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

			nextDeposit, err := CryptoAddressNextDeposit(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoAddressNextDeposit")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"PublicAddress": nextDeposit.PublicAddress,
			})

			log.Info("Next Deposit Address")

			// create & return response
			return &common.CryptoAddress{
				Chain:         request.Chain,
				AccountID:     request.AccountID,
				PublicAddress: nextDeposit.PublicAddress,
			}, nil
		})
}

func genAccountLabelFromAccountID(accountID model.AccountID) string {
	// create account label from accountID
	accountHash := fmt.Sprintf("bank.account:%d", accountID)
	return base58.Encode([]byte(accountHash), base58.BitcoinAlphabet)
}
