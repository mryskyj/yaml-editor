import { useState } from "react";

export type SchemaField = {
  name: string;
  type: string;
  required: boolean;
  description?: string;
  default?: string;
  enum?: string[];
  children?: SchemaField[];
  item?: SchemaField;
  mapValue?: SchemaField;
};

type SchemaPaneProps = {
  root: SchemaField;
  documentRoot?: SchemaField;
  content: string;
  cursor: {
    line: number;
    column: number;
  };
};

type TabID = "current" | "search" | "tree";

export function SchemaPane({ root, documentRoot, content, cursor }: SchemaPaneProps) {
  const [activeTab, setActiveTab] = useState<TabID>("current");
  const [query, setQuery] = useState("");
  const context = schemaContext(root, documentRoot, content, cursor.line);
  const searchResults = query.trim() ? searchSchema(root, query.trim()) : [];

  return (
    <aside className="schema-pane">
      <div className="schema-pane-header">
        <h2>Schema</h2>
        <div className="schema-tabs" role="tablist" aria-label="schema views">
          <button
            className={activeTab === "current" ? "schema-tab active" : "schema-tab"}
            type="button"
            onClick={() => setActiveTab("current")}
          >
            Current
          </button>
          <button
            className={activeTab === "search" ? "schema-tab active" : "schema-tab"}
            type="button"
            onClick={() => setActiveTab("search")}
          >
            Search
          </button>
          <button
            className={activeTab === "tree" ? "schema-tab active" : "schema-tab"}
            type="button"
            onClick={() => setActiveTab("tree")}
          >
            Tree
          </button>
        </div>
      </div>
      {activeTab === "current" ? <CurrentSchemaView context={context} /> : null}
      {activeTab === "search" ? (
        <SearchSchemaView query={query} results={searchResults} onQueryChange={setQuery} />
      ) : null}
      {activeTab === "tree" ? <SchemaTree root={root} /> : null}
    </aside>
  );
}

type SchemaContext = {
  rootName: string;
  path: string[];
  current: SchemaField;
  siblings: CandidateField[];
};

type CandidateField = {
  field: SchemaField;
  used: boolean;
};

type SearchResult = {
  field: SchemaField;
  path: string[];
};

function CurrentSchemaView({ context }: { context: SchemaContext }) {
  const breadcrumb = [context.rootName, ...context.path].filter(Boolean).join(" > ");
  return (
    <div className="schema-current">
      <div className="schema-breadcrumb" title={breadcrumb || context.current.name}>
        {breadcrumb || context.current.name}
      </div>
      <FieldSummary field={context.current} />
      <section className="schema-section">
        <h3>Available keys</h3>
        <div className="schema-candidate-list">
          {context.siblings.length === 0 ? (
            <p className="schema-empty">No child keys for this position.</p>
          ) : (
            context.siblings.map(({ field, used }) => (
              <div className={used ? "schema-candidate used" : "schema-candidate"} key={field.name}>
                <div className="schema-candidate-main">
                  <span className="schema-key">{field.name}</span>
                  <span className="schema-required">{field.required ? "required" : "optional"}</span>
                  {used ? <span className="schema-used">used</span> : null}
                </div>
                {field.description ? (
                  <div className="schema-candidate-description">{field.description}</div>
                ) : null}
              </div>
            ))
          )}
        </div>
      </section>
    </div>
  );
}

function SearchSchemaView({
  query,
  results,
  onQueryChange,
}: {
  query: string;
  results: SearchResult[];
  onQueryChange: (query: string) => void;
}) {
  return (
    <div className="schema-search">
      <input
        className="schema-search-input"
        type="search"
        value={query}
        placeholder="Search keys or descriptions"
        onChange={(event) => onQueryChange(event.target.value)}
      />
      <div className="schema-search-results">
        {query.trim() && results.length === 0 ? <p className="schema-empty">No matches.</p> : null}
        {results.map((result) => (
          <div className="schema-search-result" key={result.path.join(".")}>
            <div className="schema-search-path">{result.path.join(" > ")}</div>
            <FieldSummary field={result.field} compact />
          </div>
        ))}
      </div>
    </div>
  );
}

function SchemaTree({ root }: { root: SchemaField }) {
  return (
    <div className="schema-tree">
      {containerChildren(root).map((field) => (
        <SchemaNode key={field.name} field={field} depth={0} />
      ))}
    </div>
  );
}

function FieldSummary({ field, compact = false }: { field: SchemaField; compact?: boolean }) {
  return (
    <section className={compact ? "schema-summary compact" : "schema-summary"}>
      <div className="schema-primary">
        <span className="schema-key">{field.name || "root"}</span>
        {field.description ? <span className="schema-description">{field.description}</span> : null}
      </div>
      <div className="schema-details">
        <span>{field.type}</span>
        <span>{field.required ? "required" : "optional"}</span>
        {field.default ? <span>default {field.default}</span> : null}
        {field.enum && field.enum.length > 0 ? <span>{field.enum.join(", ")}</span> : null}
      </div>
    </section>
  );
}

