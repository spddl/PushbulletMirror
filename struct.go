package main

// Notification @ https://github.com/go-toast/toast/blob/master/toast.go#L108
//
// The toast notification data. The following fields are strongly recommended;
//   - AppID
//   - Title
//
// If no toastAudio is provided, then the toast notification will be silent.
// You can set the toast to have a default audio by setting "Audio" to "toast.Default", or if your go app takes
// user-provided input for audio, call the "toast.Audio(name)" func.
//
// The AppID is shown beneath the toast message (in certain cases), and above the notification within the Action
// Center - and is used to group your notifications together. It is recommended that you provide a "pretty"
// name for your app, and not something like "com.example.MyApp".
//
// If no Title is provided, but a Message is, the message will display as the toast notification's title -
// which is a slightly different font style (heavier).
//
// The Icon should be an absolute path to the icon (as the toast is invoked from a temporary path on the user's
// system, not the working directory).
//
// If you would like the toast to call an external process/open a webpage, then you can set ActivationArguments
// to the uri you would like to trigger when the toast is clicked. For example: "https://google.com" would open
// the Google homepage when the user clicks the toast notification.
// By default, clicking the toast just hides/dismisses it.
//
// The following would show a notification to the user letting them know they received an email, and opens
// gmail.com when they click the notification. It also makes the Windows 10 "mail" sound effect.
//
//     toast := toast.Notification{
//         AppID:               "Google Mail",
//         Title:               email.Subject,
//         Message:             email.Preview,
//         Icon:                "C:/Program Files/Google Mail/icons/logo.png",
//         ActivationArguments: "https://gmail.com",
//         Audio:               toast.Mail,
//     }
//
//     err := toast.Push()
type Notification struct {
	// The name of your app. This value shows up in Windows 10's Action Centre, so make it
	// something readable for your users. It can contain spaces, however special characters
	// (eg. é) are not supported.
	AppID string

	// The main title/heading for the toast notification.
	Title string

	// The single/multi line message to display for the toast notification.
	Message string

	// An optional path to an image on the OS to display to the left of the title & message.
	Icon string

	// The type of notification level action (like toast.Action)
	ActivationType string

	// The activation/action arguments (invoked when the user clicks the notification)
	ActivationArguments string

	// Optional action buttons to display below the notification title & message.
	Actions []Action

	// The audio to play when displaying the toast
	Audio string

	// Whether to loop the audio (default false)
	Loop bool

	// How long the toast should show up for (short/long)
	Duration string

	Tag string

	SourceDeviceIden string
}

// Action Defines an actionable button.
//
// See https://msdn.microsoft.com/en-us/windows/uwp/controls-and-patterns/tiles-and-notifications-adaptive-interactive-toasts for more info.
//
// Only protocol type action buttons are actually useful, as there's no way of receiving feedback from the
// user's choice. Examples of protocol type action buttons include: "bingmaps:?q=sushi" to open up Windows 10's
// maps app with a pre-populated search field set to "sushi".
//
//     toast.Action{"protocol", "Open Maps", "bingmaps:?q=sushi"}
type Action struct {
	Type      string
	Label     string
	Arguments string
	// Arguments *url.URL
}

// DismissJSON DismissNotification
type DismissJSON struct {
	Push PushDismissJSON `json:"push"`
	Type string          `json:"type"`
}

// SendJSON DismissNotification
// type SendJSON struct {
// 	Push SendMessageJSON `json:"push"`
// 	Type string          `json:"type"`
// }

// PushDismissJSON DismissNotification
type PushDismissJSON struct {
	NotificationID   string      `json:"notification_id"`
	NotificationTag  interface{} `json:"notification_tag"`
	PackageName      string      `json:"package_name"`
	SourceUserIden   string      `json:"source_user_iden"`
	Type             string      `json:"type"`
	ConversationIden string      `json:"conversation_iden"`
}

// SendMessageJSON XX
// type SendMessageJSON struct {
// 	Type             string `json:"type"`
// 	TargetDeviceIden string `json:"target_device_iden"`
// 	PackageName      string `json:"package_name"`
// 	SourceUserIden   string `json:"source_user_iden"`
// 	ConversationIden string `json:"conversation_iden"`
// 	Message          string `json:"message"`
// }

// JSONEntry xx
type JSONEntry struct {
	Title   string        `json:"title,omitempty"`
	Type    string        `json:"type"`
	Push    JSONPushEntry `json:"push"`
	Targets []string      `json:"targets"`
}

type JSONPushEntry struct {
	Actions []struct {
		Label      string `json:"label"`
		TriggerKey string `json:"trigger_key"`
	} `json:"actions,omitempty"`
	ApplicationName  string `json:"application_name,omitempty"`
	Body             string `json:"body,omitempty"`
	ClientVersion    int    `json:"client_version,omitempty"`
	ConversationIden string `json:"conversation_iden,omitempty"`
	Dismissible      bool   `json:"dismissible,omitempty"`
	Icon             string `json:"icon,omitempty"`
	NotificationID   string `json:"notification_id"`
	NotificationTag  string `json:"notification_tag,omitempty"`
	PackageName      string `json:"package_name"`
	SourceDeviceIden string `json:"source_device_iden,omitempty"`
	SourceUserIden   string `json:"source_user_iden"`
	Title            string `json:"title,omitempty"`
	Type             string `json:"type"`
}
