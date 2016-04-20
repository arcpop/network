package shell

import (
    "fmt"
	"strings"
	"os"
	"bufio"
)



func Run() {
    stdin := bufio.NewReader(os.Stdin)
    for {
        line, _, err := stdin.ReadLine()
        if err != nil {
            fmt.Println(err)
        }
        
        args := strings.Split(string(line), " ")
        
        switch (args[0]) {
            case "ping":
                runPing(args[1:])
            case "route":
                runRoute(args[1:])
            case "arp":
                runArp(args[1:])
            case "iface":
                runIface(args[1:])
        }
    }
}