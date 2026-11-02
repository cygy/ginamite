# API features

The default routes of the API are grouped by feature, chosen with `api_features`. Each feature owns its routes of
administration, and needs the storages of `api.Server` they call (`Contacts`, `Terms`, `Users` and the ones of `Account`: `Accounts`,
`Tokens`, `Processes`, `Notifications`; a plugin provides them): the start fails, naming the undefined storages, if an
enabled feature misses some. No route is loaded without feature.

```yaml
api_features: [status, locales, timezones, auth, account, contact]
```

| Feature | Routes |
| --- | --- |
| `status`, `locales`, `timezones` | `GET /status`, the supported locales, the timezones of the countries. They need nothing from the application. |
| `terms` | The terms of use, their version, their acceptance. |
| `notifications` | The notifications of the users: list, center, settings, unsubscription, and their administration. |
| `registration` | The registration with a password, and the availability of the usernames. |
| `forgot_password` | The reset of a forgotten password. |
| `auth` | The login with a password, the auth tokens of the users, and their administration. |
| `facebook`, `google` | The registration and the login with the provider. They need `facebook.app_id` and `facebook.app_secret`, or `google.client_id`. |
| `account` | The account of the users, and the administration of the users: list, update, disable, delete, abilities. |
| `contact` | The contact messages, and their administration. |

`api.Features()` returns the names. A route can still be enabled or moved in `RouterConfigureDefaultRoutes`.

The features other than `status`, `locales` and `timezones` sign and verify the auth tokens, so they need `jwt.secret`:
the start fails without it. Without a secret, the middlewares which authenticate (`MustBeAuthenticated`, `rights`,
etc.) of the custom routes answer an error: set `jwt.secret` as soon as a route authenticates.
