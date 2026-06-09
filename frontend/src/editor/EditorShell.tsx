import Editor, { type OnMount } from "@monaco-editor/react";
import { useCallback, useEffect, useRef, useState, type MutableRefObject } from "react";
import type * as Monaco from "monaco-editor";
import {
	chooseSavePath,
	completeYAML,
	loadSchema,
	saveYAML,
	type CompletionCandidate,
	validateYAML,
} from "../app/api";
import { CloseTabDialog } from "../components/CloseTabDialog";
import { ErrorList, type EditorDiagnostic } from "../components/ErrorList";
import { FileTabs } from "../components/FileTabs";
import { FileToolbar } from "../components/FileToolbar";
import { ScheduleMenu } from "../components/ScheduleMenu";
import { SchemaPane, type SchemaField } from "../components/SchemaPane";
import { dateNextBlockCompletion, dateTemplateInsertion } from "./dateTemplates";
import {
	defaultScheduleTemplate,
	sanitizeScheduleTemplate,
	scheduleTemplateInsertion,
} from "./scheduleTemplates";
import {
	activeTab,
	addUntitledTab,
	closeTab,
	closeConfirmationMessage,
	createInitialTabState,
	isUnsavedTab,
	markActiveTabSaved,
	openDocumentTab,
	switchToAdjacentTab,
	switchTab,
	updateActiveContent,
	updateTabCursor,
	updateTabDiagnostics,
	type TabState,
} from "./tabs";

const scheduleTemplateStorageKey = "yaml-struct-editor.schedule-template";

const sampleSchema: SchemaField = {
	name: "Config",
	type: "struct",
	required: true,
	children: [
		{
			name: "server",
			type: "struct",
			required: true,
			description: "server settings",
			children: [
				{
					name: "host",
					type: "string",
					required: true,
					description: "listen host",
					default: "localhost",
				},
				{
					name: "port",
					type: "int",
					required: true,
					description: "listen port",
					default: "8080",
				},
			],
		},
		{
			name: "app",
			type: "struct",
			required: true,
			description: "application settings",
			children: [
				{
					name: "mode",
					type: "string",
					required: true,
					description: "runtime mode",
					default: "dev",
					enum: ["dev", "stg", "prod"],
				},
			],
		},
	],
};

