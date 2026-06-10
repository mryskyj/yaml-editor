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
import { stepTemplateInsertion } from "./stepTemplates";
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
	const selectionKeyRef = useRef<Monaco.IDisposable | null>(null);
	const restoredTabIDRef = useRef<string | null>(null);
	const automaticEditRef = useRef(false);
	const validationRequestRef = useRef(0);
	const [tabState, setTabState] = useState<TabState>(() => createInitialTabState());
	const [pendingCloseTabID, setPendingCloseTabID] = useState<string | null>(null);
	const [isScheduleMenuOpen, setIsScheduleMenuOpen] = useState(false);
	const [scheduleTemplate, setScheduleTemplate] = useState(loadScheduleTemplate);
	const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null);
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
				<SchemaPane root={schema} content={content} cursor={cursor} />
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

function loadScheduleTemplate(): string {
	const stored = window.localStorage.getItem(scheduleTemplateStorageKey);
	if (!stored) {
		return defaultScheduleTemplate;
	}
	const sanitized = sanitizeScheduleTemplate(stored);
	return sanitized === "" ? defaultScheduleTemplate : sanitized;
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
	const range = completionRange(model, position, line, candidate);

	if (isValue) {
		return {
			label: candidate.name,
			kind: monaco.languages.CompletionItemKind.Value,
			insertText: valueInsertText(line, position.column, candidate.name),
			insertTextRules: valueInsertTextRules(monaco, line, position.column, candidate.name),
			range,
			detail: completionDetail(candidate),
			documentation: candidate.description,
			command: shouldTriggerSuggestAfterCompletion(line, position.column, candidate.name)
				? { id: "editor.action.triggerSuggest", title: "Trigger Suggest" }
				: undefined,
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
