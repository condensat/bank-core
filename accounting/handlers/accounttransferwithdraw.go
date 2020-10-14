// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/cache"
	"github.com/condensat/bank-core/logger"
	"github.com/condensat/bank-core/messaging"

	"github.com/condensat/bank-core/accounting/common"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"
	"github.com/condensat/bank-core/database/query"

	"github.com/sirupsen/logrus"
)

const (
	BankWitdrawAccountName = model.AccountName("withdraw")
)

func AccountTransferWithdraw(ctx context.Context, withdraw common.AccountTransferWithdraw) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountTransferWithdraw")
	db := appcontext.Database(ctx)

	bankAccountID, err := getBankWithdrawAccount(ctx, withdraw.Source.Currency)
	if err != nil {
		log.WithError(err).
			Error("Invalid BankAccount")
		return common.AccountTransfer{}, query.ErrInvalidAccountID
	}

	log = log.WithFields(logrus.Fields{
		"BankAccountId": bankAccountID,
		"Currency":      withdraw.Source.Currency,
	})

	// get ticker precision to convert back in BTC precision (for RPC)
	tickerPrecision := -1 // no ticker precison if not crypto
	currency, err := query.GetCurrencyByName(db, model.CurrencyName(withdraw.Source.Currency))
	if err != nil {
		return common.AccountTransfer{}, err
	}
	asset, _ := query.GetAssetByCurrencyName(db, currency.Name)

	isAsset := currency.IsCrypto() && currency.GetType() == 2 && asset.ID > 0
	if currency.IsCrypto() {
		tickerPrecision = 8 // BTC precision
	}
	if isAsset {
		tickerPrecision = 0
		if assetInfo, err := query.GetAssetInfo(db, asset.ID); err == nil {
			tickerPrecision = int(assetInfo.Precision)
		}

		if currency.Name == "LBTC" {
			tickerPrecision = 8 // BTC precision
		}
	}

	// convert amount in BTC precision
	amount := convertAssetAmountToBitcoin(withdraw.Source.Amount, tickerPrecision)
	if amount <= 0.0 {
		return common.AccountTransfer{}, query.ErrInvalidWithdrawAmount
	}

	log.WithFields(logrus.Fields{
		"IsAsset":         isAsset,
		"Asset":           asset,
		"Currency":        withdraw.Source.Currency,
		"CurrencyInfo":    currency,
		"BitcoinAmount":   amount,
		"TickerPrecision": tickerPrecision,
		"AssetAmount":     withdraw.Source.Amount,
	}).Debug("Asset to Bitcoin precision")

	batchMode := model.BatchModeNormal
	if len(withdraw.BatchMode) > 0 {
		batchMode = model.BatchMode(withdraw.BatchMode)
	}

	var result common.AccountTransfer
	// Database Query
	err = db.Transaction(func(db database.Context) error {

		// Create Witdraw for batch
		w, err := query.AddWithdraw(db,
			model.AccountID(withdraw.Source.AccountID),
			model.AccountID(bankAccountID),
			model.Float(amount), batchMode,
			"{}",
		)
		if err != nil {
			log.WithError(err).
				Error("AddWithdraw failed")
			return err
		}
		_, err = query.AddWithdrawInfo(db, w.ID, model.WithdrawStatusCreated, "{}")
		if err != nil {
			log.WithError(err).
				Error("AddWithdrawInfo failed")
			return err
		}

		wt := model.FromOnChainData(w.ID, withdraw.Crypto.Chain, model.WithdrawTargetOnChainData{
			WithdrawTargetCryptoData: model.WithdrawTargetCryptoData{
				PublicKey: withdraw.Crypto.PublicKey,
			},
		})

		_, err = query.AddWithdrawTarget(db, w.ID, wt.Type, wt.Data)
		if err != nil {
			log.WithError(err).
				Error("AddWithdrawTarget failed")
			return err
		}

		referenceID := uint64(w.ID)

		currency, err := query.GetCurrencyByName(db, model.CurrencyName(withdraw.Source.Currency))
		if err != nil {
			log.WithError(err).
				Error("GetCurrencyByName failed")
			return err
		}

		// get fee informations
		isAsset := currency.IsCrypto() && currency.GetType() == 2
		feeCurrencyName := getFeeCurrency(string(currency.Name), isAsset)

		feeBankAccountID, err := getBankWithdrawAccount(ctx, feeCurrencyName)
		if err != nil {
			log.WithError(err).
				Error("Invalid Fee BankAccount")
			return query.ErrInvalidAccountID
		}

		feeInfo, err := query.GetFeeInfo(db, model.CurrencyName(feeCurrencyName))
		if err != nil {
			log.WithError(err).
				Error("GetFeeInfo failed")
			return err
		}
		if !feeInfo.IsValid() {
			log.Error("Invalid FeeInfo")
			return errors.New("Invalid FeeInfo")
		}

		feeAmount := feeInfo.Compute(model.Float(amount))
		feeUserAccount := withdraw.Source.AccountID
		if feeCurrencyName != withdraw.Source.Currency {
			// if fee is not in the same currency (ie asset without quote)
			// take the minimum fee of the currency fee
			feeAmount = feeInfo.Minimum

			// get feeUserAccoiunt from user
			userAccount, err := query.GetAccountByID(db, model.AccountID(withdraw.Source.AccountID))
			if err != nil {
				log.WithError(err).
					Error("GetAccountByID failed")
				return err
			}
			// get user account for currency fee
			accounts, err := query.GetAccountsByUserAndCurrencyAndName(db, userAccount.UserID, model.CurrencyName(feeCurrencyName), query.AccountNameDefault)
			if err != nil {
				return errors.New("GetAccountsByUserAndCurrencyAndName failed")
			}
			if len(accounts) == 0 {
				return query.ErrAccountNotFound
			}
			// use first default account
			account := accounts[0]
			feeUserAccount = uint64(account.ID)
		}

		// Transfert fees from account to bankAccount
		timestamp := time.Now()
		result, err = AccountTransferWithDatabase(ctx, db, common.AccountTransfer{
			Source: common.AccountEntry{
				AccountID: feeUserAccount,

				OperationType:    string(model.OperationTypeTransferFee),
				SynchroneousType: "sync",
				ReferenceID:      referenceID,

				Timestamp: timestamp,
				Amount:    float64(-feeAmount),

				Currency: feeCurrencyName,
			},
			Destination: common.AccountEntry{
				AccountID: uint64(feeBankAccountID),

				OperationType:    string(model.OperationTypeTransferFee),
				SynchroneousType: "sync",
				ReferenceID:      referenceID,

				Timestamp: timestamp,
				Amount:    float64(feeAmount),

				Currency: feeCurrencyName,
			},
		})
		if err != nil {
			log.WithError(err).
				Error("AccountTransfer fee failed")
			return err
		}

		// Transfert amount from account to bank account
		result, err = AccountTransferWithDatabase(ctx, db, common.AccountTransfer{
			Source: withdraw.Source,
			Destination: common.AccountEntry{
				AccountID: uint64(bankAccountID),

				OperationType:    withdraw.Source.OperationType,
				SynchroneousType: "async-start",
				ReferenceID:      referenceID,

				Timestamp: time.Now(),
				Amount:    amount,

				Label: withdraw.Source.Label,

				LockAmount: amount,
				Currency:   withdraw.Source.Currency,
			},
		})
		if err != nil {
			log.WithError(err).
				Error("AccountTransfer failed")
			return err
		}

		log.Debug("AccountWithdraw created")

		return nil
	})
	if err != nil {
		return common.AccountTransfer{}, err
	}

	return result, err
}

func OnAccountTransferWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountTransferWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountTransferWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			response, err := AccountTransferWithdraw(ctx, request)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"AccountID": request.Source.AccountID,
					}).Errorf("Failed to AccountTransferWithdraw")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}

func getBankWithdrawAccount(ctx context.Context, currency string) (model.AccountID, error) {
	bankUser := common.BankUserFromContext(ctx)
	if bankUser.ID == 0 {
		return 0, query.ErrInvalidUserID
	}

	db := appcontext.Database(ctx)
	currencyName := model.CurrencyName(currency)
	if !query.AccountsExists(db, bankUser.ID, currencyName, BankWitdrawAccountName) {
		result, err := AccountCreate(ctx, uint64(bankUser.ID), common.AccountInfo{
			UserID: uint64(bankUser.ID),
			Name:   string(BankWitdrawAccountName),
			Currency: common.CurrencyInfo{
				Name: currency,
			},
		})
		if err != nil {
			return 0, err
		}

		_, err = AccountSetStatus(ctx, result.AccountID, model.AccountStatusNormal.String())
		if err != nil {
			return 0, err
		}
		return model.AccountID(result.AccountID), err
	}

	accounts, err := query.GetAccountsByUserAndCurrencyAndName(db, bankUser.ID, model.CurrencyName(currencyName), BankWitdrawAccountName)
	if err != nil {
		return 0, err
	}

	if len(accounts) == 0 {
		return 0, query.ErrAccountNotFound
	}
	account := accounts[0]
	if account.ID == 0 {
		return 0, query.ErrInvalidAccountID
	}

	return account.ID, nil
}

func getFeeCurrency(currency string, isAsset bool) string {
	if !isAsset {
		return currency
	}

	switch currency {
	case "USDt":
		fallthrough
	case "LCAD":
		return currency

	default:
		return "LBTC"
	}
}