export function EditorShell() {
	const editorRef = useRef<Monaco.editor.IStandaloneCodeEditor | null>(null);
	const monacoRef = useRef<typeof Monaco | null>(null);
	const completionProviderRef = useRef<Monaco.IDisposable | null>(null);
	const contentChangeRef = useRef<Monaco.IDisposable | null>(null);
	const cursorPositionRef = useRef<Monaco.IDisposable | null>(null);
	const automaticEditRef = useRef(false);
	const validationRequestRef = useRef(0);
	const [tabState, setTabState] = useState<TabState>(() => createInitialTabState());
	const [pendingCloseTabID, setPendingCloseTabID] = useState<string | null>(null);
	const [isScheduleMenuOpen, setIsScheduleMenuOpen] = useState(false);
	const [scheduleTemplate, setScheduleTemplate] = useState(loadScheduleTemplate);
	const scheduleTemplateRef = useRef(scheduleTemplate);
	const [openRequestID, setOpenRequestID] = useState(0);
	const [recentFiles, setRecentFiles] = useState<string[]>(["config.yaml"]);
	const [schema, setSchema] = useState<SchemaField>(sampleSchema);
	const currentTab = activeTab(tabState);
	const pendingCloseTab = pendingCloseTabID
		? tabState.tabs.find((tab) => tab.id === pendingCloseTabID)
		: undefined;
	const content = currentTab.content;
	const diagnostics = currentTab.diagnostics;
	const cursor = currentTab.cursor;

	const runValidation = useCallback(async (tabID: string, nextContent: string) => {
		const requestID = validationRequestRef.current + 1;
		validationRequestRef.current = requestID;
		const nextDiagnostics = await validateYAML(nextContent);
		if (validationRequestRef.current === requestID) {
			setTabState((state) => updateTabDiagnostics(state, tabID, nextDiagnostics));
		}
	}, []);

	const applyMarkers = useCallback((nextDiagnostics: EditorDiagnostic[]) => {
		const editor = editorRef.current;
		const monaco = monacoRef.current;
		const model = editor?.getModel();
		if (!monaco || !model) {
			return;
		}

		monaco.editor.setModelMarkers(
			model,
			"yaml-struct-editor",
			nextDiagnostics.map((diagnostic) => ({
				severity: monaco.MarkerSeverity.Error,
				message: diagnostic.message,
				startLineNumber: diagnostic.line,
				startColumn: diagnostic.column,
				endLineNumber: diagnostic.line,
				endColumn: diagnostic.column + 1,
			})),
		);
	}, []);

	const handleMount: OnMount = (editor, monaco) => {
		editorRef.current = editor;
		monacoRef.current = monaco;
		monaco.editor.defineTheme("yamlStructEditor", {
			base: "vs",
			inherit: true,
			rules: [],
			colors: {
				"editor.background": "#fbfbf8",
				"editorLineNumber.foreground": "#7d8178",
				"editor.selectionBackground": "#cfe5ff",
			},
		});
		monaco.editor.setTheme("yamlStructEditor");
		completionProviderRef.current?.dispose();
		contentChangeRef.current?.dispose();
		cursorPositionRef.current?.dispose();
		completionProviderRef.current = monaco.languages.registerCompletionItemProvider("yaml", {
			triggerCharacters: [" ", "\n", ":", "-"],
			provideCompletionItems: async (
				model: Monaco.editor.ITextModel,
				position: Monaco.Position,
			) => {
				const candidates = await completeYAML(
					model.getValue(),
					position.lineNumber,
					position.column,
				);
				return {
					suggestions: candidates
						.filter((candidate) => candidate.name !== "")
						.map((candidate) => toCompletionItem(monaco, model, position, candidate))
						.concat(dateBlockCompletionItem(monaco, model, position)),
				};
			},
		});
		contentChangeRef.current = editor.onDidChangeModelContent((event) => {
			if (!automaticEditRef.current) {
				for (const change of event.changes) {
					if (change.text.includes("\n")) {
						if (insertCommonTemplate(monaco, editor, automaticEditRef, scheduleTemplateRef.current)) {
							return;
						}
						if (hasDateBlockCompletion(editor)) {
							editor.trigger("yaml-struct-editor", "editor.action.triggerSuggest", null);
							return;
						}
					}
				}
			}

			for (const change of event.changes) {
				if (shouldTriggerSuggest(editor, change.text)) {
					editor.trigger("yaml-struct-editor", "editor.action.triggerSuggest", null);
					break;
				}
			}
		});
		cursorPositionRef.current = editor.onDidChangeCursorPosition((event) => {
			setTabState((state) => updateTabCursor(state, activeTab(state).id, {
				line: event.position.lineNumber,
				column: event.position.column,
			}));
		});
		const position = editor.getPosition();
		if (position) {
			setTabState((state) => updateTabCursor(state, activeTab(state).id, {
				line: position.lineNumber,
				column: position.column,
			}));
		}
		void runValidation(currentTab.id, editor.getValue());
	};

	useEffect(() => {
		return () => {
			completionProviderRef.current?.dispose();
			contentChangeRef.current?.dispose();
			cursorPositionRef.current?.dispose();
		};
	}, []);

	useEffect(() => {
		void loadSchema().then((loadedSchema) => {
			if (loadedSchema) {
				setSchema(loadedSchema);
			}
		});
	}, []);

	useEffect(() => {
		scheduleTemplateRef.current = scheduleTemplate;
	}, [scheduleTemplate]);

	useEffect(() => {
		const timerID = window.setTimeout(() => {
			void runValidation(currentTab.id, content);
		}, 200);
		return () => window.clearTimeout(timerID);
	}, [content, currentTab.id, runValidation]);

	useEffect(() => {
		applyMarkers(diagnostics);
	}, [diagnostics, applyMarkers]);

	useEffect(() => {
		const editor = editorRef.current;
		if (!editor) {
			return;
		}

		editor.setPosition({
			lineNumber: cursor.line,
			column: cursor.column,
		});
		editor.focus();
	}, [currentTab.id, cursor.column, cursor.line]);

	const handleSelectDiagnostic = (diagnostic: EditorDiagnostic) => {
		const editor = editorRef.current;
		if (!editor) {
			return;
		}

		editor.revealPositionInCenter({
			lineNumber: diagnostic.line,
			column: diagnostic.column,
		});
		editor.setPosition({
			lineNumber: diagnostic.line,
			column: diagnostic.column,
		});
		editor.focus();
	};

	const handleNew = useCallback(() => {
		setTabState(addUntitledTab);
	}, []);

	const handleRequestOpen = useCallback(() => {
		setOpenRequestID((requestID) => requestID + 1);
	}, []);

	const handleOpen = useCallback((fileName: string, nextContent: string) => {
		setTabState((state) => openDocumentTab(state, {
			name: fileName,
			content: nextContent,
		}));
		setRecentFiles((files) => [fileName, ...files.filter((file) => file !== fileName)].slice(0, 5));
	}, []);

	const handleSave = useCallback(async () => {
		try {
			const tab = activeTab(tabState);
			const path = tab.path || (await chooseSavePath(tab.name));
			if (!path) {
				return;
			}

			await saveYAML(path, tab.content);
			setTabState((state) => markActiveTabSaved(state, path));
			setRecentFiles((files) => [path, ...files.filter((file) => file !== path)].slice(0, 5));
		} catch (error) {
			window.alert(error instanceof Error ? error.message : "Save failed");
		}
	}, [tabState]);

	const handleSaveScheduleTemplate = useCallback((nextTemplate: string) => {
		const sanitized = sanitizeScheduleTemplate(nextTemplate);
		setScheduleTemplate(sanitized);
		window.localStorage.setItem(scheduleTemplateStorageKey, sanitized);
		setIsScheduleMenuOpen(false);
	}, []);

	const handleResetScheduleTemplate = useCallback(() => {
		setScheduleTemplate(defaultScheduleTemplate);
		window.localStorage.removeItem(scheduleTemplateStorageKey);
	}, []);

	const handleSelectTab = useCallback((tabID: string) => {
		setTabState((state) => switchTab(state, tabID));
	}, []);

	const handleCloseTab = useCallback((tabID: string) => {
		const tab = tabState.tabs.find((candidate) => candidate.id === tabID);
		if (!tab) {
			return;
		}
		if (isUnsavedTab(tab)) {
			setPendingCloseTabID(tabID);
			return;
		}

		setTabState((state) => closeTab(state, tabID));
	}, [tabState.tabs]);

	const handleCloseActiveTab = useCallback(() => {
		handleCloseTab(activeTab(tabState).id);
	}, [handleCloseTab, tabState]);

	const handleCancelCloseTab = useCallback(() => {
		setPendingCloseTabID(null);
	}, []);

	const handleConfirmCloseTab = useCallback(() => {
		if (!pendingCloseTabID) {
			return;
		}
		setTabState((state) => closeTab(state, pendingCloseTabID));
		setPendingCloseTabID(null);
	}, [pendingCloseTabID]);

	const handleSelectAdjacentTab = useCallback((direction: 1 | -1) => {
		setTabState((state) => switchToAdjacentTab(state, direction));
	}, []);

	useEffect(() => {
		const handleShortcut = (event: KeyboardEvent) => {
			if (event.key === "Escape" && pendingCloseTabID) {
				event.preventDefault();
				handleCancelCloseTab();
				return;
			}

			if (!isPrimaryShortcut(event)) {
				return;
			}

			const key = event.key.toLowerCase();
			if (key === "n") {
				event.preventDefault();
				handleNew();
				return;
			}
			if (key === "o") {
				event.preventDefault();
				handleRequestOpen();
				return;
			}
			if (key === "s") {
				event.preventDefault();
				void handleSave();
				return;
			}
			if (key === "w") {
				event.preventDefault();
				handleCloseActiveTab();
				return;
			}
			if (event.key === "Tab") {
				event.preventDefault();
				handleSelectAdjacentTab(event.shiftKey ? -1 : 1);
			}
		};

		window.addEventListener("keydown", handleShortcut);
		return () => window.removeEventListener("keydown", handleShortcut);
	}, [
		handleCancelCloseTab,
		handleCloseActiveTab,
		handleNew,
		handleRequestOpen,
		handleSave,
		handleSelectAdjacentTab,
		pendingCloseTabID,
	]);

	return (
		<main className="app-shell">
			<FileToolbar
				currentFileName={currentTab.name}
				recentFiles={recentFiles}
				onNew={handleNew}
				onOpen={handleOpen}
				onSave={handleSave}
				onSchedules={() => setIsScheduleMenuOpen(true)}
				onUndo={() => editorRef.current?.trigger("toolbar", "undo", null)}
				onRedo={() => editorRef.current?.trigger("toolbar", "redo", null)}
				openRequestID={openRequestID}
			/>
			<FileTabs
				activeTabID={tabState.activeTabID}
				onClose={handleCloseTab}
				onSelect={handleSelectTab}
				tabs={tabState.tabs}
			/>
			<section className="workspace">
				<div className="editor-region">
					<Editor
						height="100%"
						defaultLanguage="yaml"
						onChange={(value) => setTabState((state) => updateActiveContent(state, value ?? ""))}
						onMount={handleMount}
						options={{
							automaticLayout: true,
							folding: true,
							fontFamily: "Menlo, Monaco, Consolas, monospace",
							fontSize: 14,
							lineNumbers: "on",
							minimap: { enabled: false },
							padding: { top: 12, bottom: 12 },
							quickSuggestions: { other: true, comments: false, strings: false },
							scrollBeyondLastLine: false,
							suggestOnTriggerCharacters: true,
							tabSize: 2,
							wordWrap: "on",
						}}
						value={content}
					/>
				</div>
				<SchemaPane root={schema} content={content} cursor={cursor} />
			</section>
			<ErrorList diagnostics={diagnostics} onSelect={handleSelectDiagnostic} />
			{pendingCloseTab ? (
				<CloseTabDialog
					fileName={pendingCloseTab.name}
					message={closeConfirmationMessage(pendingCloseTab)}
					onCancel={handleCancelCloseTab}
					onConfirm={handleConfirmCloseTab}
				/>
			) : null}
			{isScheduleMenuOpen ? (
				<ScheduleMenu
					template={scheduleTemplate}
					onCancel={() => setIsScheduleMenuOpen(false)}
					onReset={handleResetScheduleTemplate}
					onSave={handleSaveScheduleTemplate}
				/>
			) : null}
		</main>
	);
}

