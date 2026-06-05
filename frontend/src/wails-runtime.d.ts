declare module "/wails/runtime.js" {
	export const Call: {
		ByName: (methodName: string, ...args: unknown[]) => Promise<unknown>;
	};
	export const Dialogs: {
		SaveFile: (options: {
			Title?: string;
			Filename?: string;
			ButtonText?: string;
			CanCreateDirectories?: boolean;
			Filters?: Array<{
				DisplayName?: string;
				Pattern?: string;
			}>;
		}) => Promise<string>;
	};
}
