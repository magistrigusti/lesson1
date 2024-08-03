package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"

	"ton-lessons2/internal/app"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// func foo(str chan string) {
// 	i := 0
// 	for {
// 		time.Sleep(time.Second)
// 		i++
// 		str <- fmt.Sprint("something ", i)
// 	}
// }

// func bar(str chan string) {
// 	for smth := range str {
// 		logrus.Info(smth)
// 	}
// }

func run() error {
	if err := app.InitApp(); err != nil {
		return err
	}

	uuid := uuid.New().String()
	jettonWallet := address.MustParseAddr("EQBds0yXEvnComaVqEymanuWiY5t28CJQk27FYBE8wLh1Lu3")
	logrus.Info("UUID for transaction: ", uuid)

	client := liteclient.NewConnectionPool()

	if err := client.AddConnectionsFromConfig(context.Background(), app.CFG.MainnetConfig); err != nil {
		return err
	}

	api := ton.NewAPIClient(client)

	wall, err := wallet.FromSeed(api, app.CFG.Wallet.Seed, wallet.V4R2)
	if err != nil {
		return err
	}

	logrus.Info(wall.Address())

	lastMaster, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		return err
	}

	acc, err := api.GetAccount(
		context.Background(),
		lastMaster,
		wall.Address(),
	)

	if err != nil {
		return err
	}

	lastLt := acc.LastTxLT

	transactions := make(chan *tlb.Transaction)

	go api.SubscribeOnTransactions(
		context.Background(),
		wall.Address(),
		lastLt,
		transactions,
	)

	logrus.Info("Start checking transactions")
	for {
		select {
		case newTransaction := <-transactions:
			if newTransaction.IO.In.MsgType != tlb.MsgTypeInternal {
				logrus.Warn("not internal message!")
				continue
			}

			internalMessage := newTransaction.IO.In.AsInternal()
			if internalMessage.Body == nil {
				logrus.Warn("empty body")
				continue
			}

			if internalMessage.SrcAddr.String() != jettonWallet.String() {
				logrus.Warn("not our jetton wallet")
				continue
			}

			bodySlice := internalMessage.Body.BeginParse()
			opcode, err := bodySlice.LoadUInt(32)
			if err != nil {
				logrus.Error("error when get opcode: ", err)
				continue
			}

			if opcode != 0x178d4519 {
				logrus.Warn("not jetton notification")
				continue
			}

			queryId, err := bodySlice.LoadUInt(64)
			if err != nil {
				logrus.Error("query id err: ", err)
				continue
			}

			amount, err := bodySlice.LoadCoins()
			if err != nil {
				logrus.Error("amount, err: ", err)
				continue
			}

			sender, err := bodySlice.LoadAddr()
			if err != nil {
				logrus.Error("address err: ", err)
				continue
			}

			fwdPayload, err := bodySlice.LoadMaybeRef()
			if err != nil {
				logrus.Error("fwd payload err: ", err)
				continue
			}

			fwdOp, err := fwdPayload.LoadUInt(32)
			if err != nil {
				logrus.Error("fwd op err: ", err)
				continue
			}

			if fwdOp != 0 {
				logrus.Error("not text comment")
				continue
			}

			textComment, err := fwdPayload.LoadStringSnake()
			if err != nil {
				logrus.Error("text comment err: ", err)
				continue
			}

			logrus.Info("[JTN] new transaction!")
			logrus.Info("[JTN] sender: ", sender)
			logrus.Info("[JTN] amount: ", amount)
			logrus.Info("[JTN] query id: ", queryId)
			logrus.Info("[JTN] comment: ", textComment)

			// if opcode != 0 {
			// 	logrus.Warn("not text comment, skip")
			// 	continue
			// }

			// comment, err := bodySlice.LoadStringSnake()
			// if err != nil {
			// 	logrus.Error("error when get text comment: ", err)
			// 	continue
			// }

			// logrus.Info("[MSG] sender: ", internalMessage.SrcAddr.String())
			// logrus.Info("[MSG] amount: ", internalMessage.Amount.String())
			// logrus.Info("[MSG] text comment: ", comment)

			// if comment == uuid {
			// 	logrus.Info("Success topup! User uuid: ", uuid)
			// }
		}
	}

	return nil
}
