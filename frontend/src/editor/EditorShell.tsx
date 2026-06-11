import Editor, { type OnMount } from "@monaco-editor/react";
import {
	useCallback,
	useEffect,
	useRef,
	useState,
	type MouseEvent as ReactMouseEvent,
	type MutableRefObject,
} from "react";
import type * as Monaco from "monaco-editor";
import {
	chooseSavePath,
	completeYAML,
	loadDefaultScheduleTemplate,
	loadRecentFiles,
	loadRootSchema,
	loadSchema,
	newYAML,
	openYAML,
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
import { dateNextBlockCompletion, dateTemplateInsertion, datesSiblingIndent } from "./dateTemplates";
import { listNextItemCompletion, listNextItemCompletions } from "./listTemplates";
import {
	sanitizeScheduleTemplate,
	scheduleTemplateInsertion,
} from "./scheduleTemplates";
import { stepTemplateInsertion } from "./stepTemplates";
import {
	activeTab,
	addUntitledTab,
	closeTab,
	closeConfirmationMessage,
	createInitialTabState,
	fillActiveUntouchedUntitledTab,
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

const sampleRootSchema: SchemaField = {
	name: "File",
	type: "struct",
	required: true,
	children: [
		{
			name: "scenario",
			type: "struct",
			required: true,
			children: [
				{ name: "id", type: "int", required: true },
				{ name: "name", type: "string", required: true },
				{ name: "description", type: "string", required: false },
				{
					name: "docs",
					type: "slice",
					required: false,
					item: {
						name: "doc",
						type: "struct",
						required: true,
						children: [{ name: "name", type: "string", required: true }],
					},
				},
				{
					name: "steps",
					type: "slice",
					required: true,
					item: {
						name: "step",
						type: "struct",
						required: true,
						children: [
							{ name: "id", type: "string", required: true },
							{ name: "name", type: "string", required: true },
							{ name: "day_ref", type: "string", required: false },
							{ name: "schedule_ref", type: "string", required: false },
							{
								name: "action",
								type: "struct",
								required: true,
								children: [
									{ name: "tool", type: "string", required: true },
									{ name: "user", type: "string", required: false },
									{ name: "password", type: "string", required: false },
									{ name: "path", type: "string", required: false },
									{ name: "args", type: "map", required: false },
								],
							},
						],
					},
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
	const selectionKeyRef = useRef<Monaco.IDisposable | null>(null);
	const restoredTabIDRef = useRef<string | null>(null);
	const automaticEditRef = useRef(false);
	const validationRequestRef = useRef(0);
	const [tabState, setTabState] = useState<TabState>(() => createInitialTabState());
	const [pendingCloseTabID, setPendingCloseTabID] = useState<string | null>(null);
	const [isScheduleMenuOpen, setIsScheduleMenuOpen] = useState(false);
	const [defaultScheduleTemplate, setDefaultScheduleTemplate] = useState("");
	const [scheduleTemplate, setScheduleTemplate] = useState(() => loadStoredScheduleTemplate() ?? "");
	const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null);
	const scheduleTemplateRef = useRef(scheduleTemplate);
	const [openRequestID, setOpenRequestID] = useState(0);
	const [recentFiles, setRecentFiles] = useState<string[]>([]);
	const [newDocumentContent, setNewDocumentContent] = useState("");
	const [schema, setSchema] = useState<SchemaField>(sampleSchema);
	const [rootSchema, setRootSchema] = useState<SchemaField>(sampleRootSchema);
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

	const refreshRecentFiles = useCallback(async () => {
		setRecentFiles(await loadRecentFiles());
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
		selectionKeyRef.current?.dispose();
		completionProviderRef.current = monaco.languages.registerCompletionItemProvider("yaml", {
			triggerCharacters: [" ", "\n", ":", "-", "."],
			provideCompletionItems: async (
				model: Monaco.editor.ITextModel,
				position: Monaco.Position,
			) => {
				const candidates = await completeYAML(
					model.getValue(),
					position.lineNumber,
					position.column,
				);
				const suggestions = await Promise.all(candidates
					.filter((candidate) => candidate.name !== "")
					.map((candidate, index) => toCompletionItem(monaco, model, position, candidate, index)));
				return {
					suggestions: suggestions
						.concat(dateBlockCompletionItems(monaco, model, position, scheduleTemplateRef.current))
						.concat(listItemCompletionItem(monaco, model, position, suggestions.length)),
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
						if (hasListItemCompletion(editor)) {
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
		selectionKeyRef.current = editor.onKeyDown((event) => {
			handleShiftArrowSelection(editor, monaco, event);
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
			selectionKeyRef.current?.dispose();
		};
	}, []);

	useEffect(() => {
		void loadDefaultScheduleTemplate().then(async (template) => {
			const sanitized = sanitizeScheduleTemplate(template);
			const stored = loadStoredScheduleTemplate();
			const effectiveScheduleTemplate = stored ?? sanitized;
			setDefaultScheduleTemplate(sanitized);
			if (stored === null) {
				setScheduleTemplate(effectiveScheduleTemplate);
			}
			const document = await newYAML();
			const content = applyScheduleTemplate(document.content, effectiveScheduleTemplate);
			setNewDocumentContent(content);
			setTabState((state) => fillActiveUntouchedUntitledTab(state, content));
		});
		void loadSchema().then((loadedSchema) => {
			if (loadedSchema) {
				setSchema(loadedSchema);
			}
		});
		void loadRootSchema().then((loadedSchema) => {
			if (loadedSchema) {
				setRootSchema(loadedSchema);
			}
		});
		void refreshRecentFiles();
	}, [refreshRecentFiles]);

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
		if (restoredTabIDRef.current === currentTab.id) {
			return;
		}

		restoredTabIDRef.current = currentTab.id;
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

	const handleNew = useCallback(async () => {
		try {
			const document = await newYAML();
			const content = applyScheduleTemplate(document.content, scheduleTemplate);
			setNewDocumentContent(content);
			setTabState((state) => addUntitledTab(state, content));
		} catch {
			setTabState((state) => addUntitledTab(state, newDocumentContent));
		}
	}, [newDocumentContent, scheduleTemplate]);

	const handleRequestOpen = useCallback(() => {
		setOpenRequestID((requestID) => requestID + 1);
	}, []);

	const handleOpen = useCallback((fileName: string, nextContent: string) => {
		setTabState((state) => openDocumentTab(state, {
			name: fileName,
			content: nextContent,
		}));
	}, []);

	const handleOpenRecent = useCallback(async (path: string) => {
		try {
			const document = await openYAML(path);
			setTabState((state) => openDocumentTab(state, document));
			await refreshRecentFiles();
		} catch (error) {
			window.alert(error instanceof Error ? error.message : "Open failed");
			await refreshRecentFiles();
		}
	}, [refreshRecentFiles]);

	const handleSave = useCallback(async () => {
		try {
			const tab = activeTab(tabState);
			const path = tab.path || (await chooseSavePath(tab.name));
			if (!path) {
				return;
			}

			await saveYAML(path, tab.content);
			setTabState((state) => markActiveTabSaved(state, path));
			await refreshRecentFiles();
		} catch (error) {
			window.alert(error instanceof Error ? error.message : "Save failed");
		}
	}, [refreshRecentFiles, tabState]);

	const handleCopy = useCallback(() => {
		setContextMenu(null);
		void copyEditorSelection(editorRef.current).catch(showClipboardError);
	}, []);

	const handleCut = useCallback(() => {
		setContextMenu(null);
		void cutEditorSelection(editorRef.current, automaticEditRef).catch(showClipboardError);
	}, []);

	const handlePaste = useCallback(() => {
		setContextMenu(null);
		void pasteIntoEditor(editorRef.current, automaticEditRef).catch(showClipboardError);
	}, []);

	const handleEditorContextMenu = useCallback((event: ReactMouseEvent<HTMLDivElement>) => {
		event.preventDefault();
		editorRef.current?.focus();
		setContextMenu({
			x: Math.min(event.clientX, window.innerWidth - 144),
			y: Math.min(event.clientY, window.innerHeight - 112),
		});
	}, []);

	const handleSaveScheduleTemplate = useCallback((nextTemplate: string) => {
		const sanitized = sanitizeScheduleTemplate(nextTemplate);
		setScheduleTemplate(sanitized);
		window.localStorage.setItem(scheduleTemplateStorageKey, sanitized);
		setIsScheduleMenuOpen(false);
	}, []);

	const handleResetScheduleTemplate = useCallback(() => {
		setScheduleTemplate(defaultScheduleTemplate);
		window.localStorage.removeItem(scheduleTemplateStorageKey);
	}, [defaultScheduleTemplate]);

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

		setTabState((state) => closeTab(state, tabID, newDocumentContent));
	}, [newDocumentContent, tabState.tabs]);

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
		setTabState((state) => closeTab(state, pendingCloseTabID, newDocumentContent));
		setPendingCloseTabID(null);
	}, [newDocumentContent, pendingCloseTabID]);

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
			if (event.key === "Escape" && contextMenu) {
				event.preventDefault();
				setContextMenu(null);
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
		contextMenu,
		handleCancelCloseTab,
		handleCloseActiveTab,
		handleNew,
		handleRequestOpen,
		handleSave,
		handleSelectAdjacentTab,
		pendingCloseTabID,
	]);

	useEffect(() => {
		if (!contextMenu) {
			return;
		}
		const closeContextMenu = () => setContextMenu(null);
		window.addEventListener("click", closeContextMenu);
		window.addEventListener("blur", closeContextMenu);
		return () => {
			window.removeEventListener("click", closeContextMenu);
			window.removeEventListener("blur", closeContextMenu);
		};
	}, [contextMenu]);

	return (
		<main className="app-shell">
			<FileToolbar
					currentFileName={currentTab.name}
					recentFiles={recentFiles}
					onNew={handleNew}
					onOpen={handleOpen}
					onOpenRecent={handleOpenRecent}
					onSave={handleSave}
				onSchedules={() => setIsScheduleMenuOpen(true)}
				onCopy={handleCopy}
				onCut={handleCut}
				onPaste={handlePaste}
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
				<div className="editor-region" onContextMenu={handleEditorContextMenu}>
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
							wordBasedSuggestions: "off",
							contextmenu: false,
						}}
						value={content}
					/>
				</div>
				<SchemaPane root={schema} documentRoot={rootSchema} content={content} cursor={cursor} />
			</section>
			{contextMenu ? (
				<div
					className="editor-context-menu"
					style={{ left: contextMenu.x, top: contextMenu.y }}
					onClick={(event) => event.stopPropagation()}
				>
					<button type="button" onClick={handleCut}>Cut</button>
					<button type="button" onClick={handleCopy}>Copy</button>
					<button type="button" onClick={handlePaste}>Paste</button>
				</div>
			) : null}
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
		?? scheduleTemplateInsertion(lines, position.lineNumber, scheduleTemplate)
		?? stepTemplateInsertion(lines, position.lineNumber);
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

function loadStoredScheduleTemplate(): string | null {
	const stored = window.localStorage.getItem(scheduleTemplateStorageKey);
	if (!stored) {
		return null;
	}
	const sanitized = sanitizeScheduleTemplate(stored);
	return sanitized === "" ? null : sanitized;
}

function applyScheduleTemplate(content: string, scheduleTemplate: string): string {
	const entries = sanitizeScheduleTemplate(scheduleTemplate)
		.split("\n")
		.filter((line) => line.trim() !== "");
	if (entries.length === 0) {
		return content;
	}

	const lines = content.replace(/\r\n/g, "\n").split("\n");
	const scheduleIndex = lines.findIndex((line) => line.trim() === "schedules:");
	if (scheduleIndex < 0) {
		return content;
	}

	const indent = lines[scheduleIndex].match(/^\s*/)?.[0] ?? "";
	const childIndent = `${indent}  `;
	let endIndex = scheduleIndex + 1;
	for (; endIndex < lines.length; endIndex++) {
		const line = lines[endIndex] ?? "";
		if (line.trim() === "") {
			continue;
		}
		const lineIndent = line.match(/^\s*/)?.[0].length ?? 0;
		if (lineIndent <= indent.length) {
			break;
		}
	}

	lines.splice(scheduleIndex + 1, endIndex - scheduleIndex - 1, ...entries.map((line) => `${childIndent}${line}`));
	return lines.join("\n");
}

async function copyEditorSelection(editor: Monaco.editor.IStandaloneCodeEditor | null): Promise<void> {
	const text = selectedOrCurrentLineText(editor);
	if (text === null) {
		return;
	}
	await writeClipboardText(text);
	editor?.focus();
}

async function cutEditorSelection(
	editor: Monaco.editor.IStandaloneCodeEditor | null,
	automaticEditRef: MutableRefObject<boolean>,
): Promise<void> {
	const text = selectedOrCurrentLineText(editor);
	if (text === null || !editor) {
		return;
	}

	await writeClipboardText(text);
	deleteSelectedOrCurrentLine(editor, automaticEditRef);
	editor.focus();
}

async function pasteIntoEditor(
	editor: Monaco.editor.IStandaloneCodeEditor | null,
	automaticEditRef: MutableRefObject<boolean>,
): Promise<void> {
	if (!editor) {
		return;
	}

	const text = await readClipboardText();
	if (text === "") {
		editor.focus();
		return;
	}

	const selections = editor.getSelections() ?? [];
	if (selections.length === 0) {
		editor.focus();
		return;
	}

	editor.pushUndoStop();
	automaticEditRef.current = true;
	editor.executeEdits("toolbar-paste", selections.map((selection) => ({
		range: selection,
		text,
		forceMoveMarkers: true,
	})));
	automaticEditRef.current = false;
	editor.pushUndoStop();
	editor.focus();
}

function selectedOrCurrentLineText(editor: Monaco.editor.IStandaloneCodeEditor | null): string | null {
	const model = editor?.getModel();
	const selection = editor?.getSelection();
	if (!model || !selection) {
		return null;
	}

	if (!selection.isEmpty()) {
		return model.getValueInRange(selection);
	}

	return `${model.getLineContent(selection.startLineNumber)}${model.getEOL()}`;
}

function deleteSelectedOrCurrentLine(
	editor: Monaco.editor.IStandaloneCodeEditor,
	automaticEditRef: MutableRefObject<boolean>,
): void {
	const model = editor.getModel();
	const selection = editor.getSelection();
	if (!model || !selection) {
		return;
	}

	const range = selection.isEmpty()
		? currentLineRange(model, selection.startLineNumber)
		: selection;

	editor.pushUndoStop();
	automaticEditRef.current = true;
	editor.executeEdits("toolbar-cut", [{ range, text: "", forceMoveMarkers: true }]);
	automaticEditRef.current = false;
	editor.pushUndoStop();
}

function currentLineRange(model: Monaco.editor.ITextModel, lineNumber: number): Monaco.IRange {
	if (lineNumber < model.getLineCount()) {
		return {
			startLineNumber: lineNumber,
			startColumn: 1,
			endLineNumber: lineNumber + 1,
			endColumn: 1,
		};
	}
	return {
		startLineNumber: lineNumber,
		startColumn: 1,
		endLineNumber: lineNumber,
		endColumn: model.getLineMaxColumn(lineNumber),
	};
}

async function writeClipboardText(text: string): Promise<void> {
	if (!navigator.clipboard?.writeText) {
		throw new Error("Clipboard write is not available");
	}
	await navigator.clipboard.writeText(text);
}

async function readClipboardText(): Promise<string> {
	if (!navigator.clipboard?.readText) {
		throw new Error("Clipboard read is not available");
	}
	return navigator.clipboard.readText();
}

function showClipboardError(error: unknown): void {
	window.alert(error instanceof Error ? error.message : "Clipboard operation failed");
}

function handleShiftArrowSelection(
	editor: Monaco.editor.IStandaloneCodeEditor,
	monaco: typeof Monaco,
	event: Monaco.IKeyboardEvent,
): void {
	if (!event.shiftKey || event.altKey || event.ctrlKey || event.metaKey) {
		return;
	}

	const model = editor.getModel();
	const selection = editor.getSelection();
	if (!model || !selection) {
		return;
	}

	let nextPosition: { lineNumber: number; column: number } | null = null;
	switch (event.keyCode) {
		case monaco.KeyCode.LeftArrow:
			nextPosition = previousPosition(model, selection.positionLineNumber, selection.positionColumn);
			break;
		case monaco.KeyCode.RightArrow:
			nextPosition = nextPositionInModel(model, selection.positionLineNumber, selection.positionColumn);
			break;
		case monaco.KeyCode.UpArrow:
			nextPosition = verticalPosition(model, selection.positionLineNumber - 1, selection.positionColumn);
			break;
		case monaco.KeyCode.DownArrow:
			nextPosition = verticalPosition(model, selection.positionLineNumber + 1, selection.positionColumn);
			break;
		default:
			return;
	}

	if (!nextPosition) {
		return;
	}

	event.preventDefault();
	event.stopPropagation();
	editor.setSelection(new monaco.Selection(
		selection.selectionStartLineNumber,
		selection.selectionStartColumn,
		nextPosition.lineNumber,
		nextPosition.column,
	));
	editor.revealPositionInCenterIfOutsideViewport(nextPosition);
}

function previousPosition(
	model: Monaco.editor.ITextModel,
	lineNumber: number,
	column: number,
): { lineNumber: number; column: number } | null {
	if (column > 1) {
		return { lineNumber, column: column - 1 };
	}
	if (lineNumber <= 1) {
		return null;
	}
	const previousLine = lineNumber - 1;
	return { lineNumber: previousLine, column: model.getLineMaxColumn(previousLine) };
}

function nextPositionInModel(
	model: Monaco.editor.ITextModel,
	lineNumber: number,
	column: number,
): { lineNumber: number; column: number } | null {
	const maxColumn = model.getLineMaxColumn(lineNumber);
	if (column < maxColumn) {
		return { lineNumber, column: column + 1 };
	}
	if (lineNumber >= model.getLineCount()) {
		return null;
	}
	return { lineNumber: lineNumber + 1, column: 1 };
}

function verticalPosition(
	model: Monaco.editor.ITextModel,
	lineNumber: number,
	column: number,
): { lineNumber: number; column: number } | null {
	if (lineNumber < 1 || lineNumber > model.getLineCount()) {
		return null;
	}
	return { lineNumber, column: Math.min(column, model.getLineMaxColumn(lineNumber)) };
}

function hasDateBlockCompletion(editor: Monaco.editor.IStandaloneCodeEditor): boolean {
	const model = editor.getModel();
	const position = editor.getPosition();
	if (!model || !position) {
		return false;
	}
	return dateNextBlockCompletion(model.getValue().split(/\r?\n/), position.lineNumber) !== null;
}

function hasListItemCompletion(editor: Monaco.editor.IStandaloneCodeEditor): boolean {
	const model = editor.getModel();
	const position = editor.getPosition();
	if (!model || !position) {
		return false;
	}
	return listNextItemCompletion(model.getValue().split(/\r?\n/), position.lineNumber) !== null;
}

function dateBlockCompletionItems(
	monaco: typeof Monaco,
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
	scheduleTemplate: string,
): Monaco.languages.CompletionItem[] {
	const lines = model.getValue().split(/\r?\n/);
	const insertion = dateNextBlockCompletion(lines, position.lineNumber);
	const scheduleInsertion = scheduleSiblingCompletionText(lines, position.lineNumber, scheduleTemplate);
	if (!insertion && !scheduleInsertion) {
		return [];
	}

	const range = {
		startLineNumber: position.lineNumber,
		endLineNumber: position.lineNumber,
		startColumn: 1,
		endColumn: model.getLineMaxColumn(position.lineNumber),
	};
	const items: Monaco.languages.CompletionItem[] = [];
	if (insertion) {
		const label = insertion.text.trimStart().split(":")[0] || "next day";
		items.push({
			label,
			kind: monaco.languages.CompletionItemKind.Snippet,
			insertText: `${insertion.text}$0`,
			insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
			range,
			detail: "date block | optional",
			documentation: "Insert the next day block.",
			sortText: "0000",
		});
	}
	if (scheduleInsertion) {
		items.push({
			label: "schedules",
			kind: monaco.languages.CompletionItemKind.Snippet,
			insertText: `${scheduleInsertion}$0`,
			insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
			range,
			detail: "schedule block | optional",
			documentation: "Insert schedules block and configured schedule entries.",
			sortText: "0001",
		});
	}
	return items;
}

function scheduleSiblingCompletionText(
	lines: string[],
	lineNumber: number,
	scheduleTemplate: string,
): string | null {
	const indent = datesSiblingIndent(lines, lineNumber);
	if (indent === null || hasSiblingSchedules(lines, lineNumber, indent.length)) {
		return null;
	}

	const childIndent = `${indent}    `;
	const entries = sanitizeScheduleTemplate(scheduleTemplate)
		.split("\n")
		.filter((line) => line.trim() !== "")
		.map((line) => `${childIndent}${line}`)
		.join("\n");
	if (entries === "") {
		return `${indent}schedules:`;
	}
	return `${indent}schedules:\n${entries}`;
}

function hasSiblingSchedules(lines: string[], lineNumber: number, indentLength: number): boolean {
	for (let index = lineNumber - 1; index < lines.length; index++) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}
		const lineIndent = line.match(/^\s*/)?.[0].length ?? 0;
		if (lineIndent < indentLength) {
			return false;
		}
		if (lineIndent === indentLength && line.trim() === "schedules:") {
			return true;
		}
	}
	return false;
}

function listItemCompletionItem(
	monaco: typeof Monaco,
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
	sortOffset = 0,
): Monaco.languages.CompletionItem[] {
	const insertions = listNextItemCompletions(model.getValue().split(/\r?\n/), position.lineNumber);
	if (insertions.length === 0) {
		return [];
	}

	const range = {
		startLineNumber: position.lineNumber,
		endLineNumber: position.lineNumber,
		startColumn: 1,
		endColumn: model.getLineMaxColumn(position.lineNumber),
	};
	return insertions.map((insertion, index) => ({
		label: insertion.label,
		kind: monaco.languages.CompletionItemKind.Snippet,
		insertText: listItemMainInsertText(model, position, insertion),
		insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
		range,
		additionalTextEdits: listItemAdditionalTextEdits(model, position, insertion),
		detail: "list item | optional",
		documentation: "Insert a list item using the selected list shape.",
		sortText: `900${sortOffset + index + 1}`,
	}));
}

function listItemAdditionalTextEdits(
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
	insertion: { text: string; insertLineNumber?: number },
): Monaco.editor.ISingleEditOperation[] | undefined {
	if (!insertion.insertLineNumber || insertion.insertLineNumber === position.lineNumber) {
		return undefined;
	}
	return [{
		range: listItemEditRange(model, insertion.insertLineNumber),
		text: listItemEditText(model, insertion),
	}];
}

function listItemEditRange(model: Monaco.editor.ITextModel, insertLineNumber: number): Monaco.IRange {
	const lineNumber = Math.min(insertLineNumber, model.getLineCount() + 1);
	if (lineNumber > model.getLineCount()) {
		const lastLine = model.getLineCount();
		return {
			startLineNumber: lastLine,
			endLineNumber: lastLine,
			startColumn: model.getLineMaxColumn(lastLine),
			endColumn: model.getLineMaxColumn(lastLine),
		};
	}
	return {
		startLineNumber: lineNumber,
		endLineNumber: lineNumber,
		startColumn: 1,
		endColumn: 1,
	};
}

function listItemEditText(
	model: Monaco.editor.ITextModel,
	insertion: { text: string; insertLineNumber?: number },
): string {
	if (!insertion.insertLineNumber || insertion.insertLineNumber <= model.getLineCount()) {
		return `${insertion.text}\n`;
	}
	const lastLine = model.getLineContent(model.getLineCount());
	return `${lastLine.trim() === "" ? "" : "\n"}${insertion.text}`;
}

function listItemMainInsertText(
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
	insertion: { text: string; insertLineNumber?: number },
): string {
	if (!insertion.insertLineNumber || insertion.insertLineNumber === position.lineNumber) {
		return `${insertion.text}$0`;
	}
	return "";
}

async function toCompletionItem(
	monaco: typeof Monaco,
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
	candidate: CompletionCandidate,
	index = 0,
): Promise<Monaco.languages.CompletionItem> {
	const line = model.getLineContent(position.lineNumber);
	const isValue = line.lastIndexOf(":", position.column - 1) >= 0;
	const range = completionRange(model, position, line, candidate);

	if (isValue) {
		if (isReferenceValueLine(line)) {
			return {
				label: candidate.description
					? { label: candidate.name, description: candidate.description }
					: candidate.name,
				kind: monaco.languages.CompletionItemKind.Value,
				insertText: candidate.name,
				range,
				documentation: candidate.description,
				sortText: `100${index + 1}`,
			};
		}

		return {
			label: candidate.name,
			kind: monaco.languages.CompletionItemKind.Value,
			insertText: valueInsertText(line, position.column, candidate.name),
			insertTextRules: valueInsertTextRules(monaco, line, position.column, candidate.name),
			range,
			additionalTextEdits: await toolArgsAdditionalTextEdits(model, position, line, candidate.name),
			detail: completionDetail(candidate),
			documentation: candidate.description,
			sortText: `100${index + 1}`,
			command: shouldTriggerSuggestAfterCompletion(line, position.column, candidate.name)
				? { id: "editor.action.triggerSuggest", title: "Trigger Suggest" }
				: undefined,
		};
	}

	const isContainer = candidate.type === "struct" || candidate.type === "map";
	return {
		label: candidate.name,
		kind: monaco.languages.CompletionItemKind.Property,
		insertText: keyInsertText(candidate, isContainer),
		insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
		range,
		detail: completionDetail(candidate),
		documentation: candidate.description,
		sortText: `100${index + 1}`,
	};
}

function isReferenceValueLine(line: string): boolean {
	const trimmed = line.trim().replace(/^- /, "");
	return /^(day_ref|schedule_ref)\s*:/.test(trimmed);
}

function completionRange(
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
	line: string,
	candidate: CompletionCandidate,
): Monaco.IRange {
	const toolContext = toolValueContext(line, position.column);
	if (toolContext !== null) {
		const startColumn = candidate.name.endsWith(".")
			? toolContext.tokenStartColumn
			: toolContext.segmentStartColumn;
		return {
			startLineNumber: position.lineNumber,
			endLineNumber: position.lineNumber,
			startColumn,
			endColumn: position.column,
		};
	}

	const word = model.getWordUntilPosition(position);
	return {
		startLineNumber: position.lineNumber,
		endLineNumber: position.lineNumber,
		startColumn: word.startColumn,
		endColumn: word.endColumn,
	};
}

function valueInsertText(line: string, column: number, candidateName: string): string {
	const toolContext = toolValueContext(line, column);
	if (toolContext === null) {
		return candidateName;
	}

	if (candidateName.endsWith(".")) {
		if (toolContext.hasClosingQuoteAfterCursor) {
			return toolContext.hasOpeningQuote ? candidateName : `"${candidateName}`;
		}
		if (toolContext.hasOpeningQuote) {
			return `${candidateName}$0"`;
		}
		return `"${candidateName}$0"`;
	}

	if (toolContext.packageName === "") {
		if (toolContext.hasOpeningQuote && toolContext.hasClosingQuoteAfterCursor) {
			return candidateName;
		}
		if (toolContext.hasOpeningQuote) {
			return `${candidateName}"`;
		}
		if (toolContext.hasClosingQuoteAfterCursor) {
			return `"${candidateName}`;
		}
		return `"${candidateName}"`;
	}
	if (toolContext.hasOpeningQuote) {
		return toolContext.hasClosingQuoteAfterCursor ? candidateName : `${candidateName}"`;
	}
	return toolContext.hasClosingQuoteAfterCursor
		? `"${toolContext.packageName}.${candidateName}`
		: `"${toolContext.packageName}.${candidateName}"`;
}

function valueInsertTextRules(
	monaco: typeof Monaco,
	line: string,
	column: number,
	candidateName: string,
): Monaco.languages.CompletionItemInsertTextRule | undefined {
	const toolContext = toolValueContext(line, column);
	if (toolContext === null || !candidateName.endsWith(".") || toolContext.hasClosingQuoteAfterCursor) {
		return undefined;
	}
	return monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet;
}

function shouldTriggerSuggestAfterCompletion(line: string, column: number, candidateName: string): boolean {
	return candidateName.endsWith(".") && toolValueContext(line, column) !== null;
}

async function toolArgsAdditionalTextEdits(
	model: Monaco.editor.ITextModel,
	position: Monaco.Position,
	line: string,
	candidateName: string,
): Promise<Monaco.editor.ISingleEditOperation[] | undefined> {
	const toolContext = toolValueContext(line, position.column);
	if (toolContext === null || candidateName.endsWith(".") || toolContext.packageName === "") {
		return undefined;
	}

	const toolIndent = leadingWhitespaceLength(line);
	const argsIndent = " ".repeat(toolIndent);
	const argsChildIndent = `${argsIndent}  `;
	const completedToolName = `${toolContext.packageName}.${candidateName}`;
	const argsFieldCandidates = await argsCandidatesForTool(model, position.lineNumber, toolIndent, completedToolName);
	const argsLines = argsBlockLines(argsFieldCandidates, argsChildIndent);
	const argsBlockText = `${argsIndent}args:\n${argsLines.join("\n")}`;

	const existingArgsRange = findSiblingKeyBlockRange(model, position.lineNumber, toolIndent, "args");
	if (existingArgsRange !== null) {
		return [{
			range: existingArgsRange.range,
			text: `${argsBlockText}${existingArgsRange.needsTrailingNewline ? "\n" : ""}`,
		}];
	}

	return [{
		range: {
			startLineNumber: position.lineNumber,
			startColumn: model.getLineMaxColumn(position.lineNumber),
			endLineNumber: position.lineNumber,
			endColumn: model.getLineMaxColumn(position.lineNumber),
		},
		text: `\n${argsBlockText}`,
	}];
}

async function argsCandidatesForTool(
	model: Monaco.editor.ITextModel,
	toolLineNumber: number,
	toolIndent: number,
	toolName: string,
): Promise<CompletionCandidate[]> {
	const lines = model.getValue().split(/\r?\n/);
	const argsIndent = " ".repeat(toolIndent);
	const argsChildIndent = `${argsIndent}  `;
	lines[toolLineNumber - 1] = `${argsIndent}tool: "${toolName}"`;
	lines.splice(toolLineNumber, 0, `${argsIndent}args:`, argsChildIndent);

	const candidates = await completeYAML(
		lines.join("\n"),
		toolLineNumber + 2,
		argsChildIndent.length + 1,
	);
	return candidates.filter((candidate) => candidate.name !== "" || candidate.root === true);
}

function argsBlockLines(candidates: CompletionCandidate[], indent: string): string[] {
	if (candidates.length === 0) {
		return [indent];
	}
	if (candidates.length === 1 && candidates[0].root === true) {
		return rootArgsLines(candidates[0], indent);
	}
	return candidates.map((candidate) => argsFieldLine(candidate, indent));
}

function rootArgsLines(candidate: CompletionCandidate, indent: string): string[] {
	if ((candidate.type === "slice" || candidate.type === "array") && candidate.item) {
		return listItemTemplate(candidate.item, indent).split("\n");
	}
	if (candidate.type === "struct" && candidate.children) {
		return candidate.children.map((child) => fieldTemplate(child, indent));
	}
	return [indent];
}

function argsFieldLine(candidate: CompletionCandidate, indent: string): string {
	return fieldTemplate(candidate, indent);
}

function keyInsertText(candidate: CompletionCandidate, isContainer: boolean): string {
	if (candidate.type === "slice" || candidate.type === "array") {
		return `${fieldTemplate(candidate, "")}$0`;
	}
	if (isContainer) {
		return `${candidate.name}:\n  $0`;
	}
	return `${candidate.name}: ${candidateValue(candidate)}$0`;
}

function fieldTemplate(candidate: CompletionCandidate, indent: string): string {
	if (candidate.type === "slice" || candidate.type === "array") {
		return sliceFieldTemplate(candidate, indent);
	}
	if (candidate.type === "struct") {
		return structFieldTemplate(candidate, indent, "");
	}
	if (candidate.type === "map") {
		return `${indent}${candidate.name}:`;
	}
	return scalarFieldTemplate(candidate, indent, "");
}

function sliceFieldTemplate(candidate: CompletionCandidate, indent: string): string {
	const item = candidate.item;
	if (!item) {
		return `${indent}${candidate.name}:\n${indent}  - `;
	}
	return `${indent}${candidate.name}:\n${listItemTemplate(item, `${indent}  `)}`;
}

function listItemTemplate(item: CompletionCandidate, indent: string): string {
	if (item.type === "struct" && item.children && item.children.length > 0) {
		const [firstChild, ...restChildren] = item.children;
		const lines = [fieldTemplateWithPrefix(firstChild, indent, "- ")];
		for (const child of restChildren) {
			lines.push(fieldTemplate(child, `${indent}  `));
		}
		return lines.join("\n");
	}
	if (item.type === "slice" || item.type === "array") {
		return `${indent}-\n${sliceFieldTemplate({ ...item, name: item.name || "item" }, `${indent}  `)}`;
	}
	if (item.type === "map") {
		return `${indent}-`;
	}
	return `${indent}- ${candidateValue(item)}`;
}

function fieldTemplateWithPrefix(candidate: CompletionCandidate, indent: string, prefix: string): string {
	if (candidate.type === "slice" || candidate.type === "array") {
		const item = candidate.item;
		const itemLine = item ? listItemTemplate(item, `${indent}  `) : `${indent}  - `;
		return `${indent}${prefix}${candidate.name}:\n${itemLine}`;
	}
	if (candidate.type === "struct") {
		return structFieldTemplate(candidate, indent, prefix);
	}
	if (candidate.type === "map") {
		return `${indent}${prefix}${candidate.name}:`;
	}
	return scalarFieldTemplate(candidate, indent, prefix);
}

function structFieldTemplate(candidate: CompletionCandidate, indent: string, prefix: string): string {
	if (!candidate.children || candidate.children.length === 0) {
		return `${indent}${prefix}${candidate.name}:`;
	}
	return [
		`${indent}${prefix}${candidate.name}:`,
		...candidate.children.map((child) => fieldTemplate(child, `${indent}  `)),
	].join("\n");
}

function scalarFieldTemplate(candidate: CompletionCandidate, indent: string, prefix: string): string {
	return `${indent}${prefix}${candidate.name}: ${candidateValue(candidate)}`;
}

function candidateValue(candidate: CompletionCandidate): string {
	const defaultValue = candidate.default ?? candidate.enum?.[0] ?? "";
	return defaultValue;
}

type SiblingKeyBlockRange = {
	range: Monaco.IRange;
	needsTrailingNewline: boolean;
};

function findSiblingKeyBlockRange(
	model: Monaco.editor.ITextModel,
	lineNumber: number,
	indent: number,
	key: string,
): SiblingKeyBlockRange | null {
	const forwardLine = findSiblingKeyLine(model, lineNumber + 1, model.getLineCount(), 1, indent, key);
	if (forwardLine !== null) {
		return siblingKeyBlockRange(model, forwardLine, indent);
	}

	const backwardLine = findSiblingKeyLine(model, lineNumber - 1, 1, -1, indent, key);
	if (backwardLine !== null) {
		return siblingKeyBlockRange(model, backwardLine, indent);
	}

	return null;
}

function findSiblingKeyLine(
	model: Monaco.editor.ITextModel,
	startLine: number,
	endLine: number,
	step: 1 | -1,
	indent: number,
	key: string,
): number | null {
	for (
		let current = startLine;
		step > 0 ? current <= endLine : current >= endLine;
		current += step
	) {
		const line = model.getLineContent(current);
		if (line.trim() === "") {
			continue;
		}
		const lineIndent = leadingWhitespaceLength(line);
		if (lineIndent < indent) {
			break;
		}
		if (lineIndent === indent && yamlLineKey(line) === key) {
			return current;
		}
	}
	return null;
}

function siblingKeyBlockRange(
	model: Monaco.editor.ITextModel,
	lineNumber: number,
	indent: number,
): SiblingKeyBlockRange {
	for (let current = lineNumber + 1; current <= model.getLineCount(); current++) {
		const line = model.getLineContent(current);
		if (line.trim() === "") {
			continue;
		}
		const lineIndent = leadingWhitespaceLength(line);
		if (lineIndent <= indent) {
			return {
				range: {
					startLineNumber: lineNumber,
					startColumn: 1,
					endLineNumber: current,
					endColumn: 1,
				},
				needsTrailingNewline: true,
			};
		}
	}

	return {
		range: {
			startLineNumber: lineNumber,
			startColumn: 1,
			endLineNumber: model.getLineCount(),
			endColumn: model.getLineMaxColumn(model.getLineCount()),
		},
		needsTrailingNewline: false,
	};
}

function yamlLineKey(line: string): string | null {
	const trimmed = line.trim().replace(/^- /, "").trim();
	const colonIndex = trimmed.indexOf(":");
	if (colonIndex <= 0) {
		return null;
	}
	return trimmed.slice(0, colonIndex).trim();
}

function isToolValueLine(line: string, column: number): boolean {
	return toolValueContext(line, column) !== null;
}

type ToolValueContext = {
	hasOpeningQuote: boolean;
	hasClosingQuoteAfterCursor: boolean;
	packageName: string;
	tokenStartColumn: number;
	segmentStartColumn: number;
};

function toolValueContext(line: string, column: number): ToolValueContext | null {
	const prefix = line.slice(0, column - 1);
	const colonIndex = prefix.lastIndexOf(":");
	if (colonIndex < 0) {
		return null;
	}

	const key = prefix.slice(0, colonIndex).trim().replace(/^- /, "").trim();
	if (key !== "tool") {
		return null;
	}

	const rawValue = prefix.slice(colonIndex + 1);
	const leadingWhitespace = leadingWhitespaceLength(rawValue);
	const value = rawValue.slice(leadingWhitespace);
	const hasOpeningQuote = value.startsWith("\"");
	const hasClosingQuoteAfterCursor = line.slice(column - 1).trimStart().startsWith("\"");
	const token = hasOpeningQuote ? value.slice(1) : value;
	const valueStartColumn = colonIndex + 2 + leadingWhitespace;
	const tokenStartColumn = valueStartColumn + (hasOpeningQuote ? 1 : 0);
	const dotIndex = token.lastIndexOf(".");
	const packageName = dotIndex > 0 ? token.slice(0, dotIndex) : "";
	const segmentStartColumn = dotIndex >= 0 ? tokenStartColumn + dotIndex + 1 : tokenStartColumn;

	return {
		hasOpeningQuote,
		hasClosingQuoteAfterCursor,
		packageName,
		tokenStartColumn,
		segmentStartColumn,
	};
}

function leadingWhitespaceLength(value: string): number {
	const match = value.match(/^\s*/);
	return match?.[0].length ?? 0;
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
		if (insertedText !== "." && !insertedText.endsWith(".")) {
			return false;
		}
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
	if (isToolValueLine(linePrefix, linePrefix.length + 1)) {
		return /:\s*"?[A-Za-z0-9_.-]*$/.test(linePrefix);
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
