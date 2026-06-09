import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { cpSync } from "node:fs";
import { resolve } from "node:path";

function copyMonacoAssets() {
  return {
    name: "copy-monaco-assets",
    closeBundle() {
      cpSync(
        resolve(__dirname, "node_modules/monaco-editor/min/vs"),
        resolve(__dirname, "dist/vs"),
        { recursive: true },
      );
    },
  };
}

export default defineConfig({
  plugins: [react(), copyMonacoAssets()],
  server: {
    host: "127.0.0.1",
    port: 5173,
  },
});
