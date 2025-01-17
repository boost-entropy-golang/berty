export enum GoLogLevel {
	debug = 'debug',
	info = 'info',
	error = 'error',
	warn = 'warn',
}

type GoLoggerOpts = {
	level: GoLogLevel
	message: string
}

export interface GoBridgeInterface {
	log(opts: GoLoggerOpts): void
	initBridge(): Promise<void>
	initBridgeRemote(address: string): Promise<void>
	clearStorage(): Promise<void>
	closeBridge(): Promise<void>
	invokeBridgeMethod(method: string, b64message: string): Promise<string>
	connectService(serviceName: string, address: string): Promise<void>
}
