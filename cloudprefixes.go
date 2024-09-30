package main

import (
	"bufio"
	"cloudprefixes/pkg/db"
	"cloudprefixes/pkg/update"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {

	flag.Usage = func() {
		fmt.Println(`
    __ _      ___  __ __ ___   ____  ____    ___ _____ ____ __ __ 
   /  | T    /   \|  T  |   \ |    \|    \  /  _|     l    |  T  T
  /  /| |   Y     |  |  |    \|  o  |  D  )/  [_|   __j|  T|  |  |
 /  / | l___|  O  |  |  |  D  |   _/|    /Y    _|  l_  |  |l_   _j
/   \_|     |     |  :  |     |  |  |    \|   [_|   _] |  ||     |
\     |     l     l     |     |  |  |  .  |     |  T   j  l|  |  |
 \____l_____j\___/ \__,_l_____l__j  l__j\_l_____l__j  |____|__j__|`)
		fmt.Printf("\nUsage\n  %s [OPTION]... [IP ADDRESS]...\n", filepath.Base(os.Args[0]))
		fmt.Println("Search cloud prefixes in database for each IP ADDRESS")
		fmt.Println("\nWith no IP ADDRESS, read standard input.")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	updateData := flag.Bool("update", false, "update all prefixes in database and exit")
	databasePath := flag.String("dbpath", "./cloudprefixes.db", "path to database file")

	flag.Parse()

	manager, err := db.NewIPRangeManager(*databasePath)
	if err != nil {
		log.Fatalf("Error creating IP range manager: %v", err)
	}
	defer manager.Close()

	if *updateData {
		u := update.NewUpdateManager(manager)
		u.UpdateAllSources()
		return
	}

	// read from argument list if supplied otherwise read from stdin
	if flag.NArg() > 0 {
		for _, ip := range flag.Args() {
			found, info, err := manager.ContainsIP(ip)
			if err != nil {
				log.Fatalf("error scanning database: %v", err)
			}
			if found {
				b, err := json.Marshal(info)
				if err != nil {
					log.Fatalf("error serializing to json: %v", err)
				}
				fmt.Println(string(b))
			}
		}
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			found, info, err := manager.ContainsIP(scanner.Text())
			if err != nil {
				log.Fatalf("error scanning database: %v", err)
			}
			if found {
				b, err := json.Marshal(info)
				if err != nil {
					log.Fatalf("error serializing to json: %v", err)
				}
				fmt.Println(string(b))
			}
		}

		if err = scanner.Err(); err != nil {
			log.Fatalf("error reading from stdin: %v", err)
		}
	}

}