function insertCommonTemplate(
	monaco: typeof Monaco,
	editor: Monaco.editor.IStandaloneCodeEditor,
	automaticEditRef: MutableRefObject<boolean>,
	scheduleTemplate: string,
): boolean {
	const model = editor.getModel();
	const position = editor.getPosition();
	if (!model || !position) {
		return false;
	}

	const lines = model.getValue().split(/\r?\n/);
	const insertion = dateTemplateInsertion(lines, position.lineNumber)
		?? scheduleTemplateInsertion(lines, position.lineNumber, scheduleTemplate);
	if (!insertion) {
		return false;
	}

	const range = new monaco.Range(
		position.lineNumber,
		1,
		position.lineNumber,
		model.getLineMaxColumn(position.lineNumber),
	);

	automaticEditRef.current = true;
	editor.executeEdits("yaml-struct-editor", [{ range, text: insertion.text }]);
	automaticEditRef.current = false;
	editor.setPosition({
		lineNumber: position.lineNumber + insertion.cursorLineOffset,
		column: insertion.cursorColumn,
	});
	return true;
}

function loadScheduleTemplate(): string {
	const stored = window.localStorage.getItem(scheduleTemplateStorageKey);
	if (!stored) {
		return defaultScheduleTemplate;
	}
	const sanitized = sanitizeScheduleTemplate(stored);
	return sanitized === "" ? defaultScheduleTemplate : sanitized;
}

