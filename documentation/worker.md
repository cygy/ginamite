# Worker

The worker server (`-worker`) consumes the messages the API sends, and runs the recurring tasks. It has no HTTP route.
Its storages and hooks are `worker.Server.Storage` and `worker.Server.Hooks`, set by the storage plugin and by the
application.

## Messages

The worker consumes these topics, or the ones listed in `worker_consumers`, see
[Message broker and Kafka](kafka.md#workers-and-kafka-consumer-groups):

| Topics | Handled |
| --- | --- |
| `task`, `task_location`, `task_user`, `task_registration` | The built-in tasks: location of an IP address, deletion, disabling and enabling of a user, update of the social networks, registration done and validated. |
| `mail` | The emails of the registration, the forgotten password, the changes of email address and password, and the deletion, disabling and enabling of an account. |
| `user_notification`, `group_notification` | The notifications of a user, or of the users with an ability. |

A handler which fails makes the message be handled again: the handlers must tolerate it, see
[Message broker and Kafka](kafka.md).

## Recurring tasks

Each task runs every `interval` seconds, from `start_at` (an hour in UTC), if it is `enabled`. A task runs one at a time:
a run longer than the interval is followed by the next one at once. The panic of a run is logged and does not stop the
worker. Each worker runs its own tasks: with several workers, enable the recurring tasks on one of them only.

```yaml
recurring_tasks:
  - { name: delete_expired_tokens,     interval: 3600,  enabled: true }
  - { name: delete_never_used_accounts, interval: 86400, start_at: "03:00", enabled: true }
  - { name: delete_inactive_accounts,  interval: 86400, start_at: "03:30", enabled: true }
  - { name: sanitize_accounts,         interval: 86400, start_at: "04:00", enabled: true }
  - { name: generate_sitemaps,         interval: 86400, start_at: "05:00", enabled: true }
```

A task whose storage is not set is not scheduled, with a warning; `delete_never_used_accounts` also needs `account.email_address_must_be_confirmed`. The accounts are deleted after `account.delete_never_used_after` days without sign-in and `account.delete_inactive_after` months of inactivity. `Hooks.SetFuncOfRecurringTasks` adds the tasks of
the application, which are then enabled in the same list: a `recurring.Func` receives the context of the worker, done
when it stops, and the worker waits for the running tasks before it returns, so a long task must return once its
context is done. `run_start_up_tasks` runs `Hooks.StartUpTasks` when the worker
starts.

## Storage

`worker.Storage` is what a database plugin implements for the worker; a field left `nil` disables what needs it.

| Field | Used for |
| --- | --- |
| `Tokens` | Deleting the expired auth tokens. |
| `Accounts` | The never used and inactive accounts, and their sanitizing. |
| `Users` | Deleting, disabling and enabling a user. |
| `Notifications` | The targets of the notifications of a user. |
| `Locations` | Saving the location of an IP address. |
| `Emails` | Checking an email address before sending. |
| `Sitemap` | The variables of the routes of the sitemaps. |

## Hooks

`worker.Hooks` are all optional: `StartUpTasks`, `SetFuncOfRecurringTasks`, `RegistrationDone`,
`RegistrationValidated`, and `HandleTask`, `HandleUserNotification`, `HandleMail` and
`HandleNotificationForUsersWithAbility`, which handle the custom messages of their topic and report whether they did.
Custom tasks are sent with `Queue.CreateTask`.
