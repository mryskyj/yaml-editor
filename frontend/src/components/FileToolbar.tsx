import { type ChangeEvent, useEffect, useRef } from "react";

type FileToolbarProps = {
  currentFileName: string;
  recentFiles: string[];
  onNew: () => void;
  onOpen: () => Promise<boolean>;
  onOpenLocalFile: (fileName: string, content: string) => void;
  onOpenRecent: (path: string) => void;
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
  onOpenLocalFile,
  onOpenRecent,
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
  const modifier = shortcutModifier();

  const handleOpen = () => {
    void onOpen().then((handled) => {
      if (!handled) {
        inputRef.current?.click();
      }
    });
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
    onOpenLocalFile(file.name, content);
    event.target.value = "";
  };

  return (
    <header className="toolbar">
      <ToolbarButton label="New" shortcut={`${modifier} + N`} onClick={onNew} />
      <ToolbarButton label="Open" shortcut={`${modifier} + O`} onClick={handleOpen} />
      <ToolbarButton label="Save" shortcut={`${modifier} + S`} onClick={onSave} variant="primary" />
      <ToolbarButton label="Schedules" onClick={onSchedules} />
      <select className="recent-select" value="" onChange={(event) => event.target.value && onOpenRecent(event.target.value)}>
        <option value="">Recent</option>
        {recentFiles.map((file) => (
          <option key={file} value={file}>
            {file}
          </option>
        ))}
      </select>
      <span className="current-file">{currentFileName}</span>
      <span className="toolbar-separator" />
      <ToolbarButton label="Copy" shortcut={`${modifier} + C`} onClick={onCopy} />
      <ToolbarButton label="Cut" shortcut={`${modifier} + X`} onClick={onCut} />
      <ToolbarButton label="Paste" shortcut={`${modifier} + V`} onClick={onPaste} />
      <span className="toolbar-separator" />
      <ToolbarButton label="Undo" shortcut={`${modifier} + Z`} onClick={onUndo} />
      <ToolbarButton label="Redo" shortcut={`${modifier} + Shift + Z`} onClick={onRedo} />
      <input ref={inputRef} className="file-input" type="file" accept=".yaml,.yml,text/yaml,text/plain" onChange={handleFileChange} />
    </header>
  );
}

function ToolbarButton({
  label,
  onClick,
  shortcut,
  variant = "default",
}: {
  label: string;
  onClick: () => void;
  shortcut?: string;
  variant?: "default" | "primary";
}) {
  const tooltip = shortcut ? `${label} (${shortcut})` : label;
  const className = variant === "primary" ? "toolbar-button primary" : "toolbar-button";
  return (
    <button
      type="button"
      className={className}
      data-tooltip={tooltip}
      aria-label={tooltip}
      onClick={onClick}
    >
      {label}
    </button>
  );
}

function shortcutModifier(): string {
  if (typeof navigator !== "undefined" && /Mac|iPhone|iPad|iPod/.test(navigator.platform)) {
    return "Cmd";
  }
  return "Ctrl";
}
