/*

maryo/setup.go

the setup function is here for easier to read code

written by superwhiskers, licensed under gnu gplv3.
if you want a copy, go to http://www.gnu.org/licenses/

*/

package main

import (
	// internals
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"math/big"
)

// cert generation function here so i don't need to rewrite it in maryo.go
// (adapted from https://www.socketloop.com/tutorials/golang-create-x509-certificate-private-and-public-keys)
func doCertGen(config string) {
	// variable stuff
	var notBefore time.Time

	// clear the screen because why not?
	clear()

	// show a neat info snippet
	fmt.Printf("- generating certificate and key pair...\n")

	// create necissary directories for this

	// maryo folder
	if doesDirExist("maryo-data") == false {

		// make it
		makeDirectory("maryo-data")

	}

	// clean the cert and key pair if they exist

	// cert.pem
	if doesFileExist("maryo-data/cert.pem") {

		// delete the cert
		deleteFile("maryo-data/cert.pem")

	}

	// key.pem
	if doesFileExist("maryo-data/key.pem") {

		// delete the pubkey
		deleteFile("maryo-data/key.pem")

	}

	// serial number limit stuff
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Printf("failed to generate serial number...\n")

		// panic
		panic(err)

	}


	// cert valid before date
	notBefore = time.Now()

	// cert valid after date
	notAfter := notBefore.Add(365*24*time.Hour)

	// populate certificate with data
    template := &x509.Certificate {

		IsCA: true,
		BasicConstraintsValid: true,
		SerialNumber: serialNumber,
		Subject: pkix.Name{

			CommonName:   "maryo-cert",
			Organization: []string{"pretendo"},
			Country:      []string{"US"},

		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,

	}

	// generate private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)

	// check for errors
	if err != nil {

		// display error
		fmt.Printf("[err]: error while generating key pair...\n")

		// panic
		panic(err)

	}

	// get the public key
	publickey := &privatekey.PublicKey

	// create a self-signed certificate. template = parent
	var parent = template
	cert, err := x509.CreateCertificate(rand.Reader, template, parent, publickey, privatekey)

	// check for errors
	if err != nil {

		// display error
		fmt.Printf("[err]: error while generating certificate...\n")

		// panic
		panic(err)

	}

	// save cert
	certOut, err := os.Create("maryo-data/cert.pem")
	if err != nil {

		// display error
		fmt.Print("failed to open the certificate for writing...\n")

		// panic
		panic(err)

	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {

		// display error
		fmt.Print("failed to write data to the certificate...\n")

		// panic
		panic(err)

	}
	if err := certOut.Close(); err != nil {

		// display error
		fmt.Print("error closing the certificate...\n")

		// panic
		panic(err)

	}
	fmt.Printf("certificate saved...\n")

	// save private key
	keyOut, err := os.OpenFile("maryo-data/key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Print("failed to open the keypair for writing...\n")

		// panic
		panic(err)

	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privatekey)}); err != nil {
		fmt.Print("failed to write data to the keypair...\n")

		// panic
		panic(err)

	}
	if err := keyOut.Close(); err != nil {
		fmt.Print("error closing the keypair...\n")

		// panic
		panic(err)

	}
	fmt.Printf("keypair saved...\n")

	// then, say they were made
	fmt.Printf("finished generating the cert and key pair...\n")
	fmt.Printf("\npress enter to continue...\n")
	_ = input("")

	// then ask if they would like to enable https
	// on the server
	var enableHTTPS string
	for true {

		// clear the screen
		clear()

		// display the message
		fmt.Printf("would you like to enable https on the server?\n")
		fmt.Printf("-> (y|n)\n")
		enableHTTPS = input(": ")

		// make sure it is a valid option
		if (enableHTTPS == "y") || (enableHTTPS == "n") {

			// exit loop if it is
			break

		// if it isn't
		} else {

			// show a message showing valid options
			fmt.Printf("-> please enter y or n\n")

			// stop the event loop to give them time to read
			time.Sleep(1500 * time.Millisecond)

		}

	}

	// clear for a sec
	clear()

	// check if the config even exists
	if !doesFileExist(config) {

		// if it doesn't exist
		// send a message
		fmt.Printf("[err]: there is no config at %s...\n", config)
		fmt.Printf("       please generate one by passing the setup flag\n")
		fmt.Printf("       when running maryo. or, maryo may detect that one has\n")
		fmt.Printf("       has not been created. or, you can just pass the path\n")
		fmt.Printf("       to one...\n")

		// exit
		os.Exit(1)

	}

	// if it does, check if it is valid
	if !checkJSONValidity(config) {

		// send a message
		fmt.Printf("[err]: your config at %s is invalid...\n", config)
		fmt.Printf("       please regenerate it and fix it, which the former\n")
		fmt.Printf("       can be done by setting up maryo again...\n")

		// exit
		os.Exit(1)

	}

	// load it if it is
	configData := readJSONFile(config)

	// do as requested
	if enableHTTPS == "y" {

		// enable https in the config
		configData["config"].(map[string]interface{})["https"] = true

	} else if enableHTTPS == "n" {

		// keep https disabled
		configData["config"].(map[string]interface{})["https"] = false

	}

	// write the log back
	writeJSONFile(config, configData)

	// let the user know it's done, and exit on enter
	fmt.Printf("finished modifying the config...\n")
	fmt.Printf("press enter to continue...\n")
	_ = input("")

}

