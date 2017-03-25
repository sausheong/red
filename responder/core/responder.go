package responder

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/go-nats"
	"log"
	"os"
	"runtime"
)

var ROUTEID string

func init() {
	ROUTEID = os.Getenv("ID")
}

// Run subscribe to a subject with a given callback function
func Run(respond func(RequestInfo, *ResponseInfo)) {
	// setup log file
	file, err := os.OpenFile("responder.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	logger = log.New(file, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)

	// connect to localhost NATS server
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		danger("Cannot connect to NATS server", err)
	}

	// create callback function for subscription
	action := func(msg *nats.Msg) {
		var req_info RequestInfo
		// unmarshal JSON from message data
		err = json.Unmarshal(msg.Data, &req_info)
		if err != nil {
			danger("Cannot unmarshal message to JSON", err)
		}
		// call act function to respond to the request
		resp_info := ResponseInfo{}

		//
		// set response to 200 OK by default
		resp_info.Status = "200"
		// initialize to an empty header
		resp_info.Header = make(map[string][]string)
		// call the respond function passed in from the responder
		respond(req_info, &resp_info)
		//

		// marshal response to JSON
		resp_json, err := json.Marshal(resp_info)
		if err != nil {
			fmt.Println("Cannot marshal response to JSON", err)
			danger("Cannot marshal response to JSON", err)
		}
		// reply through NATS server
		conn.Publish(msg.Reply, []byte(resp_json))
	}
	// subscribe using queue with queue name same as route ID
	// route ID is the subject as well as the queue name
	conn.QueueSubscribe(ROUTEID, ROUTEID, action)
	conn.Flush()

	if err := conn.LastError(); err != nil {
		danger("Cannot subscribe to NATS server", err)
	} else {
		info("Sent subscribe to NATS server")
	}

	fmt.Printf("%s responder ready\n", ROUTEID)
	info(fmt.Sprintf("%s responder ready", ROUTEID))

	runtime.Goexit()
	conn.Close()
}
