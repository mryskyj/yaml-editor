import Editor, { type OnMount } from "@monaco-editor/react";
import { useRef, useState } from "react";
import type * as Monaco from "monaco-editor";

const initialYaml = `server:
  host: localhost
  port: 8080
app:
  mode: dev
`;

export function EditorShell() {
  const editorRef = useRef<Monaco.editor.IStandaloneCodeEditor | null>(null);
  const [content, setContent] = useState(initialYaml);

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

  return (
    <main className="app-shell">
      <header className="toolbar">
        <button type="button" className="toolbar-button">
          New
        </button>
        <button type="button" className="toolbar-button">
          Open
        </button>
        <button type="button" className="toolbar-button">
          Save
        </button>
        <span className="toolbar-separator" />
        <button type="button" className="toolbar-button" onClick={() => editorRef.current?.trigger("toolbar", "undo", null)}>
          Undo
        </button>
        <button type="button" className="toolbar-button" onClick={() => editorRef.current?.trigger("toolbar", "redo", null)}>
          Redo
        </button>
      </header>
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
        <aside className="schema-pane">
          <h2>Schema</h2>
          <div className="schema-tree">
            <div className="schema-node">
              <span className="schema-key">server</span>
              <span className="schema-meta">struct required</span>
            </div>
            <div className="schema-node schema-child">
              <span className="schema-key">host</span>
              <span className="schema-meta">string required</span>
            </div>
            <div className="schema-node schema-child">
              <span className="schema-key">port</span>
              <span className="schema-meta">int required</span>
            </div>
            <div className="schema-node">
              <span className="schema-key">app</span>
              <span className="schema-meta">struct required</span>
            </div>
            <div className="schema-node schema-child">
              <span className="schema-key">mode</span>
              <span className="schema-meta">dev, stg, prod</span>
            </div>
          </div>
        </aside>
      </section>
      <footer className="error-list">
        <span className="error-status">0 errors</span>
      </footer>
    </main>
  );
}