function hasDateBlockCompletion(editor: Monaco.editor.IStandaloneCodeEditor): boolean {
	const model = editor.getModel();
	const position = editor.getPosition();
	if (!model || !position) {
		return false;
	}
	return dateNextBlockCompletion(model.getValue().split(/\r?\n/), position.lineNumber) !== null;
}

function dateBlockCompletionItem(
	monaco: typeof Monaco,
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
): Monaco.languages.CompletionItem[] {
	const insertion = dateNextBlockCompletion(model.getValue().split(/\r?\n/), position.lineNumber);
	if (!insertion) {
		return [];
	}

	const range = {
		startLineNumber: position.lineNumber,
		endLineNumber: position.lineNumber,
		startColumn: 1,
		endColumn: model.getLineMaxColumn(position.lineNumber),
	};
	const label = insertion.text.trimStart().split(":")[0] || "next day";
	return [{
		label,
		kind: monaco.languages.CompletionItemKind.Snippet,
		insertText: `${insertion.text}$0`,
		insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
		range,
		detail: "date block | optional",
		documentation: "Insert the next day block.",
		sortText: "0000",
	}];
}

function toCompletionItem(
	monaco: typeof Monaco,
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
	candidate: CompletionCandidate,
): Monaco.languages.CompletionItem {
	const line = model.getLineContent(position.lineNumber);
	const isValue = line.lastIndexOf(":", position.column - 1) >= 0;
	const word = model.getWordUntilPosition(position);
	const range = {
		startLineNumber: position.lineNumber,
		endLineNumber: position.lineNumber,
		startColumn: word.startColumn,
		endColumn: word.endColumn,
	};

	if (isValue) {
		return {
			label: candidate.name,
			kind: monaco.languages.CompletionItemKind.Value,
			insertText: candidate.name,
			range,
			detail: completionDetail(candidate),
			documentation: candidate.description,
		};
	}

	const isContainer = candidate.type === "struct" || candidate.type === "map";
	const defaultValue = candidate.default ?? candidate.enum?.[0] ?? "";
	return {
		label: candidate.name,
		kind: monaco.languages.CompletionItemKind.Property,
		insertText: isContainer
			? `${candidate.name}:\n  $0`
			: `${candidate.name}: ${defaultValue}$0`,
		insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
		range,
		detail: completionDetail(candidate),
		documentation: candidate.description,
	};
}

