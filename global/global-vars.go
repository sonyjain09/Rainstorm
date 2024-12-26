package global

import (
	"os"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/joho/godotenv"
    "sync"
)

var err = godotenv.Load(".env")

var Tcp_ports = []string{}

var Udp_ports = []string{}

var Rainstorm_ports = []string{}


var Tcp_port string = os.Getenv("TCP_PORT")
var Udp_port string = os.Getenv("UDP_PORT")
var Rainstorm_port string = os.Getenv("RAINSTORM_PORT")
var Machine_number string = os.Getenv("MACHINE_NUMBER")
var Membership_log string = os.Getenv("MEMBERSHIP_FILENAME")
var Hydfs_log string = os.Getenv("HYDFS_FILENAME")
var Udp_address string = os.Getenv("MACHINE_UDP_ADDRESS")
var Tcp_address string = os.Getenv("MACHINE_TCP_ADDRESS")
var Rainstorm_address string = os.Getenv("MACHINE_RAINSTORM_ADDRESS")
var Introducer_address string = os.Getenv("INTRODUCER_ADDRESS")
var Leader_address string = os.Getenv("LEADER_ADDRESS")

var Node_id string = ""
var Ring_id string = ""
var Inc_num int = 0
var Ring_map = treemap.NewWithIntComparator()
var Membership_list []Node
var Schedule = make(map[int][]map[string]string)
var Enabled_sus = false
var Cache_set = make(map[string]bool)
var File_prefix string = Udp_address[13:15]
var Partitions [][]int
var Batches = make(map[string][]Tuple) // destination to list of tuples
var AckBatches = make(map[string]string) // destination to list of tuple ids
var BatchesMutex sync.Mutex
var AckBatchesMutex sync.Mutex
var AppendMutex sync.Mutex
var DestMutex sync.Mutex
var ScheduleMutex sync.Mutex
var Reschedule_called = false
var StateMutex sync.Mutex
var TimerMutex sync.Mutex


