wort is an experimental data logger for digitemp sensor readings.

It uses the Bolt database to store readings, and serve them up with
a http API.

It receives sensor readings as a POST to the /api/new/ route.

A webapp is served from / and uses the api to retrieve readings from
the database.
