package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// DismissNotification https://docs.pushbullet.com/#dismissal-ephemeral
// func DismissNotification(data map[string]string) error {
func dismissNotification(data url.Values) error {
	log.Println("DismissNotification #", prettyPrint(data))

	pushData := pushDismissJSON{
		Type:             "dismissal",
		PackageName:      data.Get("PackageName"),
		NotificationID:   data.Get("NotificationID"),
		NotificationTag:  data.Get("NotificationTag"),
		SourceUserIden:   data.Get("SourceUserIden"),
		ConversationIden: data.Get("ConversationIden"),
	}

	if pushData.NotificationTag == "" {
		pushData.NotificationTag = nil
	}
	requestBody, err := json.Marshal(dismissJSON{
		Type: "push",
		Push: pushData,
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(string(requestBody))

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	request, err := http.NewRequest("POST", "https://api.pushbullet.com/v2/ephemerals", bytes.NewBuffer(requestBody))
	request.Header.Set("Access-Token", key)
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Println(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	resp.Body.Close()

	stringBody := string(body)
	log.Println("result:", stringBody)

	if stringBody != "{}" {
		return errors.New(stringBody)
	}
	return nil
}

// SendMessage Reply Notification
// func SendMessage(data map[string]string) error {
// 	log.Println("SendMessage #", prettyPrint(data))

// 	requestBody, err := json.Marshal(SendJSON{
// 		Type: "push",
// 		Push: SendMessageJSON{
// 			Type:             "messaging_extension_reply",
// 			PackageName:      data["PackageName"],
// 			SourceUserIden:   data["SourceUserIden"],
// 			TargetDeviceIden: data["TargetDeviceIden"],
// 			ConversationIden: data["ConversationIden"], // fmt.Sprintf("{\"package_name\":\"%s\",\"tag\":null,\"id\":%s}", data["PackageName"], data["NotificationID"]),
// 			Message:          data["Message"],
// 		},
// 	})

// 	// POST https://api.pushbullet.com/v2/ephemerals HTTP/1.1
// 	// User-Agent: Pushbullet Desktop 400 (gzip)
// 	// Content-Type: application/json
// 	// Accept: application/json
// 	// Authorization: Bearer xxx
// 	// Host: api.pushbullet.com
// 	// Content-Length: 350
// 	// Expect: 100-continue

// 	// {
// 	//   "type": "push",
// 	//   "push": {
// 	//     "type": "messaging_extension_reply",
// 	//     "package_name": "org.telegram.messenger",
// 	//     "source_user_iden": "ujwGN4LBxoi",
// 	//     "target_device_iden": "ujwGN4LBxoisjBawjBPRoO",
// 	//     "conversation_iden": "{\"package_name\":\"org.telegram.messenger\",\"tag\":null,\"id\":823958560}",
// 	//     "message": "hey"
// 	//   }
// 	// }

// 	if err != nil {
// 		log.Println(err)
// 	}
// 	log.Println(string(requestBody))

// 	timeout := time.Duration(5 * time.Second)
// 	client := http.Client{
// 		Timeout: timeout,
// 	}
// 	request, err := http.NewRequest("POST", "https://api.pushbullet.com/v2/ephemerals", bytes.NewBuffer(requestBody))
// 	request.Header.Set("Access-Token", key)
// 	request.Header.Set("Content-Type", "application/json")
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	resp, err := client.Do(request)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	resp.Body.Close()

// 	stringBody := string(body)
// 	log.Println("result:", stringBody)

// 	time.Sleep(time.Hour)

// 	if stringBody != "{}" {
// 		return errors.New(stringBody)
// 	}
// 	return nil
// }
