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
  return (
    <footer className="error-list">
      <div className="error-list-header">
        <span className="error-status">{diagnostics.length} errors</span>
      </div>
      <div className="error-list-body">
        {diagnostics.map((diagnostic, index) => (
          <button
            key={`${diagnostic.line}-${diagnostic.column}-${index}`}
            type="button"
            className="error-item"
            onClick={() => onSelect(diagnostic)}
          >
            <span className="error-location">
              {diagnostic.line}:{diagnostic.column}
            </span>
            <span className="error-message">{diagnostic.message}</span>
          </button>
        ))}
      </div>
    </footer>
  );
}
