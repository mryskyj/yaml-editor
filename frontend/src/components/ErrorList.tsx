import { useRef } from "react";

export type EditorDiagnostic = {
  message: string;
  line: number;
  column: number;
  severity: "error";
};

type ErrorListProps = {
  diagnostics: EditorDiagnostic[];
  onSelect: (diagnostic: EditorDiagnostic) => void;
};

export function ErrorList({ diagnostics, onSelect }: ErrorListProps) {
  const pointerStartRef = useRef<{ x: number; y: number } | null>(null);
  const didDragRef = useRef(false);

  const selectDiagnostic = (diagnostic: EditorDiagnostic) => {
    if (didDragRef.current) {
      didDragRef.current = false;
      return;
    }
    onSelect(diagnostic);
  };

  return (
    <footer className="error-list">
      <div className="error-list-header">
        <span className="error-status">{diagnostics.length} errors</span>
      </div>
      <div className="error-list-body">
        {diagnostics.map((diagnostic, index) => (
          <div
            key={`${diagnostic.line}-${diagnostic.column}-${index}`}
            role="button"
            tabIndex={0}
            className="error-item"
            onClick={() => selectDiagnostic(diagnostic)}
            onPointerDown={(event) => {
              pointerStartRef.current = { x: event.clientX, y: event.clientY };
              didDragRef.current = false;
            }}
            onPointerMove={(event) => {
              const start = pointerStartRef.current;
              if (!start) {
                return;
              }
              if (Math.abs(event.clientX - start.x) > 3 || Math.abs(event.clientY - start.y) > 3) {
                didDragRef.current = true;
              }
            }}
            onPointerUp={() => {
              pointerStartRef.current = null;
            }}
            onKeyDown={(event) => {
              if (event.key !== "Enter" && event.key !== " ") {
                return;
              }
              event.preventDefault();
              onSelect(diagnostic);
            }}
          >
            <span className="error-location">
              {diagnostic.line}:{diagnostic.column}
            </span>
            <span className="error-message">{diagnostic.message}</span>
          </div>
        ))}
      </div>
    </footer>
  );
}
