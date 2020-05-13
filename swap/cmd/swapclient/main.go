// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"log"

	"github.com/condensat/bank-core/appcontext"
	"github.com/condensat/bank-core/messaging"
	"github.com/condensat/bank-core/swap/liquid/client"
	"github.com/condensat/bank-core/swap/liquid/common"
)

func main() {
	ctx := context.Background()
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, messaging.DefaultOptions()))

	address := common.ConfidentialAddress("lq1qqv8ymngmdp5yj2jukdd78ujm92m0wjvk7yxplra2haf5dnuzsutz96dvvqscm0raftaljf9p30wg4sd2alht5epuyn2fe7vn6")
	proposal, err := client.CreateSwapProposal(ctx, 42,
		address, common.ProposalInfo{
			ProposerAsset:  "ce091c998b83c78bb71a632313ba3760f1763d9cfcffae02258ffa9865a37bd2", // USDt
			ProposerAmount: 1000 / 100000000.0,
			ReceiverAsset:  "0e99c1a6da379d1f4151fb9df90449d40d0608f6cb33a5bcbfc8c265f42bab0a", // LCAD
			ReceiverAmount: 1400 / 100000000.0,
		},
		common.DefaultFeeRate,
	)
	if err != nil {
		panic(err)
	}
	log.Printf("Create: %+v", proposal)
	if !proposal.Payload.Valid() {
		panic(common.ErrInvalidPayload)
	}

	payload := proposal.Payload
	if info, err := client.InfoSwapProposal(ctx, 42, payload); true {
		if err != nil {
			panic(err)
		}
		if !info.Payload.Valid() {
			panic(common.ErrInvalidPayload)
		}
		log.Printf("Info: %+v", info)
	}

	if proposal, err := client.FinalizeSwapProposal(ctx, 42, payload); true {
		if err != nil {
			panic(err)
		}
		if !proposal.Payload.Valid() {
			panic(common.ErrInvalidPayload)
		}
		log.Printf("Finalize: %+v", proposal)
	}
}
