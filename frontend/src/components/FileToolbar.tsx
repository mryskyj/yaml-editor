import { type ChangeEvent, useEffect, useRef } from "react";

type FileToolbarProps = {
  currentFileName: string;
  recentFiles: string[];
  onNew: () => void;
  onOpen: (fileName: string, content: string) => void;
  onSave: () => void;
  onSchedules: () => void;
  onCopy: () => void;
  onCut: () => void;
  onPaste: () => void;
  onUndo: () => void;
  onRedo: () => void;
  openRequestID: number;
};

export function FileToolbar({
  currentFileName,
  recentFiles,
  onNew,
  onOpen,
  onSave,
  onSchedules,
  onCopy,
  onCut,
  onPaste,
  onUndo,
  onRedo,
  openRequestID,
}: FileToolbarProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const lastOpenRequestIDRef = useRef(openRequestID);

  const handleOpen = () => {
    inputRef.current?.click();
  };

  useEffect(() => {
    if (openRequestID === lastOpenRequestIDRef.current) {
      return;
    }
    lastOpenRequestIDRef.current = openRequestID;
    handleOpen();
  }, [openRequestID]);

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
      <button type="button" className="toolbar-button" onClick={onSchedules}>
        Schedules
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
      <button type="button" className="toolbar-button" onClick={onCopy}>
        Copy
      </button>
      <button type="button" className="toolbar-button" onClick={onCut}>
        Cut
      </button>
      <button type="button" className="toolbar-button" onClick={onPaste}>
        Paste
      </button>
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
