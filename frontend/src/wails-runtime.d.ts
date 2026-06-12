declare module "/wails/runtime.js" {
	export const Call: {
		ByName: (methodName: string, ...args: unknown[]) => Promise<unknown>;
	};
	export const Dialogs: {
		Error: (options: {
			Title?: string;
			Message?: string;
			Buttons?: Array<{
				Label: string;
				IsCancel?: boolean;
				IsDefault?: boolean;
			}>;
		}) => Promise<string>;
		OpenFile: (options: {
			Title?: string;
			ButtonText?: string;
			CanChooseFiles?: boolean;
			CanChooseDirectories?: boolean;
			AllowsMultipleSelection?: boolean;
			Filters?: Array<{
				DisplayName?: string;
				Pattern?: string;
			}>;
		}) => Promise<string | string[]>;
		Question: (options: {
			Title?: string;
			Message?: string;
			Buttons?: Array<{
				Label: string;
				IsCancel?: boolean;
				IsDefault?: boolean;
			}>;
		}) => Promise<string>;
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