func generateRomFSPatch(encryptionKeyPath string) {

	// clear the screen
	clear()

	// check for the data directory, the certificates,
	// and the keys

	// alert about it
	fmt.Printf("checking for files...\n")

	// directory checking
	if !doesDirExist("maryo-data") {

		// warn the user
		fmt.Printf("[err]: the maryo-data directory does not exist...\n")
		fmt.Printf("       please generate it by going through setup again\n")
		fmt.Printf("       or by passing the regencerts flag when running maryo.\n")

		// exit
		os.Exit(1)

	}

	// check the files exist
	filesInDataDir := []bool {

		doesFileExist("maryo-data/key.pem"),
		doesFileExist("maryo-data/cert.pem"),

	}

	// check if any are false
	for _, fileStatus := range filesInDataDir {

		// check if that file exists
        if fileStatus == false {

			// if it doesn't, then exit
			fmt.Printf("[err]: one of the required files in the maryo-data\n")
			fmt.Printf("       directory is nonexistent. please regenerate it\n")
			fmt.Printf("       by going through setup again or passing the regencerts\n")
			fmt.Printf("       flag when running maryo again...\n")

			// exit
			os.Exit(1)

		}

	}

	// check for the aes key required for the
	// console to accept the cert and pubkey
	if !doesFileExist("maryo-data/0x0D.key") {

		// if it doesn't
		fmt.Printf("[err]: slot key 0x0D not found in maryo-data...\n")
		fmt.Printf("       please dm me for my magnet link, torrent it,\n")
		fmt.Printf("       and place it in the maryo-data directory as 0x0D.key...\n")

		// exit
		os.Exit(1)

	}

	// now, we can begin generating the patch

	// alert
	fmt.Printf("generating patch...")

	// create directory for the patch

	// but check for it first
	if doesDirExist("patch-out") {

		// remove it if it already does
		deleteFile("patch-out")

	}

	// then create it
	makeDirectory("patch-out")

	// and then proceed to make the
	// require subdirectories

	// title id folder
	makeDirectory("patch-out/0004001b00010002")

	// romfs folder
	makeDirectory("patch-out/0004001b00010002/romfs")

	// and now all we have to do
	// is encrypt the cert and key with
	// the 0x0D aes key

	/*

	aaannnd... since i can't figure out how to get past this point...

	there won't be any  dev here for a bit...

	*/

}