function SchemaNode({ field, depth }: { field: SchemaField; depth: number }) {
  return (
    <>
      <div className="schema-node" style={{ paddingLeft: depth * 18 }}>
        <FieldSummary field={field} compact />
      </div>
      {containerChildren(field).map((child) => (
        <SchemaNode key={`${field.name}.${child.name}`} field={child} depth={depth + 1} />
      ))}
    </>
  );
}

function schemaContext(
  root: SchemaField,
  documentRoot: SchemaField | undefined,
  content: string,
  cursorLine: number,
): SchemaContext {
  const path = inferPath(content, cursorLine);
  const argsContext = argsSchemaContext(root, content, cursorLine);
  if (argsContext) {
    return argsContext;
  }
  if (isToolLineInActionBlock(content, cursorLine)) {
    return toolSchemaContext(root);
  }
  const documentContext = documentRoot ? documentSchemaContext(documentRoot, content, cursorLine) : null;
  if (documentContext) {
    return documentContext;
  }

  const matched = fieldAtPath(root, path);
  const current = matched ?? root;
  const currentPath = matched ? path : [];
  const parentPath = matched ? path.slice(0, -1) : path;
  const parent = fieldAtPath(root, parentPath) ?? root;
  const currentChildren = containerChildren(current);
  const candidateParent = currentChildren.length > 0 ? current : parent;
  const candidateParentPath = currentChildren.length > 0 ? currentPath : parentPath;
  const used = usedKeysAtPath(content, candidateParentPath);

  return {
    rootName: root.name,
    path: matched ? path : parentPath,
    current,
    siblings: containerChildren(candidateParent).map((field) => ({
      field,
      used: used.has(field.name),
    })),
  };
}

function toolSchemaContext(root: SchemaField): SchemaContext {
  return {
    rootName: root.name,
    path: [],
    current: root,
    siblings: containerChildren(root).map((field) => ({
      field,
      used: false,
    })),
  };
}

function argsSchemaContext(
  root: SchemaField,
  content: string,
  cursorLine: number,
): SchemaContext | null {
  const argsBlock = argsBlockContext(content, cursorLine);
  if (!argsBlock) {
    return null;
  }

  const toolName = siblingToolValue(content, argsBlock.lineIndex, argsBlock.indent);
  if (!toolName) {
    return null;
  }

  const toolSchema = containerChildren(root).find((field) => field.name === toolName);
  if (!toolSchema) {
    return null;
  }

  const argsRoot = toolSchema.type === "slice" || toolSchema.type === "array"
    ? toolSchema.item ?? toolSchema
    : toolSchema;
  const argsCurrent = argsCurrentContainer(argsRoot, content, cursorLine, argsBlock);
  const current = argsCurrent.field;

  return {
    rootName: toolName,
    path: argsCurrent.path,
    current,
    siblings: containerChildren(current).map((field) => ({
      field,
      used: false,
    })),
  };
}

function documentSchemaContext(
  documentRoot: SchemaField,
  content: string,
  cursorLine: number,
): SchemaContext | null {
  const path = inferPath(content, cursorLine);
  const matched = fieldAtPath(documentRoot, path);
  const currentPath = matched && hasContainerChildren(matched)
    ? path
    : nearestDocumentContainerPath(documentRoot, path);
  const field = matched && hasContainerChildren(matched)
    ? matched
    : documentContainerFromPathKeys(documentRoot, path) ?? fieldAtPath(documentRoot, currentPath);
  if (!field) {
    return null;
  }

  const container = documentContainerField(field);
  const containerPath = container === field ? currentPath : collectionContainerPath(currentPath);
  const displayField = displayStructField(
    container,
    currentPath[currentPath.length - 1] ?? field.name,
  );
  const used = usedKeysAtPath(content, containerPath);

  return {
    rootName: displayField.name,
    path: [],
    current: displayField,
    siblings: containerChildren(container).map((field) => ({
      field,
      used: used.has(field.name),
    })),
  };
}

function inferPath(content: string, cursorLine: number): string[] {
  const stack: Array<{ indent: number; key: string }> = [];
  const lines = content.split(/\r?\n/);
  const currentLineIndex = Math.max(Math.min(cursorLine - 1, lines.length - 1), 0);

  for (const line of lines.slice(0, currentLineIndex)) {
    const entry = parseKeyLine(line);
    if (!entry) {
      continue;
    }
    popToIndent(stack, entry.indent);
    stack.push(entry);
  }

  const currentLine = lines[currentLineIndex] ?? "";
  const currentEntry = parseKeyLine(currentLine);
  if (currentEntry) {
    popToIndent(stack, currentEntry.indent);
    stack.push(currentEntry);
    return stack.map((entry) => entry.key);
  }

  popToIndent(stack, currentIndent(currentLine));
  return stack.map((entry) => entry.key);
}