function completionDetail(candidate: CompletionCandidate): string {
	const parts = [candidate.type, candidate.required ? "required" : "optional"];
	if (candidate.enum && candidate.enum.length > 0) {
		parts.push(`enum: ${candidate.enum.join(", ")}`);
	}
	if (candidate.default) {
		parts.push(`default: ${candidate.default}`);
	}
	return parts.filter(Boolean).join(" | ");
}

function shouldTriggerSuggest(
	editor: Monaco.editor.IStandaloneCodeEditor,
	insertedText: string,
): boolean {
	if (!/^[A-Za-z0-9_-]+$/.test(insertedText)) {
		return false;
	}

	const model = editor.getModel();
	const position = editor.getPosition();
	if (!model || !position) {
		return false;
	}

	const linePrefix = model.getLineContent(position.lineNumber).slice(0, position.column - 1);
	const trimmedPrefix = linePrefix.trimStart();
	if (trimmedPrefix.startsWith("#")) {
		return false;
	}

	if (!linePrefix.includes(":")) {
		return true;
	}
	return /:\s*[A-Za-z0-9_-]*$/.test(linePrefix);
}

function isPrimaryShortcut(event: KeyboardEvent): boolean {
	if (event.altKey) {
		return false;
	}

	const isMac = /Mac|iPhone|iPad/.test(navigator.platform);
	if (isMac) {
		return event.metaKey && !event.ctrlKey;
	}
	return event.ctrlKey && !event.metaKey;
}
