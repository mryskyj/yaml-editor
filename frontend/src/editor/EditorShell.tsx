import Editor, { type OnMount } from "@monaco-editor/react";
import { useRef, useState } from "react";
import type * as Monaco from "monaco-editor";
import { ErrorList, type EditorDiagnostic } from "../components/ErrorList";
import { FileToolbar } from "../components/FileToolbar";
import { SchemaPane, type SchemaField } from "../components/SchemaPane";

const initialYaml = `server:
  host: localhost
  port: 8080
app:
  mode: dev
`;

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
	const [content, setContent] = useState(initialYaml);
	const [currentFileName, setCurrentFileName] = useState("config.yaml");
	const [recentFiles, setRecentFiles] = useState<string[]>(["config.yaml"]);
	const [diagnostics] = useState<EditorDiagnostic[]>([]);

  const handleMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;
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
  };

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
		setContent("");
		setCurrentFileName("untitled.yaml");
	};

	const handleOpen = (fileName: string, nextContent: string) => {
		setContent(nextContent);
		setCurrentFileName(fileName);
		setRecentFiles((files) => [fileName, ...files.filter((file) => file !== fileName)].slice(0, 5));
	};

	const handleSave = () => {
		const blob = new Blob([content], { type: "text/yaml;charset=utf-8" });
		const url = URL.createObjectURL(blob);
		const link = document.createElement("a");
		link.href = url;
		link.download = currentFileName || "config.yaml";
		link.click();
		URL.revokeObjectURL(url);
	};

	return (
		<main className="app-shell">
			<FileToolbar
				currentFileName={currentFileName}
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
            onChange={(value) => setContent(value ?? "")}
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
              tabSize: 2,
              wordWrap: "on",
            }}
          />
        </div>
				<SchemaPane root={sampleSchema} />
			</section>
			<ErrorList diagnostics={diagnostics} onSelect={handleSelectDiagnostic} />
		</main>
	);
}
