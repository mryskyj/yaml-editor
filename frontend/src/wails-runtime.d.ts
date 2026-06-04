declare module "/wails/runtime.js" {
	export const Call: {
		ByName: (methodName: string, ...args: unknown[]) => Promise<unknown>;
	};
}
