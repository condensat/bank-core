// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/condensat/bank-core"
	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/logger"

	"github.com/condensat/bank-core/wallet/common"
	"github.com/condensat/bank-core/wallet/ssm/commands"

	"github.com/condensat/bank-core/database"
	"github.com/condensat/bank-core/database/model"

	"github.com/sirupsen/logrus"
)

const (
	DefaultDerivationPrefix = "84h/0h"
	SsmMaxUnusedAddress     = 1000
	SsmFillBatchSize        = 16
)

type SsmInfo struct {
	Device           string
	Chain            string
	Fingerprint      string
	DerivationPrefix string
}

// SsmPool
func SsmPool(ctx context.Context, epoch time.Time, infos []SsmInfo) {
	log := logger.Logger(ctx).WithField("Method", "task.SsmPool")
	db := appcontext.Database(ctx)

	log.WithFields(logrus.Fields{
		"Epoch": epoch.Truncate(time.Millisecond),
		"Count": len(infos),
	}).Info("Maintain ssm pool addresses")

	for _, info := range infos {
		ssm := common.SsmClientFromContext(ctx, info.Device)

		unusedCount, err := database.CountSsmAddressByState(db,
			model.SsmChain(info.Chain),
			model.SsmFingerprint(info.Fingerprint),
			model.SsmAddressStatusUnused,
		)
		if err != nil {
			log.WithError(err).Error("CountSsmAddressByState failed")
			continue
		}

		// count actual ssm addresses count for chain/fingerprint
		addressCount, err := database.CountSsmAddress(db,
			model.SsmChain(info.Chain),
			model.SsmFingerprint(info.Fingerprint),
		)
		if err != nil {
			log.WithError(err).Error("CountSsmAddress failed")
			continue
		}

		log.WithFields(logrus.Fields{
			"UnusedCount":  unusedCount,
			"AddressCount": addressCount,
		}).Debug("SsmPool status")

		// Fill ssm pool
		nextUnusedCount := unusedCount + SsmFillBatchSize
		if nextUnusedCount > SsmMaxUnusedAddress {
			nextUnusedCount = SsmMaxUnusedAddress
		}

		var lockMap sync.Mutex
		type addressPath struct {
			pathSequence int
			nextPath     string
			address      common.SsmAddress
		}
		var addresses []addressPath
		var wg sync.WaitGroup
		for unusedCount < nextUnusedCount {
			wg.Add(1)
			go func(addressCount int) {
				defer wg.Done()

				pathSequence := addressCount + 1

				// create new address for next path
				// Todo: manage annual rotation for path
				derivationPrefix := info.DerivationPrefix
				if len(derivationPrefix) == 0 {
					derivationPrefix = DefaultDerivationPrefix
				}
				nextPath := fmt.Sprintf("%s/%d", derivationPrefix, pathSequence)

				ssmAddress, err := ssm.NewAddress(ctx, commands.SsmPath{
					Chain:       info.Chain,
					Fingerprint: info.Fingerprint,
					Path:        nextPath,
				})
				if err != nil {
					log.WithError(err).Error("NewAddress failed")
					return
				}
				if info.Chain != ssmAddress.Address {
					if err != nil {
						log.WithError(err).Error("Wrong ssmAddress chain")
						return
					}
				}

				lockMap.Lock()
				defer lockMap.Unlock()
				addresses = append(addresses, addressPath{pathSequence, nextPath, ssmAddress})

			}(addressCount)

			unusedCount++
			addressCount++
		}
		wg.Wait()

		// sort addresses by pathSequence
		sort.Slice(addresses, func(i, j int) bool {
			return addresses[i].pathSequence < addresses[j].pathSequence
		})

		// store new address to database within a db transaction
		err = db.Transaction(func(db bank.Database) error {

			for _, addressPath := range addresses {
				ssmAddressID, err := database.AddSsmAddress(db,
					model.SsmAddress{
						PublicAddress: model.SsmPublicAddress(addressPath.address.Address),
						ScriptPubkey:  model.SsmPubkey(addressPath.address.PubKey),
						BlindingKey:   model.SsmBlindingKey(addressPath.address.BlindingKey),
					},
					model.SsmAddressInfo{
						Chain:       model.SsmChain(info.Chain),
						Fingerprint: model.SsmFingerprint(info.Fingerprint),
						HDPath:      model.SsmHDPath(addressPath.nextPath),
					},
				)
				if err != nil {
					log.WithError(err).Error("AddSsmAddress failed")
					return err
				}
				log.
					WithField("ssmAddressID", ssmAddressID).
					Debug("New ssm address")
			}

			return nil
		})
		if err != nil {
			log.WithError(err).Error("Database Transaction failed")
			return
		}
	}
}
