package notifications

// all the notifications that can be subscribed by a user.
var supportedNotifications = []notificationType{
	{
		Name:    TypeNewLogin,
		Default: true,
		Enabled: true,
	},
	{
		Name:    TypeCommercialPartners,
		Default: false,
		Enabled: true,
	},
}
