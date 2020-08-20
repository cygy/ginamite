package notifications

// all the notifications that can be subscribed by a user.
var supportedNotifications = []notificationType{
	notificationType{
		Name:    TypeNewLogin,
		Default: true,
		Enabled: true,
	},
	notificationType{
		Name:    TypeCommercialPartners,
		Default: false,
		Enabled: true,
	},
}
