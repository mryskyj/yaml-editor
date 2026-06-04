import { type ChangeEvent, useRef } from "react";

type FileToolbarProps = {
  currentFileName: string;
  recentFiles: string[];
  onNew: () => void;
  onOpen: (fileName: string, content: string) => void;
  onSave: () => void;
  onUndo: () => void;
  onRedo: () => void;
};

export function FileToolbar({ currentFileName, recentFiles, onNew, onOpen, onSave, onUndo, onRedo }: FileToolbarProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);

  const handleOpen = () => {
    inputRef.current?.click();
  };

  const handleFileChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }

    const content = await file.text();
    onOpen(file.name, content);
    event.target.value = "";
  };

  return (
    <header className="toolbar">
      <button type="button" className="toolbar-button" onClick={onNew}>
        New
      </button>
      <button type="button" className="toolbar-button" onClick={handleOpen}>
        Open
      </button>
      <button type="button" className="toolbar-button" onClick={onSave}>
        Save
      </button>
      <select className="recent-select" value="" onChange={(event) => event.target.value && handleOpen()}>
        <option value="">Recent</option>
        {recentFiles.map((file) => (
          <option key={file} value={file}>
            {file}
          </option>
        ))}
      </select>
      <span className="current-file">{currentFileName}</span>
      <span className="toolbar-separator" />
      <button type="button" className="toolbar-button" onClick={onUndo}>
        Undo
      </button>
      <button type="button" className="toolbar-button" onClick={onRedo}>
        Redo
      </button>
      <input ref={inputRef} className="file-input" type="file" accept=".yaml,.yml,text/yaml,text/plain" onChange={handleFileChange} />
    </header>
  );
}