// setup function goes here now
func setup(fileMap map[string]string) {

	// test data
	test := make([]string, 1)
	test[0] = "account"

	// test the 'official' pretendo servers
	testOfficial := make([]string, 1)
	testOfficial[0] = "account"

	// file status map
	fileStat := make(map[string]string)
	fileStat["ne"] = "nonexistent"
	fileStat["iv"] = "invalid"
	fileStat["va"] = "valid"
	fileStat["uk"] = "unknown"

	// setup environment
	clear()

	// set term title
	ttitle("maryo -> setup")

	// show setup screen
	fmt.Printf("== maryo -> setup ===========================================\n")
	fmt.Printf("                                        Steps:               \n")
	fmt.Printf(" welcome to the maryo setup wizard.     > intro              \n")
	fmt.Printf(" this program will walk you through       config creation    \n")
	fmt.Printf(" setting up your very own Pretendo        confirm prefs      \n")
	fmt.Printf(" proxy server for accessing the server.   make https work    \n")
	fmt.Printf(" -> press enter                           profit???          \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("=============================================================\n")
	input("")

	// show config creation screen
	var method string
	for true {
		clear()
		fmt.Printf("== maryo -> setup ===========================================\n")
		fmt.Printf("                                        Steps:               \n")
		fmt.Printf(" how would you like to configure the      intro              \n")
		fmt.Printf(" proxy?                                 > config creation    \n")
		fmt.Printf(" 1. automatic                             confirm prefs      \n")
		fmt.Printf(" 2. custom                                make https work    \n")
		fmt.Printf(" 3. template                              profit???          \n")
		fmt.Printf(" 4. skip this                                                \n")
		fmt.Printf("                                                             \n")
		fmt.Printf(" -> (1|2|3|4)                                                \n")
		fmt.Printf("=============================================================\n")
		method = input(": ")

		// make sure it is a valid option
		if (method == "1") || (method == "2") || (method == "3") || (method == "4") {

			// exit loop if it is
			break

		// if it isn't
		} else {

			// show a message showing valid options
			fmt.Printf("-> please enter 1, 2, 3, or 4\n")

			// stop the event loop to give them time to read
			time.Sleep(1500 * time.Millisecond)

		}

	}

	// create config var
	var config map[string]map[string]string

	// show log when
	clear()
	fmt.Printf("== maryo -> setup ===========================================\n")
	fmt.Printf("                                                             \n")
	fmt.Printf(" configuring proxy..                                         \n")
	fmt.Printf(" current config status: %s\n", fileStat[fileMap["config"]])
	// automatic config making
	if method == "1" {

		// show some messages
		fmt.Printf(" method: automatic..\n")
		fmt.Printf("-- beginning tests\n")
		fmt.Printf(" 1. attempting to detect endpoints running on this machine\n")

		// test for endpoints on this machine
		result := make([]bool, len(test))
		for x := 0; x < len(test); x++ {

			// test the endpoint
			fmt.Printf("  %s %s -> %s", utilIcons["uncertain"], testEndpoints["ninty"][test[x]], testEndpoints["local"][test[x]])

			// get the json
			res, err := get(strings.Join([]string{"http://", testEndpoints["local"][test[x]], "/isthisworking"}, ""))

			// prepare a struct for the json
			var parsedRes isitworkingStruct

			// parse it
			err2 := json.Unmarshal([]byte(res), &parsedRes)

			// make sure it isn't empty
			if res != "" {

				// if there is an error in JSON parsing
				if err2 != nil {

					// and not an error with the request
					if err == nil {

						// show a message
						fmt.Printf("\n[err] : error when parsing JSON during validating %s server\n", test[x])

						// show a traceback
						panic(err2)

					}

				}

			}

			// handle the results
			if (parsedRes.Server == resMap[test[x]]) && (err == nil) && (res != "") {

				// MS, step your game up and support ansi escape codes
				consoleSequence(fmt.Sprintf("%s\n", padStrToMatchStr(fmt.Sprintf("\r  %s%s%s%s %s -> %s", code("green"), code("bold"), utilIcons["success"], code("reset"), testEndpoints["ninty"][test[x]], testEndpoints["local"][test[x]]), fmt.Sprintf("  %s %s -> %s", utilIcons["uncertain"], testEndpoints["ninty"][test[x]], testEndpoints["local"][test[x]]), " ")))
				result[x] = true

			} else {

				// make windows great again (as if it ever was)
				consoleSequence(fmt.Sprintf("%s\n", padStrToMatchStr(fmt.Sprintf("\r  %s%s%s%s %s -> %s", code("red"), code("bold"), utilIcons["failiure"], code("reset"), testEndpoints["ninty"][test[x]], testEndpoints["local"][test[x]]), fmt.Sprintf("  %s %s -> %s", utilIcons["uncertain"], testEndpoints["ninty"][test[x]], testEndpoints["local"][test[x]]), " ")))
				result[x] = false

			}

		}

		// show the user that we are detecting the official server
		fmt.Printf(" 2. attempting to test endpoints on the official server\n")

		// test for endpoints on the official pretendo servers
		resultOfficial := make([]bool, len(testOfficial))
		for x := 0; x < len(testOfficial); x++ {

			// test the endpoint
			fmt.Printf("  %s %s -> %s", utilIcons["uncertain"], testEndpoints["ninty"][testOfficial[x]], testEndpoints["official"][testOfficial[x]])

			// get the json
			res2, err3 := get(strings.Join([]string{"http://", testEndpoints["official"][testOfficial[x]], "/isthisworking"}, ""))
			// prepare a struct for the json
			var parsedRes2 isitworkingStruct

			// parse it
			err4 := json.Unmarshal([]byte(res2), &parsedRes2)

			// make sure the request isn't empty
			if res2 != "" {

				// if there is an error in JSON parsing
				if err4 != nil {

					// and not an error with the request
					if err3 == nil {

						// show an error message
						fmt.Printf("\n[err] : error when parsing JSON during validating %s server\n", testOfficial[x])

						// and show a traceback
						panic(err4)

					}

				}

			}

			// handle the results
			if (parsedRes2.Server == resMap[testOfficial[x]]) && (err3 == nil) && (res2 != "") {

				// why is windows like this
				consoleSequence(fmt.Sprintf("%s\n", padStrToMatchStr(fmt.Sprintf("\r  %s%s%s%s %s -> %s", code("green"), code("bold"), utilIcons["success"], code("reset"), testEndpoints["ninty"][testOfficial[x]], testEndpoints["official"][testOfficial[x]]), fmt.Sprintf("  %s %s -> %s", utilIcons["uncertain"], testEndpoints["ninty"][testOfficial[x]], testEndpoints["official"][testOfficial[x]]), " ")))
				resultOfficial[x] = true

			} else {

				// thank goodness for my shorthand function
				consoleSequence(fmt.Sprintf("%s\n", padStrToMatchStr(fmt.Sprintf("\r  %s%s%s%s %s -> %s", code("red"), code("bold"), utilIcons["failiure"], code("reset"), testEndpoints["ninty"][testOfficial[x]], testEndpoints["official"][testOfficial[x]]), fmt.Sprintf("  %s %s -> %s", utilIcons["uncertain"], testEndpoints["ninty"][testOfficial[x]], testEndpoints["official"][testOfficial[x]]), " ")))
				resultOfficial[x] = false

			}

		}

		// print out the results
		fmt.Printf("-- printing results of tests\n")

		// for local servers
		for x := 0; x < len(result); x++ {

			// add a header saying that these are local results
			fmt.Printf("- local\n")

			// print the results
			if result[x] == true {

				// show a successful message
				fmt.Printf(" %s: success\n", test[x])

			} else {

				// or a failiure message
				fmt.Printf(" %s: failiure\n", test[x])

			}

		}

		// for the official servers
		for x := 0; x < len(resultOfficial); x++ {

			// add a header saying that these are official results
			fmt.Printf("- pretendo\n")

			// print the results
			if resultOfficial[x] == true {

				// if successful
				fmt.Printf(" %s: success\n", testOfficial[x])

			} else {

				// or failed
				fmt.Printf(" %s: failiure\n", testOfficial[x])

			}

		}

		// begin generating the config
		fmt.Printf("-- generating config\n")

		// create cfgTest, cfgResult, using, and useLocal variables
		// also make the really long name variable
		var using string
		var cfgTest []string
		var cfgResult []bool
		var useLocal bool
		var doesOfficialHaveAnyWorkingEndpoints bool

		// scan the local server result list to see if any are true, since they have priority
		useLocal = false

		// scan it
		for x := 0; x < len(result); x++ {

			// if it works, use local servers
			if result[x] == true {

				// set the variable
				useLocal = true

			}

		}

		// local servers have priority
		if useLocal == true {

			// set the needed variables
			using = "local"
			cfgTest = test
			cfgResult = result

		} else {

			// check this first to see if official even works
			for x := 0; x < len(resultOfficial); x++ {
				doesOfficialHaveAnyWorkingEndpoints = false
				if resultOfficial[x] == true {
					doesOfficialHaveAnyWorkingEndpoints = true
				}
			}

			// if the official servers work
			if doesOfficialHaveAnyWorkingEndpoints == true {

				// set these if it does
				using = "official"
				cfgTest = testOfficial
				cfgResult = resultOfficial

			} else {

				// exit the program
				clear()
				fmt.Printf("no servers are running currently, please try again later.")
				os.Exit(0)

			}

		}

		// make a map for the config
		config = make(map[string]map[string]string)

		// make the endpoints and config a map[string]string
		config["endpoints"] = make(map[string]string)
		config["config"] = make(map[string]string)

		// apply a nice helping of all of the working endpoints to the config
		for x := 0; x < len(cfgTest); x++ {

			// if this endpoint works
			if cfgResult[x] == true {

				// set it in the config
				config["endpoints"][testEndpoints["ninty"][cfgTest[x]]] = testEndpoints[using][cfgTest[x]]

			}

		}

		// set some config vars
		config["config"]["decryptOutgoing"] = "true"

		// wait for them to press enter
		fmt.Printf("\npress enter to continue...\n")
		_ = input("")

		// creating a custom config
	} else if method == "2" {

		// show a little message
		fmt.Printf(" method: custom..\n")

		// number of values in the config
		numVals := 0

		// config
		config = make(map[string]map[string]string)

		// make the endpoints and config a map[string]string
		config["endpoints"] = make(map[string]string)
		config["config"] = make(map[string]string)

		// temp vars
		var inputtedFrom string
		var inputtedTo string

		// a infinite loop for custom configs
		for true {
			clear()

			// reset temporary vars
			inputtedFrom = ""
			inputtedTo = ""

			// display ui
			fmt.Printf(" you have %s redirection(s) already in\n", strconv.Itoa(numVals))
			fmt.Printf(" press <Enter> on an empty line to stop\n")

			// ask for conf vals
			inputtedFrom = input("from: ")

			// if the from field is empty
			if inputtedFrom == "" {

				// exit the loop
				break

			}

			// ask for the to value
			inputtedTo = input("to: ")

			// if the field is empty
			if inputtedTo == "" {

				// exit the loop
				break

			}

			// place them in the config var
			config["endpoints"][inputtedFrom] = inputtedTo

			// update info
			numVals++

		}

		// set default config vars
		config["config"]["decryptOutgoing"] = "true"

		// loading a template
	} else if method == "3" {

		// template variable since i have to reserve it
		var tmpl string

		// ask for choice
		fmt.Printf(" method: template..\n")
		for true {

			// clear screen
			clear()

			// show ui
			fmt.Printf("-- please select a template\n")
			fmt.Printf(" 1. local server\n")
			fmt.Printf(" 2. pretendo servers\n")
			// TODO: make official servers work
			// not adding this one until i can figure out how
			// to make ClientCertificate.cer work
			// fmt.Printf(" 3. official servers\n")

			// ask for input
			tmpl = input(": ")

			// break if it's a valid option
			if (tmpl == "1") || (tmpl == "2") {

				// break
				break

			} else {

				// otherwise show a help message
				fmt.Printf("-> please enter 1 or 2\n")

				// sleep to let them read it
				time.Sleep(1500 * time.Millisecond)

			}

		}

		// load the selected template into the config var
		// TODO: add variable templates
		if tmpl == "1" {

			// load the template
			config = localConf

		} else if tmpl == "2" {

			// load this other template
			config = pretendoConf

		}

	}

	// not everyone wants to generate a new config
	if method != "4" {

		// prettify the JSON
		stringifiedConfig, err := json.MarshalIndent(config, "", "    ")

		// error handling
		if err != nil {

			// show an error message
			fmt.Printf("[err] : error while prettifying JSON\n")

			// show traceback
			panic(err)

		}

		// turn it into a string
		stringifiedJSON := string(stringifiedConfig[:])

		// confirm the preferences
		var areSettingsOkay string

		// for loop for confirming
		for true {

			// clear screen for cleanliness
			clear()

			// display the UI
			fmt.Printf("== maryo -> setup ===========================================\n")
			fmt.Printf("                                        Steps:               \n")
			fmt.Printf(" are you okay with the settings below?    intro              \n")
			fmt.Printf("                                          config creation    \n")
			fmt.Printf("                                        > confirm prefs      \n")
			fmt.Printf("                                          make https work    \n")
			fmt.Printf("                                          profit???          \n")
			fmt.Printf("                                                             \n")
			fmt.Printf(stringifiedJSON)
			fmt.Printf("\n                                                             \n")
			fmt.Printf("-> (y|n)                                                     \n")
			fmt.Printf("=============================================================\n")
			areSettingsOkay = input(": ")

			// check if response is okay
			if (areSettingsOkay == "y") || (areSettingsOkay == "n") {

				// exit loop if valid
				break

			} else {

				// show a help message
				fmt.Printf("-> please enter y or n")

				// let them read it
				time.Sleep(1500 * time.Millisecond)

			}

		}

		// check if they answered no
		if areSettingsOkay == "n" {

			// if so, clear
			clear()

			// and quit program
			os.Exit(0)

		}

		// error handling
		if err != nil {

			// show error message
			fmt.Printf("[err] : error when stringifying json")

			// show traceback
			panic(err)

		}

		// make sure the maryo folder exists
		if doesDirExist("maryo-data") == false {

			// make it if it doesn't
			makeDirectory("maryo-data")

		}

		// place it into the file
		if fileMap["config"] == "iv" {

			// delete the existing config
			deleteFile("maryo-data/config.json")

			// create a new one
			createFile("maryo-data/config.json")

			// write the data to the file
			writeByteToFile("maryo-data/config.json", stringifiedConfig)

		} else if fileMap["config"] == "ne" {

			// create the config
			createFile("maryo-data/config.json")

			// write the config to the file
			writeByteToFile("maryo-data/config.json", stringifiedConfig)

		} else if fileMap["config"] == "uk" {

			// detect status of config and do the
			// things to write to it.
			if doesFileExist("maryo-data/config.json") == true {

				// delete existing config
				deleteFile("maryo-data/config.json")

				// create a new one
				createFile("maryo-data/config.json")

				// write the config to it
				writeByteToFile("maryo-data/config.json", stringifiedConfig)

			} else {

				// create the config
				createFile("maryo-data/config.json")

				// write the config to the file
				writeByteToFile("maryo-data/config.json", stringifiedConfig)

			}

		}

	}

	// generate a https cert
	clear()

	fmt.Printf("== maryo -> setup ===========================================\n")
	fmt.Printf("                                        Steps:               \n")
	fmt.Printf(" now, it is time to generate a https      intro              \n")
	fmt.Printf(" cert to encrypt your data                config creation    \n")
	fmt.Printf(" -> press enter                           confirm prefs      \n")
	fmt.Printf("                                        > make https work    \n")
	fmt.Printf("                                          profit???          \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("=============================================================\n")
	_ = input("")

	// generate the certificates
	doCertGen("maryo-data/config.json")

	// show them the finished screen
	clear()

	// display the UI
	fmt.Printf("== maryo -> setup ===========================================\n")
	fmt.Printf("                                        Steps:               \n")
	fmt.Printf(" congratulations, you are finished        intro              \n")
	fmt.Printf(" setting up maryo!                        config creation    \n")
	fmt.Printf(" -> press enter                           confirm prefs      \n")
	fmt.Printf("                                          make https work    \n")
	fmt.Printf("                                        > profit???          \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("                                                             \n")
	fmt.Printf("=============================================================\n")
	_ = input("")

	// display a message saying that they need to restart the program to use the new config
	clear()
	fmt.Printf("run this program again to use the new configuration\n")

}
