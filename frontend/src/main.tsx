import React from "react";
import { createRoot } from "react-dom/client";
import { EditorShell } from "./editor/EditorShell";
import "./styles.css";

createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <EditorShell />
  </React.StrictMode>,
);