type ArgsBlockContext = {
  lineIndex: number;
  indent: number;
};

type ActionBlockContext = {
  lineIndex: number;
  indent: number;
};

function argsBlockContext(content: string, cursorLine: number): ArgsBlockContext | null {
  const lines = content.split(/\r?\n/);
  const currentLineIndex = Math.max(Math.min(cursorLine - 1, lines.length - 1), 0);
  for (let index = currentLineIndex; index >= 0; index--) {
    const line = lines[index] ?? "";
    if (line.trim() === "") {
      continue;
    }

    const entry = parseKeyLine(line);
    if (!entry) {
      continue;
    }
    if (entry.key === "args" && entry.indent < currentIndent(lines[currentLineIndex] ?? "")) {
      return { lineIndex: index, indent: entry.indent };
    }
    if (entry.indent <= currentIndent(lines[currentLineIndex] ?? "") && entry.key !== "args") {
      continue;
    }
  }
  return null;
}

function actionBlockContext(content: string, cursorLine: number): ActionBlockContext | null {
  const lines = content.split(/\r?\n/);
  const currentLineIndex = Math.max(Math.min(cursorLine - 1, lines.length - 1), 0);
  const cursorIndent = currentIndent(lines[currentLineIndex] ?? "");

  for (let index = currentLineIndex; index >= 0; index--) {
    const line = lines[index] ?? "";
    if (line.trim() === "") {
      continue;
    }

    const entry = parseKeyLine(line);
    if (!entry) {
      continue;
    }
    if (
      entry.key === "action" &&
      (index === currentLineIndex || entry.indent < cursorIndent)
    ) {
      return { lineIndex: index, indent: entry.indent };
    }
    if (entry.indent < cursorIndent && entry.key !== "action") {
      continue;
    }
  }
  return null;
}

function isToolLineInActionBlock(content: string, cursorLine: number): boolean {
  const actionBlock = actionBlockContext(content, cursorLine);
  if (!actionBlock) {
    return false;
  }
  const lines = content.split(/\r?\n/);
  const cursorIndex = Math.max(Math.min(cursorLine - 1, lines.length - 1), 0);
  const currentEntry = parseKeyLine(lines[cursorIndex] ?? "");
  return Boolean(
    currentEntry &&
      currentEntry.key === "tool" &&
      currentEntry.indent > actionBlock.indent,
  );
}

function siblingToolValue(content: string, argsLineIndex: number, argsIndent: number): string {
  const lines = content.split(/\r?\n/);
  for (let index = argsLineIndex - 1; index >= 0; index--) {
    const line = lines[index] ?? "";
    if (line.trim() === "") {
      continue;
    }

    const indent = currentIndent(line);
    if (indent < argsIndent) {
      return "";
    }
    if (indent !== argsIndent) {
      continue;
    }

    const parsed = parseKeyValueLine(line);
    if (parsed?.key === "tool") {
      return parsed.value;
    }
  }
  return "";
}

function argsCurrentContainer(
  argsRoot: SchemaField,
  content: string,
  cursorLine: number,
  argsBlock: ArgsBlockContext,
): { field: SchemaField; path: string[] } {
  const lines = content.split(/\r?\n/);
  const cursorIndex = Math.max(Math.min(cursorLine - 1, lines.length - 1), 0);

  for (let index = cursorIndex; index > argsBlock.lineIndex; index--) {
    const entry = parseKeyLine(lines[index] ?? "");
    if (!entry || entry.indent <= argsBlock.indent) {
      continue;
    }

    const child = containerChildren(argsRoot).find((field) => field.name === entry.key);
    if (!child) {
      continue;
    }

    const container = collectionValueField(child);
    if (containerChildren(container).length > 0) {
      return { field: container, path: [entry.key] };
    }
  }

  return { field: argsRoot, path: [] };
}

function popToIndent(stack: Array<{ indent: number; key: string }>, indent: number) {
  while (stack.length > 0 && stack[stack.length - 1].indent >= indent) {
    stack.pop();
  }
}

function currentIndent(line: string): number {
  const match = line.match(/^(\s*)/);
  return match ? match[1].length : 0;
}

function usedKeysAtPath(content: string, targetPath: string[]): Set<string> {
  const used = new Set<string>();
  const stack: Array<{ indent: number; key: string }> = [];

  for (const line of content.split(/\r?\n/)) {
    const entry = parseKeyLine(line);
    if (!entry) {
      continue;
    }
    while (stack.length > 0 && stack[stack.length - 1].indent >= entry.indent) {
      stack.pop();
    }
    if (samePath(stack.map((item) => item.key), targetPath)) {
      used.add(entry.key);
    }
    stack.push(entry);
  }

  return used;
}

