package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"golang.org/x/sys/windows/registry"
)

func pushbulletProtocolCheck(filename string) error {
	key, err := registry.OpenKey(registry.CLASSES_ROOT, `pushbulletapi\shell\open\command`, registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	s, _, err := key.GetStringValue("")
	if err != nil {
		return err
	}

	if strings.HasPrefix(s, "\""+filename+"\" \"%1\"") {
		return nil // korrekter Pfad
	}

	// log.Println("in der Registry gefunden:", s)
	// log.Println("cmd /c start \"PushbulletAPI\" /Min \"" + filename + " \"%1\"")
	return errors.New("Pfad in der Registry ist falsch hinterlegt")
}

func pushbulletProtocolCreateReg(filename string) { // https://www.robvanderwoude.com/ntstart.php
	templatePropertyData := `Windows Registry Editor Version 5.00

[HKEY_CLASSES_ROOT\PushbulletApi]
@="URL:PushbulletApi"
"URL Protocol"=""

[HKEY_CLASSES_ROOT\PushbulletApi\shell]

[HKEY_CLASSES_ROOT\PushbulletApi\shell\open]

[HKEY_CLASSES_ROOT\PushbulletApi\shell\open\command]
@="{{ .}} \"%1\""`

	tmplProperty := template.Must(template.New("template").Parse(string(templatePropertyData))) // erstellt aus den Daten das Template

	var buf bytes.Buffer
	err := tmplProperty.Execute(&buf, filename) // schreibt das Template mit den Daten in die Application.ini
	if err != nil {
		log.Fatal(err)
	}
	readBuf, err := ioutil.ReadAll(&buf)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./PushbulletApi.reg", readBuf, 0644)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("REG", "IMPORT", "PushbulletApi.reg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Printf("REG IMPORT PushbulletApi.reg\nfailed with %s\n", err)
	} else {
		err = os.Remove("./PushbulletApi.reg")
		if err != nil {
			log.Fatal(err)
		}
	}
}
