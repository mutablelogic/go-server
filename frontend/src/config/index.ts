// The server's configuration, from /config.json, which says where its API
// is, and a controller which follows the server, as it goes down, comes
// back, and is upgraded

export { configClient, getConfig, type ConfigOptions } from './httpclient.ts';
export { ConfigController, type ConfigControllerOptions, type ConfigListener, type ConfigServer, type ConfigState, type ServerStatus } from './controller.ts';
export type { Config } from './schema.ts';
