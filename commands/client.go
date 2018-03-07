package commands

import (
	"fmt"
	"io"
	"math/big"

	cmds "gx/ipfs/QmRv6ddf7gkiEgBs1LADv3vC1mkVGPZEfByoiiVybjE9Mc/go-ipfs-cmds"
	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
	cmdkit "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit"

	"github.com/filecoin-project/go-filecoin/abi"
	"github.com/filecoin-project/go-filecoin/core"
	"github.com/filecoin-project/go-filecoin/types"
)

var clientCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Manage client operations",
	},
	Subcommands: map[string]*cmds.Command{
		"add-bid": clientAddBidCmd,
	},
}

var clientAddBidCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Add a bid to the storage market",
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("size", true, false, "size in bytes of the bid"),
		cmdkit.StringArg("price", true, false, "the price of the bid"),
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption("from", "address to send the bid from"),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) {
		n := GetNode(env)

		fromAddr, err := addressWithDefault(req.Options["from"], n)
		if err != nil {
			re.SetError(errors.Wrap(err, "invalid from address"), cmdkit.ErrNormal)
			return
		}

		size, ok := new(big.Int).SetString(req.Arguments[0], 10)
		if !ok {
			re.SetError(fmt.Errorf("invalid size"), cmdkit.ErrNormal)
			return
		}

		price, ok := new(big.Int).SetString(req.Arguments[1], 10)
		if !ok {
			re.SetError(fmt.Errorf("invalid price"), cmdkit.ErrNormal)
			return
		}

		funds := big.NewInt(0).Mul(price, size)

		params, err := abi.ToEncodedValues(price, size)
		if err != nil {
			re.SetError(err, cmdkit.ErrNormal)
			return
		}

		msg := types.NewMessage(fromAddr, core.StorageMarketAddress, funds, "addBid", params)
		if err := n.AddNewMessage(req.Context, msg); err != nil {
			re.SetError(err, cmdkit.ErrNormal)
			return
		}

		msgCid, err := msg.Cid()
		if err != nil {
			re.SetError(err, cmdkit.ErrNormal)
			return
		}

		waitForMessage(n, msgCid, func(blk *types.Block, msg *types.Message, receipt *types.MessageReceipt) {
			id, err := abi.Deserialize(receipt.Return, abi.Integer)
			if err != nil {
				re.SetError(err, cmdkit.ErrNormal)
				return
			}
			re.Emit(id.Val) // nolint: errcheck
		})
	},
	Type: &big.Int{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, id *big.Int) error {
			return PrintString(w, id)
		}),
	},
}
