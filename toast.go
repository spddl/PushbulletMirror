package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"syscall"
	"text/template"
)

// https://docs.microsoft.com/en-us/uwp/api/windows.ui.notifications.toastnotificationhistory.remove?view=winrt-19041
// https://docs.microsoft.com/de-de/windows/uwp/design/shell/tiles-and-notifications/adaptive-interactive-toasts
// https://docs.microsoft.com/de-de/windows/uwp/design/shell/tiles-and-notifications/send-local-toast
// https://docs.microsoft.com/en-us/uwp/api/windows.ui.notifications.toastnotification.tag?view=winrt-19041
// https://docs.microsoft.com/en-us/windows/uwp/design/shell/tiles-and-notifications/toast-pending-update
func pushToast(fileName string, data Notification) {
	log.Println("Erstelle Push Toast:", fileName)
	// $app = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'

	toastTemplate := template.New("toast")
	_, err := toastTemplate.Parse(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$template = @"
<toast activationType="protocol" duration="Short">
    <visual>
        <binding template="ToastGeneric">
            {{if .Icon}}<image placement="appLogoOverride" hint-crop="circle" src="{{.Icon}}" />{{end}}
            {{if .Title}}<text><![CDATA[{{.Title}}]]></text>{{end}}
            {{if .Message}}<text><![CDATA[{{.Message}}]]></text>{{end}}
        </binding>
    </visual>
  <audio silent="true" />
    {{if .Actions}}<actions>{{range .Actions}}
        <action activationType="{{.Type}}" content="{{.Label}}" arguments="{{.Arguments}}" />{{end}}
    </actions>{{end}}
</toast>
"@
$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($template)
$toast = New-Object Windows.UI.Notifications.ToastNotification $xml
$toast.Tag = "{{.Tag}}"
$toast.Group = "{{.Tag}}"
$toast.ExpirationTime = [DateTimeOffset]::Now.AddMinutes(5)
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Pushbullet")
$notifier.Show($toast);`)
	if err != nil {
		log.Println(err)
	}

	var out bytes.Buffer
	err = toastTemplate.Execute(&out, data)
	if err != nil {
		log.Println(err)
	}

	err = saveToast(fileName, out.String())
	if err != nil {
		log.Println(err)
	}
}

func removeToast(fileName string, data map[string]string) {
	log.Println("Remove Toast:", fileName, data["tag"])
	// log.Println(prettyPrint(data))
	err := saveToast(fileName, fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.UI.Notifications.ToastNotificationManager]::History.Remove("%v", "%v", "Pushbullet")`, data["tag"], data["tag"]))
	if err != nil {
		log.Println(err)
	}
}

func saveToast(fileName, content string) error {
	// log.Println("SaveToast", content)

	bomUtf8 := []byte{0xEF, 0xBB, 0xBF}
	out := append(bomUtf8, []byte(content)...)
	err := ioutil.WriteFile(fileName, out, 0600)
	if err != nil {
		return err
	}
	cmd := exec.Command("PowerShell", "-ExecutionPolicy", "Bypass", "-File", fileName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}
