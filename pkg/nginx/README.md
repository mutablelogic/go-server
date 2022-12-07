# nginx plugin

The `nginx` plugin provides controls an nginx server, with the following operations:

  * Start and restart server
  * Reload server (for rotating the log files)
  * Test configuration without restarting the server
  * Enumeration of "available" configuration templates
  * Enabling of configurations with templating

The main use case is to provide a controllable nginx server in a Docker container, with an 
API which can be used to control the server. The plugin should be used with the following two
additional plugins:

  * `nginx-gateway`: This plugin provides a REST API gateway to the nginx plugin, and can have
    authentication controls added.
  * `nginx-client`: Can be intergated into an API client, for control of nginx through
    a client application.

