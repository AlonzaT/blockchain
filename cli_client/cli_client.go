package cli_client

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/AlonzaT/blockchain/utils"
)

type CommandLine struct{}

func (cli *CommandLine) printer() {
	fmt.Println("Usage: ")
	fmt.Println(" getbalance -address ADDRESS - get the balance for address")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT -mine - Send amount from one address to another")
	fmt.Println(" createwallet - Creates a new Wallet")
	fmt.Println(" listaddresses - Lists the addresses in our wallet file")
	fmt.Println(" reindexutxo - Rebuilds the UTXOset")
	fmt.Println(" startnode -miner ADDRESS - Start a node with ID specified in TokenDuration enviornment var, -miner enables mining")
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env is not set")
		runtime.Goexit()
	}

	getBalanceCMD := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCMD := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCMD := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCMD := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCMD := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCMD := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCMD := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCMD := flag.NewFlagSet("startNode", flag.ExitOnError)

	getBalanceAddress := getBalanceCMD.String("address", " ", "The address")
	createBlockchainAddress := createBlockchainCMD.String("address", " ", "Address to send genesis block reward to")
	sendFrom := sendCMD.String("from", "", "Source wallet address")
	sendTo := sendCMD.String("to", " ", "Destination wallet address")
	sendAmount := sendCMD.Int("amount", 0, "Amount to send")
	sendMine := sendCMD.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCMD.String("miner", "", "Enable mining mode and send reward")

	switch os.Args[1] {
	case "reindexutxo":
		err := reindexUTXOCMD.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":

		//Parse the flags that come after the add flag
		err := getBalanceCMD.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":

		//Create a blockchain
		err := createBlockchainCMD.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":

		//Print the blockchain
		err := printChainCMD.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":

		err := startNodeCMD.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":

		//Print the blockchain
		err := sendCMD.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "createwallet":

		err := createWalletCMD.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "listaddresses":

		err := listAddressesCMD.Parse(os.Args[2:])
		utils.HandleErr(err)

	default:

		cli.printer()
		//makes sure go routines close gracefully
		runtime.Goexit()
	}

	if getBalanceCMD.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCMD.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCMD.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCMD.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress, nodeID)
	}

	if reindexUTXOCMD.Parsed() {
		cli.reindexUTXO(nodeID)
	}

	if createWalletCMD.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCMD.Parsed() {
		cli.listAddresses(nodeID)
	}

	if sendCMD.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCMD.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if printChainCMD.Parsed() {
		cli.printChain(nodeID)
	}

	if startNodeCMD.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCMD.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}

}
