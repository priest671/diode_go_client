// Diode Network Client
// Copyright 2019 IoT Blockchain Technology Corporation LLC (IBTC)
// Licensed under the Diode License, Version 1.0
package main

import (
	"fmt"

	"github.com/diodechain/diode_go_client/command"
	"github.com/diodechain/diode_go_client/config"
	"github.com/diodechain/diode_go_client/contract"
	"github.com/diodechain/diode_go_client/db"
	"github.com/diodechain/diode_go_client/edge"
	"github.com/diodechain/diode_go_client/rpc"
	"github.com/diodechain/diode_go_client/util"
)

var (
	resetCmd = &command.Command{
		Name:        "reset",
		HelpText:    `  Initialize a new account and a new fleet contract in the network. WARNING deletes current credentials!`,
		ExampleText: `  diode reset`,
		Run:         resetHandler,
	}
)

func init() {
	cfg := config.AppConfig
	resetCmd.Flag.BoolVar(&cfg.Experimental, "experimental", false, "send transactions of fleet deployment and device allowlist at seme time")
}

func doInit(cfg *config.Config, client *rpc.RPCClient) error {
	if cfg.FleetAddr != config.DefaultFleetAddr {
		printInfo("Your client has been already initialized, try to publish or browse through Diode Network.")
		return nil
	}
	// deploy fleet
	bn, _ := client.GetBlockPeak()
	if bn == 0 {
		err := fmt.Errorf("not found")
		printError("Cannot find block peak: ", err)
		return err
	}

	var nonce uint64
	var fleetContract contract.FleetContract
	var err error
	fleetContract, err = contract.NewFleetContract()
	if err != nil {
		printError("Cannot create fleet contract instance: ", err)
		return err
	}
	act, _ := client.GetValidAccount(uint64(bn), cfg.ClientAddr)
	if act == nil {
		nonce = 0
	} else {
		nonce = uint64(act.Nonce)
	}
	deployData, err := fleetContract.DeployFleetContract(cfg.RegistryAddr, cfg.ClientAddr, cfg.ClientAddr)
	if err != nil {
		printError("Cannot create deploy contract data: ", err)
		return err
	}
	tx := edge.NewDeployTransaction(nonce, 0, 10000000, 0, deployData, 0)
	res, err := client.SendTransaction(tx)
	if err != nil {
		printError("Cannot deploy fleet contract: ", err)
		return err
	}
	if !res {
		printError("Cannot deploy fleet contract: ", fmt.Errorf("server return err false"))
		return err
	}
	fleetAddr := util.CreateAddress(cfg.ClientAddr, nonce)
	printLabel("New fleet address", fleetAddr.HexString())
	printInfo("Waiting for block to be confirmed - this can take up to a minute")
	watchAccount(client, fleetAddr)
	printInfo("Created fleet contract successfully")
	// generate fleet address
	// send device allowlist transaction
	allowlistData, _ := fleetContract.SetDeviceAllowlist(cfg.ClientAddr, true)
	ntx := edge.NewTransaction(nonce+1, 0, 10000000, fleetAddr, 0, allowlistData, 0)
	res, err = client.SendTransaction(ntx)
	if err != nil {
		printError("Cannot allowlist device: ", err)
		return err
	}
	if !res {
		err = fmt.Errorf("server return err false")
		printError("Cannot allowlist device: ", err)
		return err
	}
	printLabel("Allowlisting device: ", cfg.ClientAddr.HexString())
	printInfo("Waiting for block to be confirmed - this can take up to a minute")
	watchAccount(client, fleetAddr)
	printInfo("Allowlisted device successfully")
	cfg.FleetAddr = fleetAddr
	if cfg.LoadFromFile {
		err = cfg.SaveToFile()
	} else {
		err = db.DB.Put("fleet", fleetAddr[:])
	}
	if err != nil {
		printError("Cannot save fleet address: ", err)
		return err
	}
	printInfo("Client has been initialized, try to publish or browser through Diode Network.")
	return err
}

func doInitExp(cfg *config.Config, client *rpc.RPCClient) error {
	if cfg.FleetAddr != config.DefaultFleetAddr {
		printInfo("Your client has been already initialized, try to publish or browse through Diode Network.")
		return nil
	}
	// deploy fleet
	bn, _ := client.GetBlockPeak()
	if bn == 0 {
		err := fmt.Errorf("not found")
		printError("Cannot find block peak: ", err)
		return err
	}

	var nonce uint64
	var fleetContract contract.FleetContract
	var err error
	fleetContract, err = contract.NewFleetContract()
	if err != nil {
		printError("Cannot create fleet contract instance: ", err)
		return err
	}
	act, _ := client.GetValidAccount(uint64(bn), cfg.ClientAddr)
	if act == nil {
		nonce = 0
	} else {
		nonce = uint64(act.Nonce)
	}
	deployData, err := fleetContract.DeployFleetContract(cfg.RegistryAddr, cfg.ClientAddr, cfg.ClientAddr)
	if err != nil {
		printError("Cannot create deploy contract data: ", err)
		return err
	}
	tx := edge.NewDeployTransaction(nonce, 0, 10000000, 0, deployData, 0)
	res, err := client.SendTransaction(tx)
	if err != nil {
		printError("Cannot deploy fleet contract: ", err)
		return err
	}
	if !res {
		err = fmt.Errorf("server return err false")
		printError("Cannot deploy fleet contract: ", err)
		return err
	}
	fleetAddr := util.CreateAddress(cfg.ClientAddr, nonce)
	printLabel("New fleet address", fleetAddr.HexString())
	// generate fleet address
	// send device allowlist transaction
	allowlistData, _ := fleetContract.SetDeviceAllowlist(cfg.ClientAddr, true)
	ntx := edge.NewTransaction(nonce+1, 0, 10000000, fleetAddr, 0, allowlistData, 0)
	res, err = client.SendTransaction(ntx)
	if err != nil {
		printError("Cannot allowlist device: ", err)
		return err
	}
	if !res {
		err = fmt.Errorf("server return err false")
		printError("Cannot allowlist device: ", err)
		return err
	}
	printLabel("Allowlisting device: ", cfg.ClientAddr.HexString())
	printInfo("Waiting for block to be confirmed - this can take up to a minute")
	watchAccount(client, fleetAddr)
	printInfo("Created fleet contract and allowlisted device successfully")
	cfg.FleetAddr = fleetAddr
	if cfg.LoadFromFile {
		err = cfg.SaveToFile()
	} else {
		err = db.DB.Put("fleet", fleetAddr[:])
	}
	if err != nil {
		printError("Cannot save fleet address: ", err)
		return err
	}
	printInfo("Client has been initialized, try to publish or browser through Diode Network.")
	return nil
}

func resetHandler() (err error) {
	err = app.Start()
	if err != nil {
		return err
	}
	cfg := config.AppConfig
	client := app.datapool.GetClientByOrder(1)
	if cfg.Experimental {
		err = doInitExp(cfg, client)
	} else {
		err = doInit(cfg, client)
	}
	return err
}