function parseKeyLine(line: string): { indent: number; key: string } | null {
  const match = line.match(/^(\s*)(?:-\s*)?([A-Za-z0-9_-]+)\s*:/);
  if (!match) {
    return null;
  }
  return {
    indent: match[1].length,
    key: match[2],
  };
}

function parseKeyValueLine(line: string): { indent: number; key: string; value: string } | null {
  const match = line.match(/^(\s*)(?:-\s*)?([A-Za-z0-9_-]+)\s*:\s*(.*)$/);
  if (!match) {
    return null;
  }
  return {
    indent: match[1].length,
    key: match[2],
    value: match[3].trim().replace(/^['"]|['"]$/g, ""),
  };
}

function fieldAtPath(root: SchemaField, path: string[]): SchemaField | null {
  let current: SchemaField = root;
  for (const key of path) {
    const child = containerChildren(current).find((field) => field.name === key);
    if (!child) {
      return null;
    }
    current = child;
  }
  return current;
}

function containerChildren(field: SchemaField): SchemaField[] {
  if (field.children && field.children.length > 0) {
    return field.children;
  }
  if (field.item) {
    return containerChildren(field.item);
  }
  if (field.mapValue) {
    return containerChildren(field.mapValue);
  }
  return [];
}

function collectionValueField(field: SchemaField): SchemaField {
  if ((field.type === "slice" || field.type === "array") && field.item) {
    return field.item;
  }
  if (field.type === "map" && field.mapValue) {
    return field.mapValue;
  }
  return field;
}

function documentContainerField(field: SchemaField): SchemaField {
  if ((field.type === "slice" || field.type === "array" || field.type === "map") && containerChildren(field).length > 0) {
    return collectionValueField(field);
  }
  if (containerChildren(field).length > 0) {
    return field;
  }
  return field;
}

function collectionContainerPath(path: string[]): string[] {
  return path.length > 0 ? path : [];
}

function nearestDocumentContainerPath(root: SchemaField, path: string[]): string[] {
  for (let length = path.length; length >= 0; length--) {
    const candidatePath = path.slice(0, length);
    const candidate = fieldAtPath(root, candidatePath);
    if (candidate && containerChildren(documentContainerField(candidate)).length > 0) {
      return candidatePath;
    }
  }
  return [];
}

function documentContainerFromPathKeys(root: SchemaField, path: string[]): SchemaField | null {
  for (let index = path.length - 1; index >= 0; index--) {
    const candidate = findNamedContainer(root, path[index]);
    if (candidate) {
      return candidate;
    }
  }
  return null;
}

function findNamedContainer(root: SchemaField, name: string): SchemaField | null {
  const container = documentContainerField(root);
  if (root.name === name && hasContainerChildren(container)) {
    return container;
  }
  for (const child of containerChildren(root)) {
    const matched = findNamedContainer(child, name);
    if (matched) {
      return matched;
    }
  }
  return null;
}

function hasContainerChildren(field: SchemaField): boolean {
  return containerChildren(documentContainerField(field)).length > 0;
}

function displayStructField(field: SchemaField, fallbackName = ""): SchemaField {
  return {
    ...field,
    name: displayStructName(field.name || fallbackName),
  };
}

function displayStructName(name: string): string {
  const explicitNames: Record<string, string> = {
    action: "Action",
    common: "Common",
    date: "Date",
    dates: "Date",
    doc: "Doc",
    docs: "Doc",
    file: "File",
    scenario: "Scenario",
    step: "Step",
    steps: "Step",
  };
  const explicit = explicitNames[name];
  if (explicit) {
    return explicit;
  }
  const singular = name.endsWith("s") ? name.slice(0, -1) : name;
  return singular
    .split("_")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join("");
}

function searchSchema(root: SchemaField, query: string): SearchResult[] {
  const lowerQuery = query.toLowerCase();
  const results: SearchResult[] = [];

  walkSchema(root, [], (field, path) => {
    const haystack = [field.name, field.type, field.description, field.default, ...(field.enum ?? [])]
      .filter(Boolean)
      .join(" ")
      .toLowerCase();
    if (haystack.includes(lowerQuery)) {
      results.push({ field, path });
    }
  });

  return results.slice(0, 80);
}

function walkSchema(
  field: SchemaField,
  path: string[],
  visit: (field: SchemaField, path: string[]) => void,
) {
  const nextPath = field.name ? [...path, field.name] : path;
  visit(field, nextPath);
  for (const child of containerChildren(field)) {
    walkSchema(child, nextPath, visit);
  }
}

function samePath(left: string[], right: string[]): boolean {
  return left.length === right.length && left.every((item, index) => item === right[index]);
}
