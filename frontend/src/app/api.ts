import type { EditorDiagnostic } from "../components/ErrorList";
import type { SchemaField } from "../components/SchemaPane";

export type CompletionCandidate = {
	name: string;
	type: string;
	description?: string;
	required?: boolean;
	default?: string;
	enum?: string[];
	root?: boolean;
	children?: CompletionCandidate[];
	item?: CompletionCandidate;
	mapValue?: CompletionCandidate;
};

type RuntimeModule = {
	Call?: {
		ByName?: (methodName: string, ...args: unknown[]) => Promise<unknown>;
	};
	Dialogs?: {
		SaveFile?: (options: SaveFileDialogOptions) => Promise<string>;
	};
};

const serviceName = "github.com/mryskyj/yaml-editor/app.App";

let runtimePromise: Promise<RuntimeModule | null> | null = null;

type SaveFileDialogOptions = {
	Title?: string;
	Filename?: string;
	ButtonText?: string;
	CanCreateDirectories?: boolean;
	Filters?: Array<{
		DisplayName?: string;
		Pattern?: string;
	}>;
};

export async function validateYAML(content: string): Promise<EditorDiagnostic[]> {
	const result = await callBackend(`${serviceName}.ValidateYAML`, content);
	if (!Array.isArray(result)) {
		return [];
	}
	return result.map(normalizeDiagnostic);
}

export async function completeYAML(
	content: string,
	line: number,
	column: number,
): Promise<CompletionCandidate[]> {
	const result = await callBackend(`${serviceName}.CompleteYAML`, content, line, column);
	if (!Array.isArray(result)) {
		return [];
	}
	return result.map(normalizeCandidate);
}

export async function loadSchema(): Promise<SchemaField | null> {
	const result = await callBackend(`${serviceName}.Schema`);
	if (!result || typeof result !== "object") {
		return null;
	}
	return normalizeSchema(result);
}

export async function saveYAML(path: string, content: string): Promise<void> {
	await callRequiredBackend(`${serviceName}.SaveFile`, path, content);
}

export async function chooseSavePath(filename: string): Promise<string> {
	const runtime = await loadRuntime();
	if (!runtime?.Dialogs?.SaveFile) {
		throw new Error("Wails save dialog is not available");
	}

	return runtime.Dialogs.SaveFile({
		Title: "Save YAML File",
		Filename: filename || "config.yaml",
		ButtonText: "Save",
		CanCreateDirectories: true,
		Filters: [
			{ DisplayName: "YAML Files", Pattern: "*.yaml;*.yml" },
			{ DisplayName: "All Files", Pattern: "*" },
		],
	});
}

async function callBackend(methodName: string, ...args: unknown[]): Promise<unknown> {
	const runtime = await loadRuntime();
	if (!runtime?.Call?.ByName) {
		return null;
	}
	return runtime.Call.ByName(methodName, ...args);
}

async function callRequiredBackend(methodName: string, ...args: unknown[]): Promise<unknown> {
	const runtime = await loadRuntime();
	if (!runtime?.Call?.ByName) {
		throw new Error("Wails runtime is not available");
	}
	return runtime.Call.ByName(methodName, ...args);
}

async function loadRuntime(): Promise<RuntimeModule | null> {
	if (!runtimePromise) {
		const importRuntime = new Function("return import('/wails/runtime.js')") as () => Promise<unknown>;
		runtimePromise = importRuntime()
			.then((module) => module as RuntimeModule)
			.catch(() => null);
	}
	return runtimePromise;
}

function normalizeDiagnostic(value: unknown): EditorDiagnostic {
	const record = asRecord(value);
	return {
		severity: "error",
		message: String(record.message ?? record.Message ?? ""),
		line: numberValue(record.line ?? record.Line, 1),
		column: numberValue(record.column ?? record.Column, 1),
	};
}

function normalizeCandidate(value: unknown): CompletionCandidate {
	const record = asRecord(value);
	return {
		name: String(record.name ?? record.Name ?? ""),
		type: String(record.type ?? record.Type ?? ""),
		description: stringValue(record.description ?? record.Description),
		required: Boolean(record.required ?? record.Required ?? false),
		default: stringValue(record.default ?? record.Default),
		enum: arrayValue(record.enum ?? record.Enum).map(String),
		root: Boolean(record.root ?? record.Root ?? false),
		children: arrayValue(record.children ?? record.Children).map(normalizeCandidate),
		item: optionalCandidate(record.item ?? record.Item),
		mapValue: optionalCandidate(record.mapValue ?? record.MapValue),
	};
}

function optionalCandidate(value: unknown): CompletionCandidate | undefined {
	if (!value || typeof value !== "object") {
		return undefined;
	}
	return normalizeCandidate(value);
}

function normalizeSchema(value: unknown): SchemaField {
	const record = asRecord(value);
	return {
		name: String(record.name ?? record.Name ?? ""),
		type: String(record.type ?? record.Type ?? ""),
		required: Boolean(record.required ?? record.Required ?? false),
		description: stringValue(record.description ?? record.Description),
		default: stringValue(record.default ?? record.Default),
		enum: arrayValue(record.enum ?? record.Enum).map(String),
		children: arrayValue(record.children ?? record.Children).map(normalizeSchema),
		item: optionalSchema(record.item ?? record.Item),
		mapValue: optionalSchema(record.mapValue ?? record.MapValue),
	};
}

function optionalSchema(value: unknown): SchemaField | undefined {
	if (!value || typeof value !== "object") {
		return undefined;
	}
	return normalizeSchema(value);
}

function asRecord(value: unknown): Record<string, unknown> {
	if (value && typeof value === "object") {
		return value as Record<string, unknown>;
	}
	return {};
}

function numberValue(value: unknown, fallback: number): number {
	if (typeof value === "number" && Number.isFinite(value)) {
		return value;
	}
	return fallback;
}

function stringValue(value: unknown): string | undefined {
	if (typeof value === "string" && value !== "") {
		return value;
	}
	return undefined;
}

function arrayValue(value: unknown): unknown[] {
	return Array.isArray(value) ? value : [];
}
