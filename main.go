package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"image"
	"image/jpeg"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	"nhooyr.io/websocket"
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile) // https://ispycode.com/GO/Logging/Setting-output-flags

	prognamepath, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	// file, err := os.OpenFile(filepath.Join(filepath.Dir(prognamepath), fmt.Sprintf("_%d-%s.log", os.Getpid(), time.Now().Format("2006-01-02T15-04-05"))), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()
	// log.SetOutput(file)

	fmt.Println(filepath.Join(prognamepath, fmt.Sprintf("_%d-%s.log", os.Getpid(), time.Now().Format("2006-01-02T15-04-05"))))
	log.Println(os.Args)

	if len(os.Args) > 1 {
		u, err := url.Parse(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}

		if u.Scheme == "pushbulletapi" {
			switch u.Host {
			case "dismissal":
				err := DismissNotification(u.Query())
				if err != nil {
					log.Println(err)
				}

			case "reply":
				// err := DismissNotification(u.Query())
				// if err != nil {
				// 	log.Println(err)
				// }
				// err := SendMessage(u.Query())
				// if err != nil {
				// 	log.Println(err)
				// }

			default:
				log.Println("default")
			}

		}
		// log.Println("os.Exit(2)")
		// time.Sleep(time.Hour)

		os.Exit(2)
	}

	prognamedecoded := `\"` + strings.ReplaceAll(prognamepath, `\`, `\\`) + `\"`
	err = pushbulletProtocolCheck(prognamepath) // Prüft das PushbulletApi:// Protokoll
	if err != nil {
		log.Println(err)
		pushbulletProtocolCreateReg(prognamedecoded) // und falls der PFad zu dieser Exe nicht stimmt wird ein Reg file erstellt
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		var c *websocket.Conn
		for {
			var err error
			c, _, err = websocket.Dial(ctx, "wss://stream.pushbullet.com/websocket/"+key, nil) // erstellt die Websocket Verbindung
			if err != nil {
				log.Println(err)
				time.Sleep(10 * time.Second)
			} else {
				break
			}
		}
		log.Println("Websocket verbunden.")
		defer c.Close(websocket.StatusNormalClosure, "")

		for {
			_, reader, err := c.Reader(ctx)
			if err != nil {
				log.Println(err)
				break
			} else {

				// dataDEMO := make(map[string]interface{})
				// e := json.NewDecoder(reader).Decode(&dataDEMO)
				// if e != nil {
				// 	log.Println(e)
				// }
				// log.Println("dataDEMO", prettyPrint(dataDEMO))

				var data JSONEntry
				err := json.NewDecoder(reader).Decode(&data)
				if err != nil {
					log.Println(err)
				}

				if data.Type == "push" { // https://docs.pushbullet.com/#realtime-event-stream
					log.Println(prettyPrint(data))
					response(data.Push)
				} else if data.Type != "nop" {
					log.Println(prettyPrint(data))
					log.Printf("default: %+v\n", data)
				}

			}
		}
		log.Println("Reconnect")
		time.Sleep(10 * time.Second)
	}
}

func response(data JSONPushEntry) {
	id, _ := uuid.NewV4()
	fileName := id.String()

	if data.Type == "mirror" { // Benachrichtigung ist erschienen

		err := saveIcon(data.Icon, fileName)
		if err != nil {
			log.Println(err)
		}
		iconPath, err := filepath.Abs(fileName + ".jpg")
		if err != nil {
			log.Println(err)
		}

		fileNamePath := filepath.Join(".", fileName)

		notification := Notification{
			AppID:   "Pushbullet",
			Title:   data.Title,
			Message: data.Body,
			Icon:    iconPath,
			Tag:     data.PackageName + data.NotificationID + data.SourceDeviceIden + data.SourceUserIden,
			Actions: []Action{},
		}

		protocolDismissal := "pushbulletapi://dismissal"
		u, err := url.Parse(protocolDismissal)
		if err != nil {
			log.Println(err)
		}

		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			log.Println(err)
		}

		q.Add("NotificationID", data.NotificationID)
		q.Add("NotificationTag", data.NotificationTag)
		q.Add("ConversationIden", data.ConversationIden)
		q.Add("SourceUserIden", data.SourceUserIden)
		q.Add("PackageName", data.PackageName)

		u.RawQuery = q.Encode()
		notification.Actions = append(notification.Actions, Action{"protocol", "Close", strings.ReplaceAll(u.String(), "&", "&amp;")})

		pushToast(fileNamePath+".ps1", notification)

		go func(fileNamePath string) {
			time.Sleep(100 * time.Millisecond)
			os.Remove(fileNamePath + ".ps1")
			os.Remove(fileNamePath + ".jpg")
		}(fileNamePath)

	} else if data.Type == "dismissal" { // Benachrichtigung wurde entfernt

		fileNamePath := filepath.Join(".", fileName)
		removeToast(fileNamePath+"_remove.ps1", map[string]string{"tag": data.PackageName + data.NotificationID + data.SourceDeviceIden + data.SourceUserIden})
		go func(fileNamePath string) {
			time.Sleep(100 * time.Millisecond)
			os.Remove(fileNamePath + "_remove.ps1")
		}(fileNamePath)

	} else {

		log.Println("Default Case:", prettyPrint(data))

	}
}

func saveIcon(data, fileName string) error {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	m, _, err := image.Decode(reader)
	if err != nil {
		return err
	}

	// Encode from image format to writer
	pngFilename := fileName + ".jpg"
	f, err := os.OpenFile(pngFilename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	err = jpeg.Encode(f, m, &jpeg.Options{Quality: 75})
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func prettyPrint(data interface{}) string {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	return string(out)
}
