import React from "react";
import { createRoot } from "react-dom/client";
import { loader } from "@monaco-editor/react";
import { EditorShell } from "./editor/EditorShell";
import "./styles.css";

if (import.meta.env.PROD) {
  loader.config({ paths: { vs: "/vs" } });
}

createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <EditorShell />
  </React.StrictMode>,
);
