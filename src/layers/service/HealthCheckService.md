# HealthCheckService

At start, loads current health checks. Then starts a timer to check the health of the apps. 

If any of the health check tempaltes are changed, the program is notified of these changes via `NOTIFY` and the health checks are reloaded.

Results are cached inside the application, and are dumped to the database every X minutes.