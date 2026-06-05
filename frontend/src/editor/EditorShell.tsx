import Editor, { type OnMount } from "@monaco-editor/react";
import { useCallback, useEffect, useRef, useState } from "react";
import type * as Monaco from "monaco-editor";
import {
	chooseSavePath,
	completeYAML,
	loadSchema,
	saveYAML,
	type CompletionCandidate,
	validateYAML,
} from "../app/api";
import { ErrorList, type EditorDiagnostic } from "../components/ErrorList";
import { FileToolbar } from "../components/FileToolbar";
import { SchemaPane, type SchemaField } from "../components/SchemaPane";
import {
	activeTab,
	addUntitledTab,
	createInitialTabState,
	markActiveTabSaved,
	openDocumentTab,
	updateActiveContent,
	updateTabCursor,
	updateTabDiagnostics,
	type TabState,
} from "./tabs";

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
	const validationRequestRef = useRef(0);
	const [tabState, setTabState] = useState<TabState>(() => createInitialTabState());
	const [recentFiles, setRecentFiles] = useState<string[]>(["config.yaml"]);
	const [schema, setSchema] = useState<SchemaField>(sampleSchema);
	const currentTab = activeTab(tabState);
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
						.map((candidate) => toCompletionItem(monaco, model, position, candidate)),
				};
			},
		});
		contentChangeRef.current = editor.onDidChangeModelContent((event) => {
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
		const timerID = window.setTimeout(() => {
			void runValidation(currentTab.id, content);
		}, 200);
		return () => window.clearTimeout(timerID);
	}, [content, currentTab.id, runValidation]);

	useEffect(() => {
		applyMarkers(diagnostics);
	}, [diagnostics, applyMarkers]);

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

	const handleNew = () => {
		setTabState(addUntitledTab);
	};

	const handleOpen = (fileName: string, nextContent: string) => {
		setTabState((state) => openDocumentTab(state, {
			name: fileName,
			content: nextContent,
		}));
		setRecentFiles((files) => [fileName, ...files.filter((file) => file !== fileName)].slice(0, 5));
	};

	const handleSave = async () => {
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
	};

	return (
		<main className="app-shell">
			<FileToolbar
				currentFileName={currentTab.name}
				recentFiles={recentFiles}
				onNew={handleNew}
				onOpen={handleOpen}
				onSave={handleSave}
				onUndo={() => editorRef.current?.trigger("toolbar", "undo", null)}
				onRedo={() => editorRef.current?.trigger("toolbar", "redo", null)}
			/>
      <section className="workspace">
        <div className="editor-region">
          <Editor
            height="100%"
            defaultLanguage="yaml"
            value={content}
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
              scrollBeyondLastLine: false,
              quickSuggestions: { other: true, comments: false, strings: false },
              suggestOnTriggerCharacters: true,
              tabSize: 2,
              wordWrap: "on",
            }}
          />
        </div>
				<SchemaPane root={schema} content={content} cursor={cursor} />
			</section>
			<ErrorList diagnostics={diagnostics} onSelect={handleSelectDiagnostic} />
		</main>
	);
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
