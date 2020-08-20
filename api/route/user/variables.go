package user

var (
	// GetEnabledUsersForAdmin : function to return the list of the users to display in the admin panel.
	GetEnabledUsersForAdmin GetUsersForAdminFunc

	// GetDisabledUsersForAdmin : function to return the list of the disabled users to display in the admin panel.
	GetDisabledUsersForAdmin GetUsersForAdminFunc

	// UpdatePasswordByAdmin : function to  update the password of a user by a admin.
	UpdatePasswordByAdmin UpdatePasswordByAdminFunc

	// UpdateInformationsByAdmin : function to  update the informations of a user by a admin.
	UpdateInformationsByAdmin UpdateInformationsByAdminFunc

	// UpdateDescriptionByAdmin : function to  update the description of a user by a admin.
	UpdateDescriptionByAdmin UpdateDescriptionByAdminFunc

	// UpdateSocialByAdmin : function to  update the social networks of a user by a admin.
	UpdateSocialByAdmin UpdateSocialByAdminFunc

	// UpdateSettingsByAdmin : function to  update the informations of a user by a admin.
	UpdateSettingsByAdmin UpdateSettingsByAdminFunc

	// UpdateNotificationsByAdmin : function to  update the notifications of a user by a admin.
	UpdateNotificationsByAdmin UpdateNotificationsByAdminFunc

	// GetTokensForAdmin : function to return the list of the user's auth tokens to display in the admin panel.
	GetTokensForAdmin GetTokensForAdminFunc

	// DeleteTokenByIDByAdmin : function to delete a user's auth tokens by an admin.
	DeleteTokenByIDByAdmin DeleteTokenByIDByAdminFunc

	// SetAbilitiesToUser : sets abilities to a user.
	SetAbilitiesToUser SetAbilitiesToUserFunc

	// AddAbilitiesToUser : add some abilities to a user.
	AddAbilitiesToUser AddAbilitiesToUserFunc

	// RemoveAbilitiesFromUser : remove some abilities from a user.
	RemoveAbilitiesFromUser RemoveAbilitiesFromUserFunc
)
