package shell

import (
    "github.com/arcpop/network/arp"
	"fmt"
	"strconv"
)

var arpHelp = "arp - Possible commands:\n" + 
    "\tarp -> Prints the arp cache\n" + 
    "\tarp query <ip> <interface> [retries] -> Query\n\t\tthe corresponding ip in an arp request\n"
    
func runArp(args []string)  {
    var err error
    if len(args) < 1 {
        fmt.Println(arp.GetCacheAsString())
    } else if args[0] == "help" {
        fmt.Println(arpHelp)
    } else if args[0] == "query" && len(args) >= 3 {
        retries := 0
        if len(args) > 3 {
            retries, err = strconv.Atoi(args[3])
            if err != nil {
                retries = 0
            }
        }
        arp.QueryIP(args[1], args[2], retries)
    } else {
        fmt.Println(arpHelp)
    }
}