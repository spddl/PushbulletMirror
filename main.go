package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nhooyr.io/websocket"
	// "github.com/nu7hatch/gouuid"
	uuid "github.com/nu7hatch/gouuid"
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile) // https://ispycode.com/GO/Logging/Setting-output-flags

	log.Println(os.Args)

	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "pushbulletapi://") {
			cmd := strings.TrimPrefix(os.Args[1], "pushbulletapi://")
			cmdArray := strings.Split(cmd, "/")

			switch cmdArray[0] {
			case "dismissal":
				log.Println(cmdArray[1:])
				DismissNotification(map[string]string{
					"NotificationID":   cmdArray[1],
					"NotificationTag":  cmdArray[2],
					"PackageName":      cmdArray[3],
					"SourceUserIden":   cmdArray[4],
					"ConversationIden": fmt.Sprintf("{\"package_name\":\"%s\",\"tag\":null,\"id\":%s}", cmdArray[3], cmdArray[1]),
				})

			// case "reply":
			// 	log.Println(cmdArray[1:])
			// 	DismissNotification(map[string]string{
			// 		"NotificationID":  cmdArray[1],
			// 		"NotificationTag": cmdArray[2],
			// 		"PackageName":     cmdArray[3],
			// 		"SourceUserIden":  cmdArray[4],
			// 		"ConversationIden": fmt.Sprintf("{\"package_name\":\"%s\",\"tag\":null,\"id\":%s}", cmdArray[3], cmdArray[1]),
			// 	})
			// 	SendMessage(map[string]string{
			// 		"PackageName":     cmdArray[3],
			// 		"SourceUserIden":  cmdArray[4],
			// 		"TargetDeviceIden": "",
			// 		"ConversationIden": fmt.Sprintf("{\"package_name\":\"%s\",\"tag\":null,\"id\":%s}", cmdArray[3], cmdArray[1]),
			// 		"Message": "",
			// 	})

			default:
				log.Printf("default: %s\n", cmdArray)
			}

			// log.Println("os.Exit(2)")
			// time.Sleep(time.Hour)

			os.Exit(2)
		}
	}

	prognamepath, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	prognamedecoded := `\"` + strings.ReplaceAll(prognamepath, `\`, `\\`) + `\"`
	err = pushbulletProtocolCheck(prognamepath) // Prüft das PushbulletApi:// Protokoll
	if err != nil {
		log.Println(err)
		pushbulletProtocolCreateReg(prognamedecoded) // und falls der PFad zu dieser Exe nicht stimmt wird ein Reg file erstellt
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var c *websocket.Conn
	for {
		var err error
		c, _, err = websocket.Dial(ctx, "wss://stream.pushbullet.com/websocket/"+key, websocket.DialOptions{}) // erstellt die Websocket Verbindung
		if err != nil {
			log.Println(err)
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}
	log.Println("Websocket verbunden.")
	defer c.Close(websocket.StatusNormalClosure, "")
	// go httpServer(45214)

	for {
		_, reader, err := c.Reader(ctx)
		if err != nil {
			log.Println(err)
			break
		} else {
			data := make(map[string]interface{})
			err := json.NewDecoder(reader).Decode(&data)
			if err != nil {
				log.Println(err)
			}
			if data["type"] == "push" { // https://docs.pushbullet.com/#realtime-event-stream
				dataMap := parseResponse(data["push"])
				response(dataMap)
			} else if data["type"] != "nop" {
				log.Printf("default: %+v\n", data)
			}
		}
	}
}

func prettyPrint(data interface{}) string {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	return string(out)
}

func parseResponse(data interface{}) map[string]string {
	// log.Println("ParseResponse #", prettyPrint(data))

	dataMap := map[string]string{}
	switch v := data.(type) {
	case string:
		// log.Println("1_string:", v)
	case int:
		// log.Printf("1_int: %v \n", v)
	case map[string]interface{}:
		for k, vv := range v {
			switch vb := vv.(type) {
			case string:
				// log.Printf("%s -> %s\n", k, vv)
				dataMap[k] = vv.(string)
			case float64:
				// log.Printf("2_float64: %v\n", vb)
			case bool:
				// log.Printf("2_bool: %v\n", vb)
			case int:
				log.Printf("2_int: %v\n", vb)
				// dataMap[k] = vv.(int)
			case []interface{}:
				// log.Printf("2_[]interface: %v \n", vb)
			default:
				// log.Printf("2_I don't know about type %T\n", vb)
			}
		}
	default:
		// log.Printf("1_I don't know about type %T!\n", v)
	}
	return dataMap
}

func response(data map[string]string) {
	log.Println("Response #", prettyPrint(data))
	// log.Println("Response #", data["type"], data["package_name"])

	id, _ := uuid.NewV4()
	fileName := id.String()

	if data["type"] == "mirror" { // Benachrichtigung ist erschienen
		saveIcon(data["icon"], fileName)
		iconPath, err := filepath.Abs(fileName + ".jpg")
		if err != nil {
			log.Println(err)
		}

		fileNamePath := filepath.Join(".", fileName)

		notification := Notification{
			AppID:   "Pushbullet",
			Title:   data["title"],
			Message: data["body"],
			Icon:    iconPath,
			Tag:     data["package_name"] + data["notification_id"] + data["source_device_iden"] + data["source_user_iden"],
			Actions: []Action{},
		}

		// pushbulletapi Protokoll
		notification.Actions = append(notification.Actions, Action{"protocol", "Close", fmt.Sprintf(`pushbulletapi://dismissal/%s/%s/%s/%s`, data["notification_id"], "null", data["package_name"], data["source_user_iden"])})
		// http Protokoll
		// notification.Actions = append(notification.Actions, Action{"protocol", "Close", fmt.Sprintf(`http://localhost:%d/dismissal/%s/%s/%s/%s`, 45214, data["notification_id"], "null", data["package_name"], data["source_user_iden"])})

		pushToast(fileNamePath+".ps1", notification)

		go func(fileNamePath string) {
			time.Sleep(100 * time.Millisecond)
			os.Remove(fileNamePath + ".ps1")
			os.Remove(fileNamePath + ".jpg")
		}(fileNamePath)

	} else if data["type"] == "dismissal" { // Benachrichtigung wurde entfernt
		fileNamePath := filepath.Join(".", fileName)
		removeToast(fileNamePath+"_remove.ps1", map[string]string{"tag": data["package_name"] + data["notification_id"] + data["source_device_iden"] + data["source_user_iden"]})
		go func(fileNamePath string) {
			time.Sleep(100 * time.Millisecond)
			os.Remove(fileNamePath + "_remove.ps1")
		}(fileNamePath)
	} else {
		log.Println("Default Case:", prettyPrint(data))
	}
}

func saveIcon(data, fileName string) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	//Encode from image format to writer
	pngFilename := fileName + ".jpg"
	f, err := os.OpenFile(pngFilename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = jpeg.Encode(f, m, &jpeg.Options{Quality: 75})
	if err != nil {
		log.Fatal(err)
		return
	}
	f.Close()
}
