package main

import (
	"fmt"
	"github.com/bgentry/heroku-go"
	"io"
	"os"
	"text/tabwriter"
)

var cmdTransfer = &Command{
	Run:      runTransfer,
	Name:     "transfer",
	Usage:    "[-a <app>] <email>",
	Category: "app",
	Short:    "transfer app ownership to a collaborator" + extra,
}

func init() {
	cmdTransfer.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runTransfer(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	recipient := args[0]
	_, err := client.AppTransferCreate(mustApp(), recipient)
	must(err)
}

var cmdTransfers = &Command{
	Run:      runTransfers,
	Name:     "transfers",
	Usage:    "[-a <app>]",
	Category: "app",
	Short:    "list existing app transfers" + extra,
}

func init() {
	cmdTransfers.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runTransfers(cmd *Command, args []string) {
	transfers, err := client.AppTransferList(nil)
	must(err)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	for i := range transfers {
		listTransfer(w, transfers[i])
	}
}

func listTransfer(w io.Writer, t heroku.AppTransfer) {
	listRec(w,
		t.App.Name,
		abbrev(t.Owner.Email, 10),
		abbrev(t.Recipient.Email, 10),
		t.State,
		prettyTime{t.UpdatedAt},
	)
}

var cmdTransferAccept = &Command{
	Run:      runTransferAccept,
	Name:     "transfer-accept",
	Category: "app",
	Short:    "accept an inbound app transfer" + extra,
}

func runTransferAccept(cmd *Command, args []string) {
	transferId := mustLookupTransfer(mustApp())
	must(updateTransferState(transferId, "accepted"))
}

var cmdTransferDecline = &Command{
	Run:      runTransferDecline,
	Name:     "transfer-decline",
	Usage:    "[-a <app>]",
	Category: "app",
	Short:    "decline an inbound app transfer" + extra,
}

func init() {
	cmdTransferDecline.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runTransferDecline(cmd *Command, args []string) {
	transferId := mustLookupTransfer(mustApp())
	must(updateTransferState(transferId, "declined"))
}

var cmdTransferCancel = &Command{
	Run:      runTransferCancel,
	Name:     "transfer-cancel",
	Usage:    "[-a <app>]",
	Category: "app",
	Short:    "cancel an outbound app transfer" + extra,
}

func init() {
	cmdTransferCancel.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runTransferCancel(cmd *Command, args []string) {
	must(client.AppTransferDelete(mustLookupTransfer(mustApp())))
}

func mustLookupTransfer(appname string) string {
	// If the API starts allowing app identity instead of requiring
	// app-transfer UUID, this lookup will be unnecessary.
	transfers, err := client.AppTransferList(nil)
	must(err)
	var transferId string
	for i := range transfers {
		if transfers[i].App.Name == appname {
			transferId = transfers[i].Id
			break
		}
	}
	if transferId == "" {
		fmt.Printf("no pending transfer for app %s\n", appname)
		os.Exit(1)
	}
	return transferId
}

func updateTransferState(transferId, newstate string) error {
	_, err := client.AppTransferUpdate(transferId, newstate)
	return err
}
