// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

type Command string

const (
	CmdGetBlockCount   = Command("getblockcount")
	CmdGetNewAddress   = Command("getnewaddress")
	CmdListUnspent     = Command("listunspent")
	CmdLockUnspent     = Command("lockunspent")
	CmdListLockUnspent = Command("listlockunspent")
	CmdGetTransaction  = Command("gettransaction")
	CmdGetAddressInfo  = Command("getaddressinfo")
	CmdSendMany        = Command("sendmany")

	CmdDumpPrivkey                  = Command("dumpprivkey")
	CmdCreateRawTransaction         = Command("createrawtransaction")
	CmdDecodeRawTransaction         = Command("decoderawtransaction")
	CmdFundRawTransaction           = Command("fundrawtransaction")
	CmdSignRawTransactionWithKey    = Command("signrawtransactionwithkey")
	CmdSignRawTransactionWithWallet = Command("signrawtransactionwithwallet")
)
